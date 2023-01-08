package consensus

import (
	"fmt"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/state"
	"github.com/232425wxy/meta--/txspool"
	"github.com/232425wxy/meta--/types"
	"runtime/debug"
	"sync"
	"time"
)

type Core struct {
	service.BaseService
	cfg              *config.ConsensusConfig
	privateKey       *bls12.PrivateKey // 为共识消息签名的私钥
	publicKey        *bls12.PublicKey
	validators       *types.ValidatorSet
	id               crypto.ID            // 自己的节点ID
	blockStore       *state.StoreBlock    // 存储区块，也可以通过区块高度和区块哈希值加载指定的区块
	blockExec        *state.BlockExecutor // 创建区块和执行区块里的交易指令
	state            *state.State
	txsPool          *txspool.TxsPool
	eventBus         *events.EventBus
	eventSwitch      *events.EventSwitch
	stepInfo         *StepInfo
	scheduledTicker  *TimeoutTicker
	internalMsgQueue chan MessageInfo
	peerMsgQueue     chan MessageInfo
	mu               sync.RWMutex
}

func (c *Core) SetLogger(logger log.Logger) {
	c.BaseService.SetLogger(logger)
}

func (c *Core) SetEventBus(bus *events.EventBus) {
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

	case tock := <-c.scheduledTicker.TockChan():
		c.handleScheduled(tock, *c.stepInfo)
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

func (c *Core) handleScheduled(info timeoutInfo, stepInfo StepInfo) {
	if info.Height != stepInfo.height || info.Round < stepInfo.round || (info.Round == stepInfo.round && info.Step < stepInfo.step) {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	switch info.Step {
	case NewViewStep:
		c.enterNewRound(info.Height, 0)
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
		c.newStep()
		if c.hasPrepare() {
			c.enterPrepareVoteStep(height, round)
		}
	}()
	c.scheduleStep(c.cfg.TimeoutPrepare, height, round, PrepareStep) // 计划提出Prepare消息
	if c.validators.GetLeader().ID == c.publicKey.ToID() {
		logger.Debug("leader is me, it's my responsibility to propose Prepare message", "validator_id", c.publicKey.ToID())
		// 开始打包Prepare消息
		var block *types.Block
		if c.stepInfo.block != nil {
			block = c.stepInfo.block
		} else {
			block = c.createBlock()
			if block == nil {
				return
			}
		}
		prepare := types.NewPrepare(height, block, c.publicKey.ToID(), c.privateKey)
		c.sendInternalMessage(MessageInfo{Msg: prepare, NodeID: ""})
	}
}

func (c *Core) enterPrepareVoteStep(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if c.stepInfo.height != height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PrepareVoteStep) {
		logger.Warn("entering PREPARE_VOTE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PrepareVoteStep
		c.newStep()
	}()
	logger.Info("entering PREPARE_VOTE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))

	if !c.hasPrepare() {
		logger.Error("PREPARE_VOTE step: Prepare message is nil")
		return
	}
	logger.Debug("Prepare message is valid, decide to vote for it")
	vote := types.NewPrepareVote(height, c.stepInfo.prepare.Hash(), c.privateKey)
	c.sendInternalMessage(MessageInfo{Msg: vote, NodeID: ""})
}

func (c *Core) createBlock() *types.Block {
	switch {
	case c.stepInfo.height == c.state.InitialHeight:
		lastBlockHash := sha256.Sum([]byte("first block"))
		return c.blockExec.CreateBlock(c.stepInfo.height, c.state, c.publicKey.ToID(), lastBlockHash[:])
	case c.stepInfo.previousBlock != nil:
		return c.blockExec.CreateBlock(c.stepInfo.height, c.state, c.publicKey.ToID(), c.stepInfo.previousBlock.Header.Hash)
	default:
		c.Logger.Error("PREPARE step: cannot propose Prepare message")
		return nil
	}
}

func (c *Core) hasPrepare() bool {
	return c.stepInfo.prepare != nil
}

func (c *Core) newStep() {
	esi := c.stepInfo.EventStepInfo()
	if err := c.eventBus.PublishEventNewRoundStep(esi); err != nil {
		c.Logger.Error("failed to publish new round step", "err", err)
	}
	c.eventSwitch.FireEvent(events.EventNewRoundStep, &esi)
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
