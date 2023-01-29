package node

import (
	"crypto"
	"fmt"
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/abci/apps"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/consensus"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/proxy"
	"github.com/232425wxy/meta--/state"
	"github.com/232425wxy/meta--/txspool"
	"github.com/232425wxy/meta--/types"
)

type DBProvider func(name string, cfg *config.Config) (database.DB, error)

func DefaultDBProvider(name string, cfg *config.Config) (database.DB, error) {
	db, err := database.NewDB(name, cfg.BasicConfig.DBPath(), database.BackendType(cfg.BasicConfig.DBBackend))
	if err != nil {
		return nil, err
	}
	return db, nil
}

type GenesisProvider func(cfg *config.Config) (*types.Genesis, error)

func DefaultGenesisProvider(cfg *config.Config) (*types.Genesis, error) {
	genesis, err := types.GenesisReadFromFile(cfg.BasicConfig.GenesisFilePath())
	if err != nil {
		return nil, err
	}
	return genesis, nil
}

type ApplicationProvider func(cfg *config.Config) abci.Application

func DefaultApplicationProvider(cfg *config.Config) abci.Application {
	var app abci.Application
	switch cfg.BasicConfig.App {
	case "kvstore":
		app = apps.NewKVStoreApp("kvstore", cfg.BasicConfig.DBPath(), database.BackendType(cfg.BasicConfig.DBBackend))
	default:
		panic(fmt.Sprintf("unknown app type: %s", cfg.BasicConfig.App))
	}
	return app
}

type TxspoolProvider func(cfg *config.Config, proxyAppConn *proxy.AppConns, state *state.State, logger log.Logger) (*txspool.TxsPool, *txspool.Reactor)

func DefaultTxsPoolProvider(cfg *config.Config, proxyAppConn *proxy.AppConns, state *state.State, logger log.Logger) (*txspool.TxsPool, *txspool.Reactor) {
	pool := txspool.NewTxsPool(cfg.TxsPoolConfig, proxyAppConn.TxsPool(), state.LastBlockHeight)
	reactor := txspool.NewReactor(cfg.TxsPoolConfig, pool)
	reactor.SetLogger(logger.New("module", "TxsPool"))
	return pool, reactor
}

type ConsensusProvider func(cfg *config.Config, stat *state.State, exec *state.BlockExecutor, blockStore *state.StoreBlock, txsPool *txspool.TxsPool, privateKey *crypto.PrivateKey, eventBus *events.EventBus, logger log.Logger) (*consensus.Core, *consensus.Reactor)

func DefaultConsensusProvider(cfg *config.Config, stat *state.State, exec *state.BlockExecutor, blockStore *state.StoreBlock, txsPool *txspool.TxsPool, privateKey *crypto.PrivateKey, eventBus *events.EventBus, logger log.Logger) (*consensus.Core, *consensus.Reactor) {
	core := consensus.NewCore(cfg.ConsensusConfig, stat)
}

type Provider struct {
	DBProvider          DBProvider
	GenesisProvider     GenesisProvider
	ApplicationProvider ApplicationProvider
	TxspoolProvider     TxspoolProvider
}

type Node struct {
	service.BaseService
	cfg        *config.Config
	genesis    *types.Genesis
	transport  *p2p.Transport
	sw         *p2p.Switch
	addrBook   *p2p.AddrBook
	nodeInfo   *p2p.NodeInfo
	nodeKey    *p2p.NodeKey
	eventBUs   *events.EventBus
	stateStore *state.StoreState
	blockStore *state.StoreBlock
	txsPool    *txspool.TxsPool

	txsPoolReactor   *txspool.Reactor
	consensusReactor *consensus.Reactor
}

func NewNode(cfg *config.Config, logger log.Logger, provider Provider) (*Node, error) {
	nodeKey, err := p2p.LoadNodeKey(cfg.BasicConfig.KeyFilePath())
	if err != nil {
		return nil, err
	}

	eventBus, err := events.CreateAndStartEventBus(logger)
	if err != nil {
		return nil, err
	}

	blockStoreDB, err := provider.DBProvider("blocks", cfg)
	if err != nil {
		return nil, err
	}
	blockStore := state.NewStoreBlock(blockStoreDB)

	stateDB, err := provider.DBProvider("state", cfg)
	if err != nil {
		return nil, err
	}
	stateStore := state.NewStoreState(stateDB)

	genesis, err := provider.GenesisProvider(cfg)
	if err != nil {
		return nil, err
	}

	stat := stateStore.LoadFromDBOrGenesis(genesis)

	nodeInfo := &p2p.NodeInfo{
		NodeID:     nodeKey.GetID(),
		ListenAddr: cfg.P2PConfig.ListenAddress,
		Channels:   []byte{p2p.LeaderProposeChannel, p2p.ReplicaVoteChannel, p2p.ReplicaStateChannel, p2p.TxsChannel},
		RPCAddress: "",
		TxIndex:    "on",
	}

	application := provider.ApplicationProvider(cfg)
	proxyAppConns := proxy.NewAppConns(application, logger)
	if err = proxyAppConns.Start(); err != nil {
		return nil, err
	}

	txsPool, txsPoolReactor := provider.TxspoolProvider(cfg, proxyAppConns, stat, logger)

	blockExec := state.NewBlockExecutor(stateStore, proxyAppConns.Consensus(), txsPool, logger.New("module", "state"))

	return &Node{
		nodeInfo:   nodeInfo,
		eventBUs:   eventBus,
		blockStore: blockStore,
		stateStore: stateStore,
	}, nil
}
