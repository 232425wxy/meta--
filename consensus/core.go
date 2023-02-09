package consensus

import "C"
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

const msgQueueSize = 10000

type Core struct {
	service.BaseService
	cfg                 *config.ConsensusConfig
	privateKey          *bls12.PrivateKey // 为共识消息签名的私钥
	publicKey           *bls12.PublicKey
	id                  crypto.ID            // 自己的节点ID
	blockStore          *state.StoreBlock    // 存储区块，也可以通过区块高度和区块哈希值加载指定的区块
	blockExec           *state.BlockExecutor // 创建区块和执行区块里的交易指令
	state               *state.State
	txsPool             *txspool.TxsPool
	eventBus            *events.EventBus
	eventSwitch         *events.EventSwitch
	stepInfo            *StepInfo
	scheduledTicker     *TimeoutTicker
	internalMsgQueue    chan MessageInfo
	externalMsgQueue    chan MessageInfo
	prepareVotesQueue   chan *types.PrepareVote
	preCommitVotesQueue chan *types.PreCommitVote
	commitVotesQueue    chan *types.CommitVote
	mu                  sync.RWMutex
	cryptoBLS12         *bls12.CryptoBLS12
}

func NewCore(cfg *config.ConsensusConfig, privateKey *bls12.PrivateKey, state *state.State, blockExec *state.BlockExecutor, blockStore *state.StoreBlock, txsPool *txspool.TxsPool, cryptoBLS12 *bls12.CryptoBLS12) *Core {
	core := &Core{
		BaseService:         *service.NewBaseService(nil, "Consensus_Core"),
		cfg:                 cfg,
		privateKey:          privateKey,
		publicKey:           privateKey.PublicKey(),
		id:                  privateKey.PublicKey().ToID(),
		blockStore:          blockStore,
		blockExec:           blockExec,
		state:               state,
		txsPool:             txsPool,
		eventSwitch:         events.NewEventSwitch(),
		stepInfo:            NewStepInfo(),
		scheduledTicker:     NewTimeoutTicker(),
		internalMsgQueue:    make(chan MessageInfo, msgQueueSize),
		externalMsgQueue:    make(chan MessageInfo, msgQueueSize),
		prepareVotesQueue:   make(chan *types.PrepareVote, msgQueueSize/100),
		preCommitVotesQueue: make(chan *types.PreCommitVote, msgQueueSize/100),
		commitVotesQueue:    make(chan *types.CommitVote, msgQueueSize/100),
		cryptoBLS12:         cryptoBLS12,
	}
	core.stepInfo.height = state.InitialHeight
	core.stepInfo.startTime = time.Now().Add(time.Second)
	return core
}

func (c *Core) Start() error {
	if err := c.scheduledTicker.Start(); err != nil {
		return err
	}
	go c.receiveRoutine()
	//c.scheduleRound0(c.stepInfo)
	return nil
}

func (c *Core) SetLogger(logger log.Logger) {
	c.BaseService.SetLogger(logger)
	//c.scheduledTicker.SetLogger(logger.New("module", "Ticker"))
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
	for {
		select {
		case <-c.txsPool.TxsAvailable():
			c.handleAvailableTxs()
		case tock := <-c.scheduledTicker.TockChan():
			c.handleScheduled(tock, *c.stepInfo)
		case mi := <-c.internalMsgQueue:
			c.handleMsg(mi)
		case mi := <-c.externalMsgQueue:
			c.handleMsg(mi)
		case <-c.WaitStop():
			return
		}
	}
}

func (c *Core) handleAvailableTxs() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stepInfo.round != 0 {
		// 早在之前的round 0阶段已经打包好交易数据了
		return
	}

	if c.isLeader() && (c.stepInfo.step == NewHeightStep || c.stepInfo.step == NewRoundStep) {
		c.scheduleStep(time.Second, c.stepInfo.height, c.stepInfo.round, PrepareStep)
	}
}

func (c *Core) handleScheduled(info timeoutInfo, stepInfo StepInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch info.Step {
	case NewHeightStep:
		c.stepInfo.Reset()
	case PrepareStep:
		c.proposePrepareMsg(c.stepInfo.height, c.stepInfo.round)
	case PreCommitStep:
		c.proposePreCommitMsg(c.stepInfo.height, c.stepInfo.round)
	case CommitStep:
		c.proposeCommitMsg(c.stepInfo.height, c.stepInfo.round)
	case DecideStep:
		c.proposeDecideMsg(c.stepInfo.height, c.stepInfo.round)
	}
}

func (c *Core) handleMsg(mi MessageInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error

	switch msg := mi.Msg.(type) {
	case *types.NextView:
		err = c.handleNextView(msg)
		if err != nil {
			c.Logger.Error("failed to handle NextView message", "err", err)
			err = nil
		}
	case *types.Prepare:
		err = c.handlePrepare(msg)
		if err != nil {
			c.Logger.Error("failed to handle Prepare message", "err", err)
			err = nil
		}
	case *types.PrepareVote:
		if mi.NodeID != "" {
			err = c.handlePrepareVote(msg)
			if err != nil {
				c.Logger.Error("failed to handle PrepareVote message", "err", err)
				err = nil
			}
		} else {
			select {
			case c.prepareVotesQueue <- msg:
			default:
				go func() { c.prepareVotesQueue <- msg }()
			}
		}
	case *types.PreCommit:
		err = c.handlePreCommit(msg)
		if err != nil {
			c.Logger.Error("failed to handle PreCommit message", "err", err)
			err = nil
		}
	case *types.PreCommitVote:
		if mi.NodeID != "" { // 这表示自己是主节点，收到了其他副本节点发送来的投票
			err = c.handlePreCommitVote(msg)
			if err != nil {
				c.Logger.Error("failed to handle PreCommitVote message", "err", err)
				err = nil
			}
		} else {
			select {
			case c.preCommitVotesQueue <- msg:
			default:
				go func() { c.preCommitVotesQueue <- msg }()
			}
		}
	case *types.Commit:
		err = c.handleCommit(msg)
		if err != nil {
			c.Logger.Error("failed to handle Commit message", "err", err)
			err = nil
		}
	case *types.CommitVote:
		if mi.NodeID != "" { // 这表示自己是主节点，收到了其他副本节点发送来的投票
			err = c.handleCommitVote(msg)
			if err != nil {
				c.Logger.Error("failed to handle CommitVote message", "err", err)
				err = nil
			}
		} else {
			select {
			case c.commitVotesQueue <- msg:
			default:
				go func() { c.commitVotesQueue <- msg }()
			}
		}
	case *types.Decide:
		err = c.handleDecide(msg)
		if err != nil {
			c.Logger.Error("failed to handle Decide message", "err", err)
			err = nil
		}
	default:
		c.Logger.Error("unknown message type", "type", fmt.Sprintf("%T", msg))
		return
	}
}

func (c *Core) handleNextView(view *types.NextView) error {
	if !c.isNextLeader() {
		// 只有主节点才会处理其他节点发送过来的NextView消息
		return nil
	}
	if view.Type != pbtypes.NextViewType {
		return fmt.Errorf("want message type %s, but got %s", pbtypes.NextViewType.String(), view.Type.String())
	}
	if validator := c.state.Validators.GetValidatorByID(view.ID); validator == nil {
		return fmt.Errorf("an unknown validator %s sent NextView message to me", view.ID)
	}
	if view.Height != c.stepInfo.height+1 {
		return fmt.Errorf("validator %s sent invalid NextView message to me, because \"height\" is wrong", view.ID)
	}
	//c.Logger.Debug("receive a valid NextView message", "from", view.ID)
	c.stepInfo.AddNextView(view)
	if c.stepInfo.CheckCollectNextViewIsComplete(c.state.Validators) {
		c.Logger.Info("receive enough NextView messages", "height", c.stepInfo.height)
		c.stepInfo.height += 1
		c.scheduleRound0(c.stepInfo)
	}
	return nil
}

func (c *Core) handlePrepare(prepare *types.Prepare) error {
	if prepare.Height != c.stepInfo.height {
		return nil
	}
	hash := sha256.Hash{}
	copy(hash[:], prepare.Block.Header.Hash)
	ok := c.state.Validators.GetLeader(c.stepInfo.height).PublicKey.Verify(prepare.Signature, hash)
	if !ok {
		if c.isLeader() {
			panic("why I created an invalid Prepare message?")
		}
		return fmt.Errorf("leader %s sent an invalid prepare message to me", prepare.Signature.Signer())
	}
	if c.isLeader() {
		c.stepInfo.prepare <- prepare // reactor循环检测c.stepInfo.prepare是否有东西，有的话就发送给其他节点
	}
	c.stepInfo.block = prepare.Block
	c.enterPrepareVoteStep(c.stepInfo.height, c.stepInfo.round)
	return nil
}

func (c *Core) handlePrepareVote(vote *types.PrepareVote) error {
	if !c.isLeader() {
		// 不是主节点，直接忽略
		return nil
	}
	if vote.Vote.Height != c.stepInfo.height {
		return fmt.Errorf("invalid PrepareVote message, my height: %d, message's height: %d", c.stepInfo.height, vote.Vote.Height)
	}
	if vote.Vote.VoteType != pbtypes.PrepareVoteType {
		return fmt.Errorf("invalid vote type, want %s, got %s", pbtypes.PrepareVoteType.String(), vote.Vote.VoteType.String())
	}
	validator := c.state.Validators.GetValidatorByID(vote.Vote.Signature.Signer())
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
	c.stepInfo.voteSet.AddPrepareVote(c.stepInfo.round, vote)
	ok = c.stepInfo.voteSet.CheckPrepareVoteIsComplete(c.stepInfo.round, c.state.Validators)
	if ok { // TODO 这里需要搞一个超时机制，就是哪怕收到了足够数量的投票，也不要立即去组装门陷签名，防止接下来还会有投票过来
		// 收集齐了副本节点对Prepare消息的投票，那么开始构造PreCommit消息
		//c.proposePreCommitMsg(c.stepInfo.height, c.stepInfo.round)
		c.scheduleStep(time.Millisecond*100, c.stepInfo.height, c.stepInfo.round, PreCommitStep)
	}
	return nil
}

func (c *Core) handlePreCommit(preCommit *types.PreCommit) error {
	if preCommit.Height != c.stepInfo.height {
		return nil
	}
	if c.stepInfo.block == nil {
		return nil
	}
	if preCommit.ID != c.state.Validators.GetLeader(c.stepInfo.height).ID {
		return fmt.Errorf("PreCommit message is not from leader %s at height %d", c.state.Validators.GetLeader(c.stepInfo.height).ID, c.stepInfo.height)
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
	if c.isLeader() {
		c.stepInfo.preCommit <- preCommit
	}
	c.enterPreCommitVoteStep(c.stepInfo.height, c.stepInfo.round)
	return nil
}

func (c *Core) handlePreCommitVote(vote *types.PreCommitVote) error {
	if !c.isLeader() {
		return nil
	}
	if vote.Vote.Height != c.stepInfo.height {
		return fmt.Errorf("invalid PreCommitVote message, my height: %d, message's height: %d", c.stepInfo.height, vote.Vote.Height)
	}
	if vote.Vote.VoteType != pbtypes.PreCommitVoteType {
		return fmt.Errorf("invalid vote type, want %s, got %s", pbtypes.PreCommitVoteType.String(), vote.Vote.VoteType.String())
	}
	validator := c.state.Validators.GetValidatorByID(vote.Vote.Signature.Signer())
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
		return fmt.Errorf("validator %s sent invalid PreCommitVote message to me", vote.Vote.Signature.Signer())
	}
	c.stepInfo.voteSet.AddPreCommitVote(c.stepInfo.round, vote)
	ok = c.stepInfo.voteSet.CheckPreCommitVoteIsComplete(c.stepInfo.round, c.state.Validators)
	if ok {
		//c.proposeCommitMsg(c.stepInfo.height, c.stepInfo.round)
		c.scheduleStep(time.Millisecond*100, c.stepInfo.height, c.stepInfo.round, CommitStep)
	}
	return nil
}

func (c *Core) handleCommit(commit *types.Commit) error {
	if c.stepInfo.height != commit.Height {
		return nil
	}
	if c.stepInfo.block == nil {
		return nil
	}
	if commit.ID != c.state.Validators.GetLeader(c.stepInfo.height).ID {
		return fmt.Errorf("Commit message is not from leader %s at height %d", c.state.Validators.GetLeader(c.stepInfo.height).ID, c.stepInfo.height)
	}
	hash := types.GenerateCommitValueHash(c.stepInfo.block.Header.Hash)
	equal := bytes.Equal(hash[:], commit.ValueHash[:])
	if !equal {
		return fmt.Errorf("leader %s sent invalid Commit message to me", commit.ID)
	}
	ok := c.cryptoBLS12.VerifyThresholdSignature(commit.AggregateSignature, hash)
	if !ok {
		return fmt.Errorf("leader %s sent invalid Commit message to me, aggregated signature is wrong", commit.ID)
	}
	if c.isLeader() {
		c.stepInfo.commit <- commit
	}
	c.enterCommitVoteStep(c.stepInfo.height, c.stepInfo.round)
	return nil
}

func (c *Core) handleCommitVote(vote *types.CommitVote) error {
	if !c.isLeader() {
		return nil
	}
	if vote.Vote.Height != c.stepInfo.height {
		return fmt.Errorf("invalid CommitVote message, my height: %d, message's height: %d", c.stepInfo.height, vote.Vote.Height)
	}
	if vote.Vote.VoteType != pbtypes.CommitVoteType {
		return fmt.Errorf("invalid vote type, want %s, got %s", pbtypes.CommitVoteType.String(), vote.Vote.VoteType.String())
	}
	validator := c.state.Validators.GetValidatorByID(vote.Vote.Signature.Signer())
	if validator == nil {
		return fmt.Errorf("cannot find this validator: %s", vote.Vote.Signature.Signer())
	}
	valueHash := types.GenerateCommitVoteValueHash(c.stepInfo.block.Header.Hash)
	equal := bytes.Equal(valueHash[:], vote.Vote.ValueHash[:])
	if !equal {
		return fmt.Errorf("validator %s vote for different block", vote.Vote.Signature.Signer())
	}
	ok := validator.PublicKey.Verify(vote.Vote.Signature, vote.Vote.ValueHash)
	if !ok {
		return fmt.Errorf("validator %s sent invalid CommitVote message to me", vote.Vote.Signature.Signer())
	}
	c.stepInfo.voteSet.AddCommitVote(c.stepInfo.round, vote)
	ok = c.stepInfo.voteSet.CheckCommitVoteIsComplete(c.stepInfo.round, c.state.Validators)
	if ok {
		c.scheduleStep(time.Millisecond*100, c.stepInfo.height, c.stepInfo.round, DecideStep)
	}
	return nil
}

func (c *Core) handleDecide(decide *types.Decide) error {
	if c.stepInfo.height != decide.Height {
		return nil
	}
	if c.stepInfo.block == nil {
		return nil
	}
	if decide.ID != c.state.Validators.GetLeader(c.stepInfo.height).ID {
		return fmt.Errorf("Decide message is not from leader %s at height %d", c.state.Validators.GetLeader(c.stepInfo.height).ID, c.stepInfo.height)
	}
	hash := types.GenerateDecideValueHash(c.stepInfo.block.Header.Hash)
	equal := bytes.Equal(hash[:], decide.ValueHash[:])
	if !equal {
		return fmt.Errorf("leader %s sent invalid Decide message to me", decide.ID)
	}
	ok := c.cryptoBLS12.VerifyThresholdSignature(decide.AggregateSignature, hash)
	if !ok {
		return fmt.Errorf("leader %s sent invalid Decide message to me, aggregated signature is wrong", decide.ID)
	}
	if c.isLeader() {
		c.stepInfo.decide <- decide
	}
	c.stepInfo.startTime = decide.Timestamp
	if !c.isLeader() {
		c.stepInfo.step = DecideStep
		c.newStep()
	}
	c.applyBlock()
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
		c.stepInfo.block = nil
		c.stepInfo.prepare = nil
		c.stepInfo.preCommit = nil
		c.stepInfo.commit = nil
		c.stepInfo.decide = nil
	}
	c.proposePrepareMsg(height, round)
}

func (c *Core) proposePrepareMsg(height int64, round int16) {
	if c.stepInfo.height != height || round < c.stepInfo.round || (c.stepInfo.round == round && PrepareStep <= c.stepInfo.step) {
		c.Logger.Warn("entering PREPARE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	c.Logger.Info("=> PREPARE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PrepareStep
		c.newStep()
	}()
	// 作为一个普通节点，执行到此处该方法就结束了，接下来就是等待主节点发送来Prepare消息
	if c.isLeader() {
		c.Logger.Trace("leader is me, it's my responsibility to propose Prepare message", "validator_id", c.publicKey.ToID())
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
		// 将Prepare消息发送到内部的消息通道里，这样在recvRoutine进程中可以捕获该消息，然后就会去处理该消息
		c.sendInternalMessage(MessageInfo{Msg: prepare, NodeID: ""})
	}
}

// enterPrepareVoteStep 主节点生成Prepare消息，并将其广播给其他节点，自己也保留这个Prepare消息，然后拥有Prepare消息的节点
// 为Prepare消息进行投票，主节点也会为自己生成的Prepare消息进行投票。
func (c *Core) enterPrepareVoteStep(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if c.stepInfo.height != height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PrepareVoteStep) {
		logger.Warn("entering PREPARE_VOTE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	logger.Info(">> PREPARE_VOTE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PrepareVoteStep
		c.newStep()
	}()
	if !c.hasPrepare() {
		logger.Error("PREPARE_VOTE step: Prepare message is nil")
		return
	}
	//logger.Debug("Prepare message is valid, decide to vote for it")
	vote := types.NewPrepareVote(height, types.GeneratePrepareVoteValueHash(c.stepInfo.block.Header.Hash), c.privateKey)
	if c.isLeader() {
		c.stepInfo.voteSet.AddPrepareVote(c.stepInfo.round, vote)
		ok := c.stepInfo.voteSet.CheckPrepareVoteIsComplete(c.stepInfo.round, c.state.Validators)
		if ok {
			c.scheduleStep(time.Millisecond*100, c.stepInfo.height, c.stepInfo.round, PreCommitStep)
		}
	} else {
		c.sendInternalMessage(MessageInfo{Msg: vote, NodeID: ""})
		c.stepInfo.prepare = nil
	}
}

func (c *Core) proposePreCommitMsg(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PreCommitStep) {
		logger.Warn("entering PRE_COMMIT step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	logger.Info("=> PRE_COMMIT step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PreCommitStep
		c.newStep()
	}()
	agg := c.stepInfo.voteSet.CreateThresholdSigForPrepareVote(round, c.cryptoBLS12)
	preCommit := types.NewPreCommit(agg, types.GeneratePreCommitValueHash(c.stepInfo.block.Header.Hash), c.publicKey.ToID(), height)
	c.sendInternalMessage(MessageInfo{Msg: preCommit, NodeID: ""})
}

func (c *Core) enterPreCommitVoteStep(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PreCommitVoteStep) {
		logger.Warn("entering PRE_COMMIT_VOTE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step), "should be", fmt.Sprintf("height:%d round:%d step:%s", height, round, PreCommitStep))
		return
	}
	logger.Info(">> PRE_COMMIT_VOTE step", "my_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PreCommitVoteStep
		c.newStep()
	}()
	if !c.hasPreCommit() {
		logger.Error("PRE_COMMIT_VOTE step: PreCommit message is nil")
		return
	}
	vote := types.NewPreCommitVote(height, types.GeneratePreCommitVoteValueHash(c.stepInfo.block.Header.Hash), c.privateKey)
	if c.isLeader() {
		c.stepInfo.voteSet.AddPreCommitVote(c.stepInfo.round, vote)
		ok := c.stepInfo.voteSet.CheckPreCommitVoteIsComplete(c.stepInfo.round, c.state.Validators)
		if ok {
			c.scheduleStep(time.Millisecond*100, c.stepInfo.height, c.stepInfo.round, CommitStep)
		}
	} else {
		c.sendInternalMessage(MessageInfo{Msg: vote, NodeID: ""})
		c.stepInfo.preCommit = nil
	}
}

func (c *Core) proposeCommitMsg(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= CommitStep) {
		logger.Warn("entering COMMIT step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	logger.Info("=> COMMIT step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = CommitStep
		c.newStep()
	}()
	agg := c.stepInfo.voteSet.CreateThresholdSigForPreCommitVote(round, c.cryptoBLS12)
	commit := types.NewCommit(agg, types.GenerateCommitValueHash(c.stepInfo.block.Header.Hash), c.publicKey.ToID(), height)
	c.sendInternalMessage(MessageInfo{Msg: commit, NodeID: ""})
}

func (c *Core) enterCommitVoteStep(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= CommitVoteStep) {
		logger.Warn("entering COMMIT_VOTE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	logger.Info(">> COMMIT_VOTE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = CommitVoteStep
		c.newStep()
	}()
	if !c.hasCommit() {
		logger.Error("COMMIT_VOTE step: Commit message is nil")
		return
	}
	vote := types.NewCommitVote(height, types.GenerateCommitVoteValueHash(c.stepInfo.block.Header.Hash), c.privateKey)
	if c.isLeader() {
		c.stepInfo.voteSet.AddCommitVote(c.stepInfo.round, vote)
		ok := c.stepInfo.voteSet.CheckCommitVoteIsComplete(c.stepInfo.round, c.state.Validators)
		if ok {
			c.scheduleStep(time.Millisecond*100, c.stepInfo.height, c.stepInfo.round, DecideStep)
		}
	} else {
		c.sendInternalMessage(MessageInfo{Msg: vote, NodeID: ""})
		c.stepInfo.commit = nil
	}
}

func (c *Core) proposeDecideMsg(height int64, round int16) {
	logger := c.Logger.New("height", height, "round", round)
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= DecideStep) {
		logger.Warn("entering DECIDE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	logger.Info("=> DECIDE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = DecideStep
		c.newStep()
	}()
	agg := c.stepInfo.voteSet.CreateThresholdSigForCommitVote(round, c.cryptoBLS12)
	decide := types.NewDecide(agg, types.GenerateDecideValueHash(c.stepInfo.block.Header.Hash), c.publicKey.ToID(), height)
	c.sendInternalMessage(MessageInfo{Msg: decide, NodeID: ""})
}

func (c *Core) applyBlock() {
	newState, err := c.blockExec.ApplyBlock(c.state, c.stepInfo.block)
	if err != nil {
		c.Logger.Error("failed to apply block", "err", err)
		return
	}
	c.state = newState
	c.stepInfo.previousBlock = c.stepInfo.block
	if !c.isNextLeader() {
		c.eventSwitch.FireEvent(events.EventNextView, c.nextView())
		c.scheduleRound0(c.stepInfo)
	}
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

func (c *Core) hasCommit() bool {
	return c.stepInfo.commit != nil
}

func (c *Core) newStep() {
	esi := c.stepInfo.EventStepInfo()
	c.eventSwitch.FireEvent(events.EventNewStep, &esi)
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
	//duration := time.Now().Sub(c.stepInfo.startTime)
	c.scheduledTicker.ScheduleTimeout(timeoutInfo{Duration: time.Second, Height: stepInfo.height, Round: 0, Step: NewHeightStep})
}

func (c *Core) sendInternalMessage(info MessageInfo) {
	select {
	case c.internalMsgQueue <- info:
	default:
		go func() { c.internalMsgQueue <- info }()
	}
}

func (c *Core) sendExternalMessage(info MessageInfo) {
	select {
	case c.externalMsgQueue <- info:
	default:
		go func() { c.externalMsgQueue <- info }()
	}
}

func (c *Core) nextView() *types.NextView {
	c.stepInfo.height += 1
	return &types.NextView{
		Type:   pbtypes.NextViewType,
		ID:     c.publicKey.ToID(),
		Height: c.stepInfo.height,
	}
}

func (c *Core) isLeader() bool {
	return c.state.Validators.GetLeader(c.stepInfo.height).ID == c.publicKey.ToID()
}

func (c *Core) isNextLeader() bool {
	return c.state.Validators.GetLeader(c.stepInfo.height+1).ID == c.publicKey.ToID()
}
