package node

import (
	"fmt"
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/abci/apps"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/consensus"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/proxy"
	"github.com/232425wxy/meta--/state"
	"github.com/232425wxy/meta--/txspool"
	"github.com/232425wxy/meta--/types"
	"time"
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

type ConsensusProvider func(cfg *config.Config, stat *state.State, exec *state.BlockExecutor, blockStore *state.StoreBlock, txsPool *txspool.TxsPool, privateKey *bls12.PrivateKey, bls *bls12.CryptoBLS12, logger log.Logger) (*consensus.Core, *consensus.Reactor)

func DefaultConsensusProvider(cfg *config.Config, stat *state.State, exec *state.BlockExecutor, blockStore *state.StoreBlock, txsPool *txspool.TxsPool, privateKey *bls12.PrivateKey, bls *bls12.CryptoBLS12, logger log.Logger) (*consensus.Core, *consensus.Reactor) {
	core := consensus.NewCore(cfg.ConsensusConfig, privateKey, stat, exec, blockStore, txsPool, bls)
	core.SetLogger(logger.New("module", "Consensus"))
	reactor := consensus.NewReactor(core)
	return core, reactor
}

type P2PProvider func(cfg *config.Config, nodeInfo *p2p.NodeInfo, nodeKey *p2p.NodeKey, txsPoolReactor *txspool.Reactor, consensusReactor *consensus.Reactor, logger log.Logger) (*p2p.Transport, *p2p.Switch)

func DefaultP2PProvider(cfg *config.Config, nodeInfo *p2p.NodeInfo, nodeKey *p2p.NodeKey, txsPoolReactor *txspool.Reactor, consensusReactor *consensus.Reactor, logger log.Logger) (*p2p.Transport, *p2p.Switch) {
	addr, err := p2p.NewNetAddressString(p2p.IDAddressString(nodeKey.GetID(), cfg.P2PConfig.ListenAddress))
	if err != nil {
		panic(err)
	}
	transport := p2p.NewTransport(addr, nodeInfo, nodeKey, cfg.P2PConfig)
	sw := p2p.NewSwitch(transport, p2p.P2PMetrics())
	sw.SetLogger(logger.New("module", "Switch"))
	sw.AddReactor("TXSPOOL", txsPoolReactor)
	sw.AddReactor("CONSENSUS", consensusReactor)
	return transport, sw
}

type Provider struct {
	DBProvider          DBProvider
	GenesisProvider     GenesisProvider
	ApplicationProvider ApplicationProvider
	TxspoolProvider     TxspoolProvider
	ConsensusProvider   ConsensusProvider
	P2PProvider         P2PProvider
}

func DefaultProvider() Provider {
	return Provider{
		DBProvider:          DefaultDBProvider,
		GenesisProvider:     DefaultGenesisProvider,
		ApplicationProvider: DefaultApplicationProvider,
		TxspoolProvider:     DefaultTxsPoolProvider,
		ConsensusProvider:   DefaultConsensusProvider,
		P2PProvider:         DefaultP2PProvider,
	}
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
		NodeID:      nodeKey.GetID(),
		ListenAddr:  cfg.P2PConfig.ListenAddress,
		Channels:    []byte{p2p.LeaderProposeChannel, p2p.ReplicaVoteChannel, p2p.ReplicaStateChannel, p2p.TxsChannel},
		RPCAddress:  "",
		TxIndex:     "on",
		CryptoBLS12: bls12.NewCryptoBLS12(),
	}
	nodeInfo.CryptoBLS12.Init(nodeKey.PrivateKey)

	application := provider.ApplicationProvider(cfg)
	proxyAppConns := proxy.NewAppConns(application, logger)
	if err = proxyAppConns.Start(); err != nil {
		return nil, err
	}

	txsPool, txsPoolReactor := provider.TxspoolProvider(cfg, proxyAppConns, stat, logger)

	blockExec := state.NewBlockExecutor(cfg, stateStore, proxyAppConns.Consensus(), txsPool, logger.New("module", "state"))

	consensusStat, consensusReactor := provider.ConsensusProvider(cfg, stat, blockExec, blockStore, txsPool, nodeKey.PrivateKey, nodeInfo.CryptoBLS12, logger.New("module", "Consensus"))
	consensusStat.SetEventBus(eventBus)

	transport, sw := provider.P2PProvider(cfg, nodeInfo, nodeKey, txsPoolReactor, consensusReactor, logger)

	addrBook := p2p.NewAddrBook(cfg.P2PConfig.AddrBookPath())
	if cfg.P2PConfig.ListenAddress != "" {
		addr, err := p2p.NewNetAddressString(p2p.IDAddressString(nodeKey.GetID(), cfg.P2PConfig.ListenAddress))
		if err != nil {
			return nil, err
		}
		addrBook.AddOurAddress(addr)
	}
	sw.SetAddrBook(addrBook)

	n := &Node{
		BaseService:      *service.NewBaseService(logger.New("node", nodeKey.GetID()), "Node"),
		cfg:              cfg,
		genesis:          genesis,
		transport:        transport,
		sw:               sw,
		addrBook:         addrBook,
		nodeInfo:         nodeInfo,
		nodeKey:          nodeKey,
		eventBUs:         eventBus,
		stateStore:       stateStore,
		blockStore:       blockStore,
		txsPool:          txsPool,
		txsPoolReactor:   txsPoolReactor,
		consensusReactor: consensusReactor,
	}
	return n, nil
}

func (n *Node) Start() error {
	if n.genesis.GenesisTime.After(time.Now()) {
		n.Logger.Info("genesis time is in the future, sleeping until then...", "sleep_duration", n.genesis.GenesisTime.Sub(time.Now()).Seconds())
		time.Sleep(n.genesis.GenesisTime.Sub(time.Now()))
	}

	// 开始监听网络中其他peer的连接请求

	if err := n.transport.Listen(); err != nil {
		return err
	}

	if err := n.sw.Start(); err != nil {
		return err
	}

	n.sw.DialPeerAsync(n.cfg.P2PConfig.NeighboursSlice())
	return n.BaseService.Start()
}
