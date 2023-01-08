package consensus

import (
	"fmt"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/event"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/state"
	"github.com/232425wxy/meta--/txspool"
	"runtime/debug"
	"sync"
	"time"
)

type Core struct {
	service.BaseService
	cfg              *config.ConsensusConfig
	privateKey       *bls12.PrivateKey // 为共识消息签名的私钥
	publicKey        *bls12.PublicKey
	id               crypto.ID            // 自己的节点ID
	blockStore       *state.StoreBlock    // 存储区块，也可以通过区块高度和区块哈希值加载指定的区块
	blockExec        *state.BlockExecutor // 创建区块和执行区块里的交易指令
	state            *state.State
	txsPool          *txspool.TxsPool
	eventBus         *event.EventBus
	stepInfo         *StepInfo
	scheduledTicker  *TimeoutTicker
	internalMsgQueue chan MessageInfo
	peerMsgQueue     chan MessageInfo
	mu               sync.RWMutex
}

func (c *Core) SetLogger(logger log.Logger) {
	c.BaseService.SetLogger(logger)
}

func (c *Core) SetEventBus(bus *event.EventBus) {
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

func (c *Core) handleAvailableTxs() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stepInfo.round != 0 {
		// 早在之前的round 0阶段已经打包好交易数据了
		return
	}
	switch c.stepInfo.step {
	case NewViewStep:
		if c.isFirstBlock(c.stepInfo.height) {
			return
		}
		duration := c.stepInfo.startTime.Sub(time.Now()) + time.Millisecond
		c.scheduleStep(duration, c.stepInfo.height, 0, NewRoundStep)
	case NewRoundStep:
		c.enterPrepareStep(c.stepInfo.height, 0)
	}
}

func (c *Core) enterPrepareStep(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if c.stepInfo.height != height || round < c.stepInfo.round || (c.stepInfo.round == round && PrepareStep <= c.stepInfo.step) {
		logger.Warn("entering PREPARE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	logger.Info("entering PREPARE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))

	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PrepareStep

	}()
}

func (c *Core) newStep() {
	esi := c.stepInfo.EventStepInfo()
	if err := c.eventBus.PublishEventNewRoundStep(esi); err != nil {
		c.Logger.Error("failed to publish new round step", "err", err)
	}

}

func (c *Core) isFirstBlock(height int64) bool {
	if height == c.state.InitialHeight {
		return true
	}
	return false
}

func (c *Core) scheduleStep(duration time.Duration, height int64, round int16, step Step) {
	c.scheduledTicker.ScheduleTimeout(timeoutInfo{Duration: duration, Height: height, Round: round, Step: step})
}

func (c *Core) scheduleRound0(stepInfo *StepInfo) {
	duration := stepInfo.startTime.Sub(time.Now())
	c.scheduledTicker.ScheduleTimeout(timeoutInfo{Duration: duration, Height: stepInfo.height, Round: 0, Step: NewViewStep})
}

func (c *Core) sendInternalMessage(info MessageInfo) {
	select {
	case c.internalMsgQueue <- info:
	default:
		go func() { c.internalMsgQueue <- info }()
	}
}
