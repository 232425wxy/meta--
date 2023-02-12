package stch

import (
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/types"
	"math/big"
	"sync"
)

type Task struct {
	Height int64
}

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
	sk           *big.Int // 节点自己的私钥
	pk           *big.Int // 节点自己的公钥
	n            int      // 分布式成员数量
	participants *ParticipantSet
	hk           *big.Int // 变色龙哈希函数的公钥
	cid          *big.Int
	alpha        *big.Int
	mu           sync.Mutex
}

func NewChameleon(n int) *Chameleon {
	ch := &Chameleon{}
	ch.k, ch.x = GenerateKAndX()
	ch.fn = &polynomial{items: make(map[int]*big.Int)}
	ch.n = n
	ch.participants = NewParticipantSet()
	ch.generateFn(n)
	ch.fnX = ch.fn.calculate(ch.x, q)
	ch.hk = new(big.Int).SetInt64(1)
	ch.cid = new(big.Int).SetInt64(0)
	return ch
}

func (ch *Chameleon) generateFn(num int) {
	ch.mu.Lock()
	ch.mu.Unlock()
	for i := 0; i < num; i++ {
		ch.fn.items[i] = GeneratePolynomialItem()
	}
}

func (ch *Chameleon) GetX() *big.Int {
	return ch.x
}

func (ch *Chameleon) calculateFnXForPeer(identity *IdentityX, myID crypto.ID, peerID crypto.ID) *FnX {
	res := &FnX{}
	res.Data = ch.fn.calculate(identity.X, q)
	res.From = myID
	res.X = ch.x
	ch.mu.Lock()
	ch.participants.ps[peerID].fnX = res.Data
	ch.mu.Unlock()
	return res
}

func (ch *Chameleon) handleIdentityX(peer *p2p.Peer, identityX *IdentityX) error {
	if peer.NodeID() != identityX.ID {
		return fmt.Errorf("identity mismatch, from %s, but identity is %s", peer.NodeID(), identityX.ID)
	}
	ch.mu.Lock()
	defer ch.mu.Unlock()
	participant := &Participant{
		x:    identityX.X,
		fnX:  nil,
		pk:   nil,
		peer: peer,
	}
	ch.participants.ps[peer.NodeID()] = participant
	return nil
}

func (ch *Chameleon) handleFnX(peer *p2p.Peer, fnX *FnX) bool {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	if peer.NodeID() != fnX.From {
		return false
	}
	if _, ok := ch.participants.ps[fnX.From]; !ok {
		participant := &Participant{
			x:        fnX.X,
			fnX:      ch.fn.calculate(fnX.X, q),
			fnXForMe: nil,
			pk:       nil,
			peer:     peer,
		}
		ch.participants.ps[fnX.From] = participant
	}
	ch.participants.ps[fnX.From].fnXForMe = fnX.Data
	receivedFull := true
	for _, participant := range ch.participants.ps {
		if participant.fnXForMe == nil {
			receivedFull = false
		}
	}
	return receivedFull && len(ch.participants.ps) == ch.n-1
}

func (ch *Chameleon) calculateSK(g, q *big.Int) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	fn := new(big.Int).Set(ch.fnX)
	x := new(big.Int).SetInt64(1)
	for _, participant := range ch.participants.ps {
		fn.Add(fn, participant.fnXForMe)
		fn.Mod(fn, q)
		neg := new(big.Int).Neg(participant.x)
		diff := new(big.Int).Sub(ch.x, participant.x)
		inverse := calcInverseElem(diff, q)
		d := new(big.Int).Mul(neg, inverse)
		x.Mul(x, d)
	}
	ch.sk = new(big.Int).Mul(fn, x)
	ch.sk.Mod(ch.sk, q)
	ch.pk = new(big.Int).Exp(g, ch.sk, q)
}

func (ch *Chameleon) handlePublicKeySeg(peer *p2p.Peer, key *PublicKeySeg) bool {
	if peer.NodeID() != key.From {
		panic(fmt.Sprintf("identity mismatch, peer is %s, but key is from %s", peer.NodeID(), key.From))
	}
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.participants.ps[key.From].pk = key.PublicKey
	receivedFull := true
	for _, participant := range ch.participants.ps {
		if participant.pk == nil {
			receivedFull = false
		}
	}
	return receivedFull && len(ch.participants.ps) == ch.n-1
}

func (ch *Chameleon) calculateHKAndCID(q *big.Int) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	for _, participant := range ch.participants.ps {
		ch.hk.Mul(ch.hk, participant.pk)
		ch.hk.Mod(ch.hk, q)

		ch.cid.Add(ch.cid, participant.x)
		ch.cid.Mod(ch.cid, q)
	}

	ch.hk.Mul(ch.hk, ch.pk)
	ch.hk.Mod(ch.hk, q)

	ch.cid.Add(ch.cid, ch.x)
	ch.cid.Mod(ch.cid, q)

	hashFn := sha256.New()
	hashFn.Write(ch.cid.Bytes())
	hashFn.Write(ch.hk.Bytes())
	h := hashFn.Sum(nil)
	ch.alpha = new(big.Int).SetBytes(h)
}

func (ch *Chameleon) Hash(block *types.Block) {
	blockDataHash := block.Hash()
	if block.ChameleonHash == nil {
		block.ChameleonHash = &types.ChameleonHash{}
	}
	sigma := new(big.Int).SetBytes(blockDataHash)
	block.ChameleonHash.GSigma = new(big.Int).Exp(g, sigma, q)
	fmt.Println(block.ChameleonHash.HKSigma)
	block.ChameleonHash.HKSigma = new(big.Int).Exp(ch.hk, sigma, q)
	block.ChameleonHash.Alpha = ch.alpha
	block.ChameleonHash.Hash = new(big.Int).Mul(block.ChameleonHash.GSigma, new(big.Int).Exp(block.ChameleonHash.Alpha, sigma, q)).Bytes()
}
