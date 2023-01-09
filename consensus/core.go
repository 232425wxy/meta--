package consensus

import (
	"bytes"
	"fmt"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/proto/pbtypes"
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
	cryptoBLS12      *bls12.CryptoBLS12
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
	case mi := <-c.internalMsgQueue:
		c.handleMsg(mi)
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
	case NewHeightStep:
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
	case NewHeightStep:
		c.enterNewRound(info.Height, 0)
	case NewRoundStep:
		c.enterPrepareStep(info.Height, 0)
	case PrepareStep:
		c.enterPrepareVoteStep(info.Height, info.Round) // TODO 存疑
	}
}

func (c *Core) handleMsg(mi MessageInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m, nodeID := mi.Msg, mi.NodeID
	var err error

	switch msg := m.(type) {
	case *types.Prepare:
		err = c.handlePrepare(msg)
	case *types.PrepareVote:
		err = c.handlePrepareVote(msg)
	case *types.PreCommit:
		err = c.handlePreCommit(msg)
	case *types.PreCommitVote:
		err = c.handlePreCommitVote(msg)
	}
}

func (c *Core) handlePrepare(prepare *types.Prepare) error {
	if c.stepInfo.prepare != nil {
		return nil
	}
	if prepare.Height != c.stepInfo.height {
		return nil
	}
	hash := sha256.Hash{}
	copy(hash[:], prepare.Block.Header.Hash)
	ok := c.validators.GetLeader().PublicKey.Verify(prepare.Signature, hash)
	if !ok {
		return fmt.Errorf("node %s sent an invalid prepare message to me", prepare.Signature.Signer())
	}
	c.stepInfo.prepare = prepare
	c.stepInfo.block = prepare.Block
	if c.validators.GetLeader().ID != c.publicKey.ToID() {
		c.Logger.Info("received Prepare message, plan to vote for it", "from", prepare.Signature.Signer())
		// 如果我自己不是主节点，那么在收到Prepare消息后，就应该进入PREPARE_VOTE阶段
	}
	c.enterPrepareVoteStep(c.stepInfo.height, c.stepInfo.round)
	return nil
}

func (c *Core) handlePrepareVote(vote *types.PrepareVote) error {
	if c.validators.GetLeader().ID != c.publicKey.ToID() {
		// 不是主节点，直接忽略
		return nil
	}
	if vote.Vote.Height != c.stepInfo.height {
		return fmt.Errorf("invalid PrepareVote message, my height: %d, message's height: %d", c.stepInfo.height, vote.Vote.Height)
	}
	if vote.Vote.VoteType != pbtypes.PrepareVoteType {
		return fmt.Errorf("invalid vote type, want %s, got %s", pbtypes.PrepareVoteType.String(), vote.Vote.VoteType.String())
	}
	validator := c.validators.GetValidatorByID(vote.Vote.Signature.Signer())
	if validator == nil {
		return fmt.Errorf("cannot find this validator: %s", vote.Vote.Signature.Signer())
	}
	valueHash := types.GeneratePrepareVoteValueHash(c.stepInfo.block.Header.Hash)
	equal := bytes.Equal(valueHash[:], vote.Vote.ValueHash[:])
	if !equal {
		return fmt.Errorf("validator %s vote for different block", vote.Vote.Signature.Signer())
	}
	ok := validator.PublicKey.Verify(vote.Vote.Signature, vote.Vote.ValueHash)
	if !ok {
		return fmt.Errorf("validator %s sent invalid PrepareVote message to me", vote.Vote.Signature.Signer())
	}
	c.stepInfo.heightVoteSet.AddPrepareVote(c.stepInfo.round, vote)
	ok = c.stepInfo.heightVoteSet.CheckPrepareVoteIsComplete(c.stepInfo.round)
	if ok {
		// 收集齐了副本节点对Prepare消息的投票，那么开始构造PreCommit消息
		c.enterPreCommitStep(c.stepInfo.height, c.stepInfo.round)
	}
	return nil
}

func (c *Core) handlePreCommit(preCommit *types.PreCommit) error {
	if c.stepInfo.preCommit != nil {
		return nil
	}
	if preCommit.Height != c.stepInfo.height {
		return nil
	}
	if c.stepInfo.block == nil {
		return nil
	}
	if preCommit.ID != c.validators.GetLeader().ID {
		return fmt.Errorf("PreCommit message is not from leader %s at height %d", c.validators.GetLeader().ID, c.stepInfo.height)
	}
	hash := types.GeneratePreCommitValueHash(c.stepInfo.block.Header.Hash)
	equal := bytes.Equal(hash[:], preCommit.ValueHash[:])
	if !equal {
		return fmt.Errorf("leader %s sent invalid PreCommit message to me", preCommit.ID)
	}
	ok := c.cryptoBLS12.VerifyThresholdSignature(preCommit.AggregateSignature, preCommit.ValueHash)
	if !ok {
		return fmt.Errorf("leader %s sent invalid PreCommit message to me, aggregated signature is wrong", preCommit.ID)
	}
	c.stepInfo.preCommit = preCommit
	if c.validators.GetLeader().ID != c.publicKey.ToID() {
		c.Logger.Info("received PreCommit message, plan to vote for it", "from", preCommit.ID)
	}
	c.enterPreCommitVoteStep(c.stepInfo.height, c.stepInfo.round)
	return nil
}

func (c *Core) handlePreCommitVote(vote *types.PreCommitVote) error {
	if c.validators.GetLeader().ID != c.publicKey.ToID() {
		return nil
	}
	if vote.Vote.Height != c.stepInfo.height {
		return fmt.Errorf("invalid PreCommitVote message, my height: %d, message's height: %d", c.stepInfo.height, vote.Vote.Height)
	}
	if vote.Vote.VoteType != pbtypes.PreCommitVoteType {
		return fmt.Errorf("invalid vote type, want %s, got %s", pbtypes.PreCommitVoteType.String(), vote.Vote.VoteType.String())
	}
	validator := c.validators.GetValidatorByID(vote.Vote.Signature.Signer())
	if validator == nil {
		return fmt.Errorf("cannot find this validator: %s", vote.Vote.Signature.Signer())
	}
	valueHash := types.GeneratePreCommitVoteValueHash(c.stepInfo.block.Header.Hash)
	equal := bytes.Equal(valueHash[:], vote.Vote.ValueHash[:])
	if !equal {
		return fmt.Errorf("validator %s vote for different block", vote.Vote.Signature.Signer())
	}
	ok := validator.PublicKey.Verify(vote.Vote.Signature, vote.Vote.ValueHash)
	if !ok {
		return fmt.Errorf("validator %s sent invalid PrepareVote message to me", vote.Vote.Signature.Signer())
	}
	c.stepInfo.heightVoteSet.AddPreCommitVote(c.stepInfo.round, vote)
	ok = c.stepInfo.heightVoteSet.CheckPreCommitVoteIsComplete(c.stepInfo.round)
	if ok {
		c.enterCommit(c.stepInfo.height, c.stepInfo.round)
	}
	return nil
}

func (c *Core) enterNewRound(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	logger.Info("entering NEW ROUND step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	if c.stepInfo.height != height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step != NewHeightStep) {
		logger.Warn("entering NEW ROUND step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	c.stepInfo.round = round
	c.stepInfo.step = NewRoundStep
	if round != 0 {
		c.stepInfo.prepare = nil
		c.stepInfo.block = nil
	}
	c.stepInfo.heightVoteSet.round = round + 1
	c.enterPrepareStep(height, round)
}

func (c *Core) enterPrepareStep(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	logger.Info("entering PREPARE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	if c.stepInfo.height != height || round < c.stepInfo.round || (c.stepInfo.round == round && PrepareStep <= c.stepInfo.step) {
		logger.Warn("entering PREPARE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}

	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PrepareStep
		c.newStep()
		if c.hasPrepare() {
			c.enterPrepareVoteStep(height, round)
		}
	}()
	// 非leader节点在将来收到prepare消息后会将这个超时时间顶替掉
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
	logger.Info("entering PREPARE_VOTE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	if c.stepInfo.height != height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PrepareVoteStep) {
		logger.Warn("entering PREPARE_VOTE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PrepareVoteStep
		c.newStep()
	}()

	if !c.hasPrepare() {
		logger.Error("PREPARE_VOTE step: Prepare message is nil")
		return
	}
	logger.Debug("Prepare message is valid, decide to vote for it")
	vote := types.NewPrepareVote(height, types.GeneratePrepareVoteValueHash(c.stepInfo.prepare.Block.Header.Hash), c.privateKey)
	c.sendInternalMessage(MessageInfo{Msg: vote, NodeID: ""})
}

func (c *Core) enterPreCommitStep(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PreCommitStep) {
		logger.Warn("entering PRE_COMMIT step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	}
	logger.Info("entering PRE_COMMIT step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PreCommitStep
		c.newStep()
	}()
	agg := c.stepInfo.heightVoteSet.CreateThresholdSigForPrepareVote(round, c.cryptoBLS12)
	preCommit := types.NewPreCommit(agg, types.GeneratePreCommitValueHash(c.stepInfo.block.Header.Hash), c.publicKey.ToID(), height)
	c.sendInternalMessage(MessageInfo{Msg: preCommit, NodeID: ""})
}

func (c *Core) enterPreCommitVoteStep(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PrepareVoteStep) {
		logger.Warn("entering PRE_COMMIT_VOTE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	}
	logger.Info("entering PRE_COMMIT_VOTE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.height = height
		c.stepInfo.round = round
		c.stepInfo.step = PreCommitVoteStep
	}()
	if !c.hasPreCommit() {
		logger.Error("PRE_COMMIT_VOTE step: PreCommit message is nil")
		return
	}
	logger.Debug("PreCommit message is valid, decide to vote for it")
	vote := types.NewPreCommitVote(height, types.GeneratePreCommitVoteValueHash(c.stepInfo.block.Header.Hash), c.privateKey)
	c.sendInternalMessage(MessageInfo{Msg: vote, NodeID: ""})
}

func (c *Core) enterCommit(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PreCommitStep) {
		logger.Warn("entering COMMIT step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	}
	logger.Info("entering COMMIT step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = CommitStep
		c.newStep()
	}()
	agg := c.stepInfo.heightVoteSet.CreateThresholdSigForPreCommitVote(round, c.cryptoBLS12)
	commit := types.NewCommit(agg, types.GenerateCommitValueHash(c.stepInfo.block.Header.Hash), c.publicKey.ToID(), height)
	c.sendInternalMessage(MessageInfo{Msg: commit, NodeID: ""})
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

func (c *Core) hasPreCommit() bool {
	return c.stepInfo.preCommit != nil
}

func (c *Core) newStep() {
	esi := c.stepInfo.EventStepInfo()
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
	c.scheduledTicker.ScheduleTimeout(timeoutInfo{Duration: duration, Height: stepInfo.height, Round: 0, Step: NewHeightStep})
}

func (c *Core) sendInternalMessage(info MessageInfo) {
	select {
	case c.internalMsgQueue <- info:
	default:
		go func() { c.internalMsgQueue <- info }()
	}
}
