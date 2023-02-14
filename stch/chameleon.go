package stch

import (
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/merkle"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/store"
	"github.com/232425wxy/meta--/types"
	"math/big"
	"sync"
)

type Task struct {
	BlockHeight int64
	TxIndex     int
	Key         []byte
	Value       []byte
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
	k               *big.Int
	x               *big.Int
	fn              *polynomial
	fnX             *big.Int
	sk              *big.Int // 节点自己的私钥
	pk              *big.Int // 节点自己的公钥
	n               int      // 分布式成员数量
	participants    *ParticipantSet
	hk              *big.Int // 变色龙哈希函数的公钥
	cid             *big.Int
	alpha           *big.Int
	redactTaskChan  chan *Task
	redactAvailable bool
	blockStore      *store.BlockStore
	mu              sync.Mutex
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
	ch.redactTaskChan = make(chan *Task, 1)
	return ch
}

func (ch *Chameleon) SetBlockStore(bs *store.BlockStore) {
	ch.blockStore = bs
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
		x.Mul(x, new(big.Int).Mul(neg, inverse))
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
	blockDataHash := block.BlockDataHash()
	if block.ChameleonHash == nil {
		block.ChameleonHash = &types.ChameleonHash{}
	}
	sigma := new(big.Int).SetBytes(blockDataHash)
	block.ChameleonHash.GSigma = new(big.Int).Exp(g, sigma, q)
	block.ChameleonHash.HKSigma = new(big.Int).Exp(ch.hk, sigma, q)
	block.ChameleonHash.Alpha = ch.alpha
	block.ChameleonHash.Hash = new(big.Int).Mul(block.ChameleonHash.GSigma, new(big.Int).Exp(block.ChameleonHash.Alpha, sigma, q)).Bytes()
}

func (ch *Chameleon) AppendRedactTask(task *Task) {
	select {
	case ch.redactTaskChan <- task:
		ch.redactAvailable = true
	default:
		go func() {
			ch.redactTaskChan <- task
			ch.redactAvailable = true
		}()
	}
}

func (ch *Chameleon) handleRedactTask(task *Task) ([]byte, error) {
	block := ch.blockStore.LoadBlockByHeight(task.BlockHeight)
	redactBlock := block.Copy()
	old_msg := redactBlock.BlockDataHash()
	if task.TxIndex >= len(redactBlock.Body.Txs)+1 {
		return nil, fmt.Errorf("you can only redact existed txs or append tx, origin_txs_num: %d, redact_tx_index: %d", len(redactBlock.Body.Txs), task.TxIndex)
	}
	if task.TxIndex == len(redactBlock.Body.Txs) {
		redactBlock.Body.Txs = append(redactBlock.Body.Txs, []byte(fmt.Sprintf("%x=%x", task.Key, task.Value)))
	}
	if task.TxIndex < len(redactBlock.Body.Txs) {
		tx := []byte(fmt.Sprintf("%x=%x", task.Key, task.Value))
		redactBlock.Body.Txs[task.TxIndex] = tx
	}
	_txs := make([][]byte, len(redactBlock.Body.Txs))
	for i, tx := range redactBlock.Body.Txs {
		_txs[i] = tx
	}
	redactBlock.Body.RootHash = merkle.ComputeMerkleRoot(_txs)
	new_msg := redactBlock.BlockDataHash()

	ss := &LeaderSchnorrSig{}
	e := new(big.Int).Sub(new(big.Int).SetBytes(old_msg), new(big.Int).SetBytes(new_msg))
	s := new(big.Int).Add(new(big.Int).Mul(ch.sk, e), ch.k)
	d := new(big.Int)
	alpha := new(big.Int).Set(redactBlock.ChameleonHash.Alpha)
	if s.Cmp(new(big.Int).SetInt64(0)) < 0 {
		inverseAlpha := calcInverseElem(alpha, q)
		_s := new(big.Int).Neg(s)
		d = d.Exp(inverseAlpha, _s, q)
		ss.Flag = true
	} else {
		d = d.Exp(alpha, s, q)
	}
	ss.S = s
	ss.D = d
	ss.BlockHeight = task.BlockHeight
	ss.TxIndex = task.TxIndex
	ss.NewTx = []byte(fmt.Sprintf("%x=%x", task.Key, task.Value))
	bz := MustEncode(ss)
	return bz, nil
}

func (ch *Chameleon) verifyLeaderSchnorrSig(sig *LeaderSchnorrSig, peer *p2p.Peer) {
	block := ch.blockStore.LoadBlockByHeight(sig.BlockHeight)
	originBlockDataHash := block.BlockDataHash()

	redactBlock := block.Copy()
	redactBlock.Body.Txs[sig.TxIndex] = sig.NewTx
	redactBlockDataHash := redactBlock.BlockDataHash()

	e := new(big.Int).Sub(new(big.Int).SetBytes(redactBlockDataHash), new(big.Int).SetBytes(originBlockDataHash))
	_e := new(big.Int).Neg(e)
	if sig.Flag {
		sig.S.Neg(sig.S)
	}
	x_ := new(big.Int)
	if sig.S.Cmp(new(big.Int).SetInt64(0)) < 0 {
		_g := calcInverseElem(g, q)
		_s := new(big.Int).Neg(sig.S)
		x_ = new(big.Int).Exp(_g, _s, q)
	} else {
		x_ = new(big.Int).Exp(g, sig.S, q)
	}
	if e.Cmp(new(big.Int).SetInt64(0)) < 0 {
		_pk := calcInverseElem(ch.participants.ps[peer.NodeID()].pk, q)
		x_ = new(big.Int).Mul(x_, new(big.Int).Exp(_pk, _e, q))
	} else {
		x_ = new(big.Int).Mul(x_, new(big.Int).Exp(ch.participants.ps[peer.NodeID()].pk, e, q))
	}
	x_.Mod(x_, q)
	if x_.Cmp(ch.participants.ps[peer.NodeID()].x) != 0 {
		fmt.Println("验证失败！", x_)
	} else {
		fmt.Println("验证成功！", x_)
	}
}
