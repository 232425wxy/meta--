package consensus

import (
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/types"
)

type Step int8

const (
	NewHeightStep Step = iota
	NewRoundStep
	PrepareStep
	PrepareVoteStep
	PreCommitStep
	PreCommitVoteStep
	CommitStep
	CommitVoteStep
	DecideStep
	ConsensusTimeout
)

func (s Step) String() string {
	switch s {
	case NewHeightStep:
		return "NEW_HEIGHT_STEP 1/9"
	case NewRoundStep:
		return "NEW_ROUND_STEP 2/9"
	case PrepareStep:
		return "PREPARE_STEP 3/9"
	case PrepareVoteStep:
		return "PREPARE_VOTE_STEP 4/9"
	case PreCommitStep:
		return "PRE_COMMIT_STEP 5/9"
	case PreCommitVoteStep:
		return "PRE_COMMIT_VOTE_STEP 6/9"
	case CommitStep:
		return "COMMIT_STEP 7/9"
	case CommitVoteStep:
		return "COMMIT_VOTE_STEP 8/9"
	case DecideStep:
		return "DECIDE_STEP 9/9"
	default:
		panic("unknown step")
	}
}

type StepInfo struct {
	height          int64
	round           int16
	step            Step
	block           *types.Block
	previousBlock   *types.Block
	prepare         chan *types.Prepare
	preCommit       chan *types.PreCommit
	commit          chan *types.Commit
	decide          chan *types.Decide
	voteSet         *VoteSet
	collectNextView map[crypto.ID]*types.NextView
}

func NewStepInfo() *StepInfo {
	return &StepInfo{
		round:           1,
		voteSet:         NewVoteSet(),
		collectNextView: make(map[crypto.ID]*types.NextView),
		prepare:         make(chan *types.Prepare, 1),
		preCommit:       make(chan *types.PreCommit, 1),
		commit:          make(chan *types.Commit, 1),
		decide:          make(chan *types.Decide, 1),
	}
}

func (si *StepInfo) Reset() {
	si.round = 1
	si.step = NewHeightStep
	si.block = nil
	si.prepare = make(chan *types.Prepare, 1)
	si.preCommit = make(chan *types.PreCommit, 1)
	si.commit = make(chan *types.Commit, 1)
	si.decide = make(chan *types.Decide, 1)
	si.voteSet.Reset()
	si.collectNextView = make(map[crypto.ID]*types.NextView)
}

func (si *StepInfo) EventStepInfo() events.EventDataNewStep {
	return events.EventDataNewStep{
		Height: si.height,
		Round:  si.round,
		Step:   int8(si.step),
	}
}

func (si *StepInfo) AddNextView(view *types.NextView) {
	collect := si.collectNextView
	if collect == nil {
		collect = make(map[crypto.ID]*types.NextView)
	}
	collect[view.ID] = view
	si.collectNextView = collect
}

func (si *StepInfo) CheckCollectNextViewIsComplete(validators *types.ValidatorSet) bool {
	var hasPower int64 = 0
	for id := range si.collectNextView {
		validator := validators.GetValidatorByID(id)
		hasPower += validator.VotingPower
	}
	if hasPower >= validators.PowerMajor23() {
		return true
	}
	return false
}

type RoundVoteSet struct {
	PrepareVoteSet   map[crypto.ID]*types.PrepareVote
	PreCommitVoteSet map[crypto.ID]*types.PreCommitVote
	CommitVoteSet    map[crypto.ID]*types.CommitVote
}

type VoteSet struct {
	roundVoteSets map[int16]*RoundVoteSet // round -> RoundVoteSet
}

func NewVoteSet() *VoteSet {
	return &VoteSet{
		roundVoteSets: make(map[int16]*RoundVoteSet),
	}
}

func (vs *VoteSet) Reset() {
	vs.roundVoteSets = make(map[int16]*RoundVoteSet)
}

func (vs *VoteSet) AddPrepareVote(round int16, vote *types.PrepareVote) {
	if vs.roundVoteSets == nil {
		vs.roundVoteSets = make(map[int16]*RoundVoteSet)
	}
	roundVoteSet := vs.roundVoteSets[round]
	if roundVoteSet == nil {
		roundVoteSet = &RoundVoteSet{
			PrepareVoteSet:   make(map[crypto.ID]*types.PrepareVote),
			PreCommitVoteSet: make(map[crypto.ID]*types.PreCommitVote),
			CommitVoteSet:    make(map[crypto.ID]*types.CommitVote),
		}
	}
	roundVoteSet.PrepareVoteSet[vote.Vote.Signature.Signer()] = vote
	vs.roundVoteSets[round] = roundVoteSet
}

func (vs *VoteSet) AddPreCommitVote(round int16, vote *types.PreCommitVote) {
	if vs.roundVoteSets == nil {
		vs.roundVoteSets = make(map[int16]*RoundVoteSet)
	}
	roundVoteSet := vs.roundVoteSets[round]
	if roundVoteSet == nil {
		roundVoteSet = &RoundVoteSet{
			PrepareVoteSet:   make(map[crypto.ID]*types.PrepareVote),
			PreCommitVoteSet: make(map[crypto.ID]*types.PreCommitVote),
			CommitVoteSet:    make(map[crypto.ID]*types.CommitVote),
		}
	}
	roundVoteSet.PreCommitVoteSet[vote.Vote.Signature.Signer()] = vote
	vs.roundVoteSets[round] = roundVoteSet
}

func (vs *VoteSet) AddCommitVote(round int16, vote *types.CommitVote) {
	if vs.roundVoteSets == nil {
		vs.roundVoteSets = make(map[int16]*RoundVoteSet)
	}
	roundVoteSet := vs.roundVoteSets[round]
	if roundVoteSet == nil {
		roundVoteSet = &RoundVoteSet{
			PrepareVoteSet:   make(map[crypto.ID]*types.PrepareVote),
			PreCommitVoteSet: make(map[crypto.ID]*types.PreCommitVote),
			CommitVoteSet:    make(map[crypto.ID]*types.CommitVote),
		}
	}
	roundVoteSet.CommitVoteSet[vote.Vote.Signature.Signer()] = vote
	vs.roundVoteSets[round] = roundVoteSet
}

func (vs *VoteSet) CheckPrepareVoteIsComplete(round int16, validators *types.ValidatorSet) bool {
	roundVoteSet := vs.roundVoteSets[round]
	var hasVotePower int64 = 0
	for id := range roundVoteSet.PrepareVoteSet {
		validator := validators.GetValidatorByID(id)
		hasVotePower += validator.VotingPower
	}
	if hasVotePower >= validators.PowerMajor23() {
		return true
	}
	return false
}

func (vs *VoteSet) CheckPreCommitVoteIsComplete(round int16, validators *types.ValidatorSet) bool {
	roundVoteSet := vs.roundVoteSets[round]
	var hasVotePower int64 = 0
	for id := range roundVoteSet.PreCommitVoteSet {
		validator := validators.GetValidatorByID(id)
		hasVotePower += validator.VotingPower
	}
	if hasVotePower >= validators.PowerMajor23() {
		return true
	}
	return false
}

func (vs *VoteSet) CheckCommitVoteIsComplete(round int16, validators *types.ValidatorSet) bool {
	roundVoteSet := vs.roundVoteSets[round]
	var hasVotePower int64 = 0
	for id := range roundVoteSet.CommitVoteSet {
		validator := validators.GetValidatorByID(id)
		hasVotePower += validator.VotingPower
	}
	if hasVotePower >= validators.PowerMajor23() {
		return true
	}
	return false
}

func (vs *VoteSet) CreateThresholdSigForPrepareVote(round int16, cb *bls12.CryptoBLS12) *bls12.AggregateSignature {
	roundVoteSet := vs.roundVoteSets[round]
	sigs := make([]*bls12.Signature, 0)
	for _, vote := range roundVoteSet.PrepareVoteSet {
		sigs = append(sigs, vote.Vote.Signature)
	}
	agg, err := cb.CreateThresholdSignature(sigs)
	if err != nil {
		return nil
	}
	return agg
}

func (vs *VoteSet) CreateThresholdSigForPreCommitVote(round int16, cb *bls12.CryptoBLS12) *bls12.AggregateSignature {
	roundVoteSet := vs.roundVoteSets[round]
	sigs := make([]*bls12.Signature, 0)
	for _, vote := range roundVoteSet.PreCommitVoteSet {
		sigs = append(sigs, vote.Vote.Signature)
	}
	agg, err := cb.CreateThresholdSignature(sigs)
	if err != nil {
		return nil
	}
	return agg
}

func (vs *VoteSet) CreateThresholdSigForCommitVote(round int16, cb *bls12.CryptoBLS12) *bls12.AggregateSignature {
	roundVoteSet := vs.roundVoteSets[round]
	sigs := make([]*bls12.Signature, 0)
	for _, vote := range roundVoteSet.CommitVoteSet {
		sigs = append(sigs, vote.Vote.Signature)
	}
	agg, err := cb.CreateThresholdSignature(sigs)
	if err != nil {
		return nil
	}
	return agg
}
