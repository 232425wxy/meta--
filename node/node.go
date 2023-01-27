package node

import (
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/consensus"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/p2p"
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

func NewNode(cfg *config.Config, logger log.Logger) *Node {
	nodeKey, err := p2p.LoadNodeKey(cfg.BasicConfig.KeyFilePath())
	if err != nil {
		panic(err)
	}

}

func NewNode(cfg *config.Config, logger log.Logger, dbProvider DBProvider, genesisProvider GenesisProvider, nodeKey *p2p.NodeKey) (*Node, error) {
	eventBus, err := events.CreateAndStartEventBus(logger)
	if err != nil {
		return nil, err
	}

	blockStoreDB, err := dbProvider("blocks", cfg)
	if err != nil {
		return nil, err
	}
	blockStore := state.NewStoreBlock(blockStoreDB)

	stateDB, err := dbProvider("state", cfg)
	if err != nil {
		return nil, err
	}
	stateStore := state.NewStoreState(stateDB)

	genesis, err := genesisProvider(cfg)
	if err != nil {
		return nil, err
	}

	state := stateStore.LoadFromDBOrGenesis(genesis)

	nodeInfo := &p2p.NodeInfo{
		NodeID:     nodeKey.GetID(),
		ListenAddr: cfg.P2PConfig.ListenAddress,
		Channels:   []byte{p2p.LeaderProposeChannel, p2p.ReplicaVoteChannel, p2p.ReplicaStateChannel, p2p.TxsChannel},
		RPCAddress: "",
		TxIndex:    "on",
	}

}
