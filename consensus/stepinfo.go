package consensus

import (
	"github.com/232425wxy/meta--/types"
)

type Step int8

const (
	NewViewStep Step = iota
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
		return "NEW_VIEW_STEP 1/8"
	case PrepareStep:
		return "PREPARE_STEP 2/8"
	case PrepareVoteStep:
		return "PREPARE_VOTE_STEP 3/8"
	case PreCommitStep:
		return "PRE_COMMIT_STEP 4/8"
	case PreCommitVoteStep:
		return "PRE_COMMIT_VOTE_STEP 5/8"
	case CommitStep:
		return "COMMIT_STEP 6/8"
	case CommitVoteStep:
		return "COMMIT_VOTE_STEP 7/8"
	case DecideStep:
		return "DECIDE_STEP 8/8"
	default:
		panic("unknown step")
	}
}

type StepInfo struct {
	height    int64
	block     *types.Block
	blockHash []byte
	prepare   *types.Prepare
	preCommit *types.PreCommit
}
