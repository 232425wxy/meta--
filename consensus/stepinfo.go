package consensus

import (
	"github.com/232425wxy/meta--/event"
	"github.com/232425wxy/meta--/types"
	"time"
)

type Step int8

const (
	NewViewStep Step = iota
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
	case NewViewStep:
		return "NEW_VIEW_STEP 1/9"
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
	height    int64
	round     int16
	step      Step
	startTime time.Time
	block     *types.Block
	blockHash []byte
	prepare   *types.Prepare
	preCommit *types.PreCommit
}

func (si *StepInfo) EventStepInfo() event.EventDataStep {
	return event.EventDataStep{
		Height: si.height,
		Round:  si.round,
		Step:   si.step.String(),
	}
}
