package consensus

import (
	"github.com/232425wxy/meta--/crypto"
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
	roundVoteSets map[int16]RoundVoteSet
}
