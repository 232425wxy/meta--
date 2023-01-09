package consensus

import (
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/types"
	"time"
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
	height        int64
	round         int16
	step          Step
	startTime     time.Time
	block         *types.Block
	previousBlock *types.Block
	prepare       *types.Prepare
	preCommit     *types.PreCommit
	heightVoteSet *HeightVoteSet
	validators    *types.ValidatorSet
}

func (si *StepInfo) EventStepInfo() events.EventDataStep {
	return events.EventDataStep{
		Height: si.height,
		Round:  si.round,
		Step:   si.step.String(),
	}
}

func (si *StepInfo) EventNewRound() events.EventDataNewRound {
	return events.EventDataNewRound{
		Height:   si.height,
		Round:    si.round,
		Step:     si.step.String(),
		LeaderID: si.validators.GetLeader().ID,
	}
}

type RoundVoteSet struct {
	PrepareVoteSet   map[crypto.ID]*types.PrepareVote
	PreCommitVoteSet map[crypto.ID]*types.PreCommitVote
	CommitVoteSet    map[crypto.ID]*types.CommitVote
}

type HeightVoteSet struct {
	height        int64
	round         int16
	validators    *types.ValidatorSet
	roundVoteSets map[int16]*RoundVoteSet // round -> RoundVoteSet
}

func (hvs *HeightVoteSet) AddPrepareVote(round int16, vote *types.PrepareVote) {
	if hvs.roundVoteSets == nil {
		hvs.roundVoteSets = make(map[int16]*RoundVoteSet)
	}
	roundVoteSet := hvs.roundVoteSets[round]
	if roundVoteSet == nil {
		roundVoteSet = &RoundVoteSet{
			PrepareVoteSet:   make(map[crypto.ID]*types.PrepareVote),
			PreCommitVoteSet: make(map[crypto.ID]*types.PreCommitVote),
			CommitVoteSet:    make(map[crypto.ID]*types.CommitVote),
		}
	}
	roundVoteSet.PrepareVoteSet[vote.Vote.Signature.Signer()] = vote
}

func (hvs *HeightVoteSet) CheckPrepareVoteIsComplete(round int16, cb *bls12.CryptoBLS12) (bool, *bls12.AggregateSignature) {
	roundVoteSet := hvs.roundVoteSets[round]
	var hasVoePower int64 = 0
	sigs := make([]*bls12.Signature, 0)
	for id, vote := range roundVoteSet.PrepareVoteSet {
		validator := hvs.validators.GetValidatorByID(id)
		hasVoePower += validator.VotingPower
		sigs = append(sigs, vote.Vote.Signature)
	}
	if hasVoePower >= hvs.validators.Major23() {
		aggregateSignature, err := cb.CreateThresholdSignature(sigs)
		if err != nil {
			return false, nil
		}
		return true, aggregateSignature
	}
	return false, nil
}
