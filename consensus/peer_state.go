package consensus

import (
	"bytes"
	"github.com/232425wxy/meta--/types"
)

const PeerStateKey = "Consensus.Peer.State"

type PeerState struct {
	Height    int64 `json:"height"`
	Round     int16 `json:"round"`
	Step      Step  `json:"step"`
	prepare   []byte
	preCommit []byte
	commit    []byte
	decide    []byte
}

func (ps *PeerState) SetHeight(height int64) {
	ps.Height = height
}

func (ps *PeerState) SetRound(round int16) {
	ps.Round = round
}

func (ps *PeerState) SetStep(step Step) {
	ps.Step = step
}

func (ps *PeerState) SetPrepare(prepare *types.Prepare) {
	ps.prepare = prepare.Block.Header.Hash
}

func (ps *PeerState) HasPrepare(prepare *types.Prepare) bool {
	return bytes.Equal(ps.prepare, prepare.Block.Header.Hash)
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
	ps.decide = decide.ValueHash[:]
}

func (ps *PeerState) HasDecide(decide *types.Decide) bool {
	return bytes.Equal(ps.decide, decide.ValueHash[:])
}
