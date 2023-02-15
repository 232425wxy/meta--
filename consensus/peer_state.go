package consensus

import (
	"bytes"
	"fmt"
	"github.com/232425wxy/meta--/types"
)

type PeerState struct {
	Height    int64 `json:"height"`
	Round     int16 `json:"round"`
	Step      Step  `json:"step"`
	prepare   []byte
	preCommit []byte
	commit    []byte
	decide    []byte
}

func NewPeerState() *PeerState {
	return &PeerState{
		Height:    -1,
		Round:     -1,
		Step:      -1,
		prepare:   nil,
		preCommit: nil,
		commit:    nil,
		decide:    nil,
	}
}

func (ps *PeerState) SetHeight(height int64) {
	ps.Height = height
}

func (ps *PeerState) GetHeight() int64 {
	return ps.Height
}

func (ps *PeerState) SetRound(round int16) {
	ps.Round = round
}

func (ps *PeerState) SetStep(step Step) {
	ps.Step = step
}

func (ps *PeerState) SetPrepare(prepare *types.Prepare) {
	ps.prepare = prepare.Block.ChameleonHash.Hash
}

func (ps *PeerState) HasPrepare(prepare *types.Prepare) bool {
	return bytes.Equal(ps.prepare, prepare.Block.ChameleonHash.Hash)
}

func (ps *PeerState) SetPreCommit(preCommit *types.PreCommit) {
	ps.preCommit = preCommit.ValueHash[:]
}

func (ps *PeerState) HasPreCommit(preCommit *types.PreCommit) bool {
	return bytes.Equal(ps.preCommit, preCommit.ValueHash[:])
}

func (ps *PeerState) SetCommit(commit *types.Commit) {
	ps.commit = commit.ValueHash[:]
}

func (ps *PeerState) HasCommit(commit *types.Commit) bool {
	return bytes.Equal(ps.commit, commit.ValueHash[:])
}

func (ps *PeerState) SetDecide(decide *types.Decide) {
	if decide == nil {
		fmt.Println("decide是nil")
	}
	if len(decide.ValueHash[:]) == 0 {
		fmt.Println("decide.ValueHash是空的")
	}
	ps.decide = decide.ValueHash[:]
}

func (ps *PeerState) HasDecide(decide *types.Decide) bool {
	return bytes.Equal(ps.decide, decide.ValueHash[:])
}
