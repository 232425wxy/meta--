package node

import (
	"fmt"
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/abci/apps"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/consensus"
	state2 "github.com/232425wxy/meta--/consensus/state"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/proxy"
	"github.com/232425wxy/meta--/stch"
	"github.com/232425wxy/meta--/store"
	"github.com/232425wxy/meta--/syncer"
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

type TxspoolProvider func(cfg *config.Config, proxyAppConn *proxy.AppConns, state *state2.State, logger log.Logger) (*txspool.TxsPool, *txspool.Reactor)

func DefaultTxsPoolProvider(cfg *config.Config, proxyAppConn *proxy.AppConns, state *state2.State, logger log.Logger) (*txspool.TxsPool, *txspool.Reactor) {
	pool := txspool.NewTxsPool(cfg.TxsPoolConfig, proxyAppConn.TxsPool(), state.LastBlockHeight)
	reactor := txspool.NewReactor(cfg.TxsPoolConfig, pool)
	reactor.SetLogger(logger.New("module", "TxsPool"))
	return pool, reactor
}

type ConsensusProvider func(cfg *config.Config, stat *state2.State, exec *state2.BlockExecutor, txsPool *txspool.TxsPool, privateKey *bls12.PrivateKey, bls *bls12.CryptoBLS12, logger log.Logger) (*consensus.Core, *consensus.Reactor)

func DefaultConsensusProvider(cfg *config.Config, stat *state2.State, exec *state2.BlockExecutor, txsPool *txspool.TxsPool, privateKey *bls12.PrivateKey, bls *bls12.CryptoBLS12, logger log.Logger) (*consensus.Core, *consensus.Reactor) {
	core := consensus.NewCore(cfg.ConsensusConfig, privateKey, stat, exec, txsPool, bls)
	core.SetLogger(logger.New("module", "Consensus"))
	reactor := consensus.NewReactor(core)
	reactor.SetLogger(logger.New("module", "Consensus_Reactor"))
	return core, reactor
}

type P2PProvider func(cfg *config.Config, nodeInfo *p2p.NodeInfo, nodeKey *p2p.NodeKey, txsPoolReactor *txspool.Reactor, consensusReactor *consensus.Reactor, syncerReactor *syncer.Reactor, stchReactor *stch.Reactor, logger log.Logger) (*p2p.Transport, *p2p.Switch)

func DefaultP2PProvider(cfg *config.Config, nodeInfo *p2p.NodeInfo, nodeKey *p2p.NodeKey, txsPoolReactor *txspool.Reactor, consensusReactor *consensus.Reactor, syncerReactor *syncer.Reactor, stchReactor *stch.Reactor, logger log.Logger) (*p2p.Transport, *p2p.Switch) {
	addr, err := p2p.NewNetAddressString(p2p.IDAddressString(nodeKey.GetID(), cfg.P2PConfig.ListenAddress))
	if err != nil {
		panic(err)
	}
	transport := p2p.NewTransport(addr, nodeInfo, nodeKey, cfg.P2PConfig)
	sw := p2p.NewSwitch(transport, p2p.P2PMetrics())
	sw.SetLogger(logger.New("module", "Switch"))
	sw.AddReactor("TXSPOOL", txsPoolReactor)
	sw.AddReactor("CONSENSUS", consensusReactor)
	sw.AddReactor("SYNCER", syncerReactor)
	sw.AddReactor("STCH", stchReactor)
	return transport, sw
}

type SyncerProvider func(stat *state2.State, blockExec *state2.BlockExecutor, blockStore *store.BlockStore, logger log.Logger) *syncer.Reactor

func DefaultSyncerProvider(stat *state2.State, blockExec *state2.BlockExecutor, blockStore *store.BlockStore, logger log.Logger) *syncer.Reactor {
	reactor := syncer.NewReactor(stat, blockExec, blockStore, logger.New("module", "Syncer"))
	return reactor
}

type STCHProvider func(id crypto.ID, participantsNum int, logger log.Logger) *stch.Reactor

func DefaultSTCHProvider(id crypto.ID, participantsNum int, logger log.Logger) *stch.Reactor {
	ch := stch.NewChameleon(id, participantsNum)
	r := stch.NewReactor(ch)
	r.SetLogger(logger.New("module", "STCH"))
	return r
}

type Provider struct {
	DBProvider          DBProvider
	GenesisProvider     GenesisProvider
	ApplicationProvider ApplicationProvider
	TxspoolProvider     TxspoolProvider
	ConsensusProvider   ConsensusProvider
	P2PProvider         P2PProvider
	SyncerProvider      SyncerProvider
	STCHProvider        STCHProvider
}

func DefaultProvider() Provider {
	return Provider{
		DBProvider:          DefaultDBProvider,
		GenesisProvider:     DefaultGenesisProvider,
		ApplicationProvider: DefaultApplicationProvider,
		TxspoolProvider:     DefaultTxsPoolProvider,
		ConsensusProvider:   DefaultConsensusProvider,
		P2PProvider:         DefaultP2PProvider,
		SyncerProvider:      DefaultSyncerProvider,
		STCHProvider:        DefaultSTCHProvider,
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
	stateStore *state2.StoreState
	blockStore *store.BlockStore
	txsPool    *txspool.TxsPool

	txsPoolReactor   *txspool.Reactor
	consensusReactor *consensus.Reactor
}

func NewNode(cfg *config.Config, logger log.Logger, provider Provider) (*Node, error) {
	nodeKey, err := p2p.LoadNodeKey(cfg.BasicConfig.KeyFilePath())
	if err != nil {
		return nil, err
	}

	kp := stch.LoadInitConfig(cfg.BasicConfig.ChameleonKeyFilePath())

	eventBus, err := events.CreateAndStartEventBus(logger)
	if err != nil {
		return nil, err
	}

	blockStoreDB, err := provider.DBProvider("blocks", cfg)
	if err != nil {
		return nil, err
	}

	blockStore := store.NewStoreBlock(blockStoreDB)

	stateDB, err := provider.DBProvider("state", cfg)
	if err != nil {
		return nil, err
	}
	stateStore := state2.NewStoreState(stateDB)

	genesis, err := provider.GenesisProvider(cfg)
	if err != nil {
		return nil, err
	}

	stat := stateStore.LoadFromDBOrGenesis(genesis)

	nodeInfo := &p2p.NodeInfo{
		PublicKey:   nodeKey.PublicKey.ToBytes(),
		NodeID:      nodeKey.GetID(),
		ListenAddr:  cfg.P2PConfig.ListenAddress,
		Channels:    []byte{p2p.LeaderProposeChannel, p2p.ReplicaVoteChannel, p2p.ReplicaNextViewChannel, p2p.TxsChannel},
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
	txsPool.SetLogger(logger)

	blockExec := state2.NewBlockExecutor(cfg, stateStore, blockStore, proxyAppConns.Consensus(), txsPool, logger.New("module", "state"))

	consensusCore, consensusReactor := provider.ConsensusProvider(cfg, stat, blockExec, txsPool, nodeKey.PrivateKey, nodeInfo.CryptoBLS12, logger)
	consensusCore.SetEventBus(eventBus)

	syncerReactor := provider.SyncerProvider(stat, blockExec, blockStore, logger)

	stchReactor := provider.STCHProvider(nodeInfo.ID(), len(cfg.P2PConfig.NeighboursSlice()), logger)
	stchReactor.Chameleon().Init(kp)
	stat.SetChameleon(stchReactor.Chameleon())
	stat.SetBlockStore(blockStore)
	stchReactor.Chameleon().SetBlockStore(blockStore)
	transport, sw := provider.P2PProvider(cfg, nodeInfo, nodeKey, txsPoolReactor, consensusReactor, syncerReactor, stchReactor, logger)

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

func (n *Node) State() *state2.State {
	return n.consensusReactor.State()
}
