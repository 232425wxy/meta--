package stch

import (
	"fmt"
	"github.com/232425wxy/meta--/p2p"
	"math/big"
)

type polynomial struct {
	items map[int]*big.Int
}

// calculate 计算：fn(x) mod q
func (p *polynomial) calculate(x, q *big.Int) *big.Int {
	res := new(big.Int).SetInt64(0)
	for order, item := range p.items {
		e := new(big.Int).Exp(x, new(big.Int).SetInt64(int64(order)), q)
		e.Mul(e, item)
		res.Add(res, e)
	}
	return res.Mod(res, q)
}

type Chameleon struct {
	k            *big.Int
	x            *big.Int
	fn           *polynomial
	fnX          *big.Int
	n            int // 分布式成员数量
	participants *ParticipantSet
}

func NewChameleon(n int) *Chameleon {
	ch := &Chameleon{}
	ch.k, ch.x = GenerateKAndX()
	ch.fn = &polynomial{items: make(map[int]*big.Int)}
	ch.n = n
	ch.participants = NewParticipantSet()
	ch.GenerateFn(n)
	ch.fnX = ch.fn.calculate(ch.x, q)
	return ch
}

func (ch *Chameleon) GenerateFn(num int) {
	for i := 0; i < num; i++ {
		ch.fn.items[i] = GeneratePolynomialItem()
	}
}

func (ch *Chameleon) GetX() *big.Int {
	return ch.x
}

func (ch *Chameleon) handleIdentityX(peer *p2p.Peer, identityX *IdentityX) error {
	if peer.NodeID() != identityX.ID {
		return fmt.Errorf("identity mismatch, from %s, but identity is %s", peer.NodeID(), identityX.ID)
	}
	participant := &Participant{
		x:    identityX.X,
		fnX:  nil,
		pk:   nil,
		peer: peer,
	}
	return ch.participants.AddParticipant(participant)
}
