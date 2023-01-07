package consensus

import (
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/state"
	"github.com/232425wxy/meta--/txspool"
	"github.com/232425wxy/meta--/types"
	"runtime/debug"
)

type Core struct {
	service.BaseService
	cfg              *config.ConsensusConfig
	signerPrivateKey *bls12.PrivateKey    // 为共识消息签名的私钥
	blockStore       *state.StoreBlock    // 存储区块，也可以通过区块高度和区块哈希值加载指定的区块
	blockExec        *state.BlockExecutor // 创建区块和执行区块里的交易指令
	state            *state.State
	txsPool          *txspool.TxsPool
	eventBus         *types.EventBus
	stepInfo         *StepInfo
}

func (c *Core) SetLogger(logger log.Logger) {
	c.BaseService.SetLogger(logger)
}

func (c *Core) SetEventBus(bus *types.EventBus) {
	c.eventBus = bus
	c.blockExec.SetEventBUs(bus)
}

func (c *Core) receiveRoutine() {
	defer func() {
		if r := recover(); r != nil {
			c.Logger.Error("CONSENSUS FAILURE!!!", "err", r, "stack", string(debug.Stack()))
		}
	}()

	select {
	case <-c.txsPool.TxsAvailable():

	}
}