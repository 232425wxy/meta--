package stch

import (
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

type Participant struct {
	x         *big.Int
	fnX       *big.Int
	fnXForMe  *big.Int
	pk        *big.Int // 节点的公钥
	alphaExpK *big.Int
	peer      *p2p.Peer
}
