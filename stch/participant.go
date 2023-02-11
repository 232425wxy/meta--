package stch

import (
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/p2p"
	"math/big"
)

type ParticipantSet struct {
	ps map[crypto.ID]*Participant
}

func NewParticipantSet() *ParticipantSet {
	return &ParticipantSet{ps: make(map[crypto.ID]*Participant)}
}

func (set *ParticipantSet) AddParticipant(participant *Participant) error {
	if _, ok := set.ps[participant.peer.NodeID()]; ok {
		return fmt.Errorf("%s already exists", participant.peer.NodeID())
	}
	set.ps[participant.peer.NodeID()] = participant
	return nil
}

type Participant struct {
	x        *big.Int
	fnX      *big.Int
	fnXForMe *big.Int
	pk       *big.Int // 节点的公钥
	peer     *p2p.Peer
}
