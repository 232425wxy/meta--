package consensus

import "C"
import (
	"bytes"
	"fmt"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	state2 "github.com/232425wxy/meta--/consensus/state"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"github.com/232425wxy/meta--/txspool"
	"github.com/232425wxy/meta--/types"
	"sync"
	"time"
)

const msgQueueSize = 10000

type Core struct {
	service.BaseService
	cfg                 *config.ConsensusConfig
	privateKey          *bls12.PrivateKey // 为共识消息签名的私钥
	publicKey           *bls12.PublicKey
	id                  crypto.ID             // 自己的节点ID
	blockExec           *state2.BlockExecutor // 创建区块和执行区块里的交易指令
	state               *state2.State
	txsPool             *txspool.TxsPool
	hasTxs              bool // hasTxs与交易池里的notifiedAvailable相互配合，保证主节点不会错失有交易数据到来的信号
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

func NewCore(cfg *config.ConsensusConfig, privateKey *bls12.PrivateKey, state *state2.State, blockExec *state2.BlockExecutor, txsPool *txspool.TxsPool, cryptoBLS12 *bls12.CryptoBLS12) *Core {
	core := &Core{
		BaseService:         *service.NewBaseService(nil, "Consensus_Core"),
		cfg:                 cfg,
		privateKey:          privateKey,
		publicKey:           privateKey.PublicKey(),
		id:                  privateKey.PublicKey().ToID(),
		blockExec:           blockExec,
		state:               state,
		txsPool:             txsPool,
		hasTxs:              false,
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
	return core
}

// for test
func (c *Core) testStatus() {
	for {
		time.Sleep(time.Second * 10)
		c.Logger.Info("观测共识状态", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		if c.isLeader() {
		}
	}
}

func (c *Core) Start() error {
	if err := c.scheduledTicker.Start(); err != nil {
		return err
	}
	go c.receiveRoutine()
	//go c.testStatus()

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
	for {
		select {
		case <-c.txsPool.TxsAvailable():
			if !c.hasTxs {
				go c.handleAvailableTxs()
				c.hasTxs = true
			}
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

	if c.isLeader() {
		switch c.stepInfo.step {
		case NewHeightStep, NewRoundStep:
			c.proposePrepareMsg(c.stepInfo.height, c.stepInfo.round)
		case DecideStep:
			// 当前高度的主节点收集齐其他节点发来的NextView消息后，本身的区块高度状态会自增1，
			// 凭借isLeader方法可以判定自己就是主节点，此外，收集齐其他节点发来的NextView消
			// 息后，主节点会进入1秒的超时等待状态，等待将状态从DecideStep切换为NewHeightStep，
			// 如果在这个阶段获得了需要打包交易数据的提醒，则超前进入打包区块的超时等待阶段。
			c.scheduleStep(time.Second, c.stepInfo.height, c.stepInfo.round, PrepareStep)
		}
	}
}

func (c *Core) handleScheduled(info timeoutInfo, stepInfo StepInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch info.Step {
	case NewHeightStep:
		c.stepInfo.Reset()
		if c.hasTxs && c.isLeader() {
			c.mu.Unlock()
			c.handleAvailableTxs()
			c.mu.Lock()
		}
	case PrepareStep:
		// 在从DecideStep状态转为NewHeightStep状态的过程中收到了交易池里有交易数据的信号，那么会
		// 重新设置一个超时时间，从DecideStep状态直接进入到PrepareStep，提出新的区块数据。
		c.stepInfo.Reset()
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
	if !c.isLeader() {
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
	c.Logger.Trace("receive a valid NextView message", "from", view.ID)
	c.stepInfo.AddNextView(view)
	if c.stepInfo.CheckCollectNextViewIsComplete(c.state.Validators) {
		c.Logger.Debug("receive enough NextView messages", "height", c.stepInfo.height)
		c.stepInfo.height += 1
		c.scheduleNewHeight(c.stepInfo)
	}
	return nil
}

func (c *Core) handlePrepare(prepare *types.Prepare) error {
	if prepare.Height != c.stepInfo.height {
		return nil
	}
	hash := make([]byte, len(prepare.Block.ChameleonHash.Hash))
	copy(hash[:], prepare.Block.ChameleonHash.Hash)
	ok := c.state.Validators.GetLeader(c.stepInfo.round).PublicKey.Verify(prepare.Signature, hash)
	if !ok {
		if c.isLeader() {
			panic(fmt.Sprintf("%s: \"why I created an invalid Prepare message?\" %d", c.state.Validators.GetLeader(c.stepInfo.round).PublicKey.ToID(), c.stepInfo.round))
		}
		return fmt.Errorf("leader %s sent an invalid prepare message to me", c.state.Validators.GetLeader(c.stepInfo.round).ID)
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
	valueHash := types.GeneratePrepareVoteValueHash(c.stepInfo.block.ChameleonHash.Hash)
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
	if preCommit.ID != c.state.Validators.GetLeader(c.stepInfo.round).ID {
		return fmt.Errorf("PreCommit message is not from leader %s at height %d", c.state.Validators.GetLeader(c.stepInfo.round).ID, c.stepInfo.height)
	}
	hash := types.GeneratePreCommitValueHash(c.stepInfo.block.ChameleonHash.Hash)
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
	valueHash := types.GeneratePreCommitVoteValueHash(c.stepInfo.block.ChameleonHash.Hash)
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
	if commit.ID != c.state.Validators.GetLeader(c.stepInfo.round).ID {
		return fmt.Errorf("Commit message is not from leader %s at height %d", c.state.Validators.GetLeader(c.stepInfo.round).ID, c.stepInfo.height)
	}
	hash := types.GenerateCommitValueHash(c.stepInfo.block.ChameleonHash.Hash)
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
	valueHash := types.GenerateCommitVoteValueHash(c.stepInfo.block.ChameleonHash.Hash)
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
	if decide.ID != c.state.Validators.GetLeader(c.stepInfo.round).ID {
		return fmt.Errorf("Decide message is not from leader %s at height %d", c.state.Validators.GetLeader(c.stepInfo.round).ID, c.stepInfo.height)
	}
	hash := types.GenerateDecideValueHash(c.stepInfo.block.ChameleonHash.Hash)
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
	if !c.isLeader() {
		c.stepInfo.step = DecideStep
		c.newStep()
	}
	c.applyBlock()
	return nil
}

func (c *Core) enterNewRound(height int64, round int16) {
	if c.stepInfo.height != height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step != NewHeightStep) {
		//logger.Warn("entering NEW ROUND step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	c.stepInfo.round = round + 1
	c.stepInfo.step = NewRoundStep
	//c.Logger.Info(">>> NEW ROUND step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	if round != 1 {
		c.stepInfo.block = nil
		c.stepInfo.prepare = nil
		c.stepInfo.preCommit = nil
		c.stepInfo.commit = nil
		c.stepInfo.decide = nil
	}
	//round += 1
	//c.proposePrepareMsg(height, round)
}

func (c *Core) proposePrepareMsg(height int64, round int16) {
	if c.stepInfo.height != height || round < c.stepInfo.round || (c.stepInfo.round == round && PrepareStep <= c.stepInfo.step) {
		c.Logger.Warn("entering PREPARE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}

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
			} else {
				if len(block.Body.Txs) == 0 {
					return
				}
			}
		}
		prepare := types.NewPrepare(height, block, c.publicKey.ToID(), c.privateKey)
		// 将Prepare消息发送到内部的消息通道里，这样在recvRoutine进程中可以捕获该消息，然后就会去处理该消息
		c.sendInternalMessage(MessageInfo{Msg: prepare, NodeID: ""})
		c.Logger.Info("=> PREPARE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	}
}

// enterPrepareVoteStep 主节点生成Prepare消息，并将其广播给其他节点，自己也保留这个Prepare消息，然后拥有Prepare消息的节点
// 为Prepare消息进行投票，主节点也会为自己生成的Prepare消息进行投票。
func (c *Core) enterPrepareVoteStep(height int64, round int16) {
	if c.stepInfo.height != height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PrepareVoteStep) {
		c.Logger.Warn("entering PREPARE_VOTE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	c.Logger.Info(">> PREPARE_VOTE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PrepareVoteStep
		c.newStep()
	}()
	if !c.hasPrepare() {
		c.Logger.Error("PREPARE_VOTE step: Prepare message is nil")
		return
	}
	//logger.Debug("Prepare message is valid, decide to vote for it")
	vote := types.NewPrepareVote(height, types.GeneratePrepareVoteValueHash(c.stepInfo.block.ChameleonHash.Hash), c.privateKey)
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
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PreCommitStep) {
		c.Logger.Warn("entering PRE_COMMIT step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	c.Logger.Info("=> PRE_COMMIT step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PreCommitStep
		c.newStep()
	}()
	agg := c.stepInfo.voteSet.CreateThresholdSigForPrepareVote(round, c.cryptoBLS12)
	preCommit := types.NewPreCommit(agg, types.GeneratePreCommitValueHash(c.stepInfo.block.ChameleonHash.Hash), c.publicKey.ToID(), height)
	c.sendInternalMessage(MessageInfo{Msg: preCommit, NodeID: ""})
}

func (c *Core) enterPreCommitVoteStep(height int64, round int16) {
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= PreCommitVoteStep) {
		c.Logger.Warn("entering PRE_COMMIT_VOTE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step), "should be", fmt.Sprintf("height:%d round:%d step:%s", height, round, PreCommitStep))
		return
	}
	c.Logger.Info(">> PRE_COMMIT_VOTE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = PreCommitVoteStep
		c.newStep()
	}()
	if !c.hasPreCommit() {
		c.Logger.Error("PRE_COMMIT_VOTE step: PreCommit message is nil")
		return
	}
	vote := types.NewPreCommitVote(height, types.GeneratePreCommitVoteValueHash(c.stepInfo.block.ChameleonHash.Hash), c.privateKey)
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
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= CommitStep) {
		c.Logger.Warn("entering COMMIT step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	c.Logger.Info("=> COMMIT step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = CommitStep
		c.newStep()
	}()
	agg := c.stepInfo.voteSet.CreateThresholdSigForPreCommitVote(round, c.cryptoBLS12)
	commit := types.NewCommit(agg, types.GenerateCommitValueHash(c.stepInfo.block.ChameleonHash.Hash), c.publicKey.ToID(), height)
	c.sendInternalMessage(MessageInfo{Msg: commit, NodeID: ""})
}

func (c *Core) enterCommitVoteStep(height int64, round int16) {
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= CommitVoteStep) {
		c.Logger.Warn("entering COMMIT_VOTE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	c.Logger.Info(">> COMMIT_VOTE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = CommitVoteStep
		c.newStep()
	}()
	if !c.hasCommit() {
		c.Logger.Error("COMMIT_VOTE step: Commit message is nil")
		return
	}
	vote := types.NewCommitVote(height, types.GenerateCommitVoteValueHash(c.stepInfo.block.ChameleonHash.Hash), c.privateKey)
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
	if height != c.stepInfo.height || c.stepInfo.round > round || (c.stepInfo.round == round && c.stepInfo.step >= DecideStep) {
		c.Logger.Warn("entering DECIDE step with invalid args", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
		return
	}
	c.Logger.Info("=> DECIDE step", "consensus_step", fmt.Sprintf("height:%d round:%d step:%s", c.stepInfo.height, c.stepInfo.round, c.stepInfo.step))
	defer func() {
		c.stepInfo.round = round
		c.stepInfo.step = DecideStep
		c.newStep()
	}()
	agg := c.stepInfo.voteSet.CreateThresholdSigForCommitVote(round, c.cryptoBLS12)
	decide := types.NewDecide(agg, types.GenerateDecideValueHash(c.stepInfo.block.ChameleonHash.Hash), c.publicKey.ToID(), height)
	c.sendInternalMessage(MessageInfo{Msg: decide, NodeID: ""})
}

func (c *Core) applyBlock() {
	newState, err := c.blockExec.ApplyBlock(c.state, c.stepInfo.block)
	c.hasTxs = false
	if err != nil {
		c.Logger.Error("failed to apply block", "err", err)
		return
	}
	c.state = newState
	c.stepInfo.previousBlock = c.stepInfo.block
	if !c.isLeader() {
		c.stepInfo.height += 1
		c.eventSwitch.FireEvent(events.EventNextView, c.nextView())
		c.scheduleNewHeight(c.stepInfo)
		//c.stepInfo.step = NewHeightStep
	}
}

func (c *Core) createBlock() *types.Block {
	switch {
	case c.stepInfo.height == c.state.InitialHeight:
		lastBlockHash := sha256.Sum([]byte("first block"))
		return c.blockExec.CreateBlock(c.stepInfo.height, c.state, c.publicKey.ToID(), lastBlockHash[:])
	case c.stepInfo.previousBlock != nil:
		return c.blockExec.CreateBlock(c.stepInfo.height, c.state, c.publicKey.ToID(), c.stepInfo.previousBlock.ChameleonHash.Hash)
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

func (c *Core) scheduleStep(duration time.Duration, height int64, round int16, step Step) {
	c.scheduledTicker.ScheduleTimeout(timeoutInfo{Duration: duration, Height: height, Round: round, Step: step})
}

// scheduleNewHeight 副本节点在确认过一个区块后，本地的区块高度会自增1，然后等待1秒中后进入下一个区块高度，如果自己是主导下一个
// 区块的主节点，那么在交易池里有交易数据的情况下，打包交易数据，提出新的区块，促使其他节点和自己进入到下一轮共识中；如果自己在下一
// 轮共识中依然是副本节点，那就只将自己的step更新为NewHeightStep。
func (c *Core) scheduleNewHeight(stepInfo *StepInfo) {
	//duration := time.Now().Sub(c.stepInfo.startTime)
	c.scheduledTicker.ScheduleTimeout(timeoutInfo{Duration: time.Second, Height: stepInfo.height, Round: 1, Step: NewHeightStep})
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
	return &types.NextView{
		Type:   pbtypes.NextViewType,
		ID:     c.publicKey.ToID(),
		Height: c.stepInfo.height,
	}
}

func (c *Core) isLeader() bool {
	return c.state.Validators.GetLeader(c.stepInfo.round).ID == c.publicKey.ToID()
}

func (c *Core) isNextLeader() bool {
	return c.state.Validators.GetLeader(c.stepInfo.round+1).ID == c.publicKey.ToID()
}
