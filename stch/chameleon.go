package stch

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/merkle"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/proto/pbstch"
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

type redactBag struct {
	bag map[crypto.ID]*pbstch.SchnorrSig
}

func (bag *redactBag) add(peerID crypto.ID, ss *pbstch.SchnorrSig) {
	if bag.bag == nil {
		bag.bag = make(map[crypto.ID]*pbstch.SchnorrSig)
	}
	bag.bag[peerID] = ss
}

type Chameleon struct {
	id              crypto.ID
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
	alphaExpK       *big.Int
	redactTaskChan  chan *Task
	redactAvailable bool
	blockStore      *store.BlockStore
	redacts         map[string]struct{}
	bag             map[string]*redactBag
	verCh           chan *FinalVer
	verMap          map[crypto.ID]*FinalVer
	waitRedactBlock *types.Block
	mu              sync.Mutex
}

func NewChameleon(id crypto.ID, n int) *Chameleon {
	ch := &Chameleon{}
	ch.id = id
	ch.k, ch.x = GenerateKAndX()
	ch.fn = &polynomial{items: make(map[int]*big.Int)}
	ch.n = n
	ch.participants = NewParticipantSet()
	ch.generateFn(n)
	ch.fnX = ch.fn.calculate(ch.x, q)
	ch.hk = new(big.Int).SetInt64(1)
	ch.cid = new(big.Int).SetInt64(0)
	ch.redactTaskChan = make(chan *Task, 100)
	ch.redacts = make(map[string]struct{})
	ch.bag = make(map[string]*redactBag)
	ch.verCh = make(chan *FinalVer, 1)
	ch.verMap = make(map[crypto.ID]*FinalVer)
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
	ch.alphaExpK = new(big.Int).Exp(ch.alpha, ch.k, q)
}

func (ch *Chameleon) handleAlphaExpKAndHK(ah *AlphaExpKAndHK, peer *p2p.Peer) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	if ch.hk.Int64() != 1 {
		if ch.hk.Cmp(ah.HK) != 0 {
			return fmt.Errorf("peer %s generate different hk from mine", peer.NodeID())
		}
	}
	ch.participants.ps[peer.NodeID()].alphaExpK = new(big.Int).Set(ah.AlphaExpK)
	return nil
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
	h := new(big.Int).Mul(block.ChameleonHash.GSigma, new(big.Int).Exp(block.ChameleonHash.Alpha, sigma, q))
	h.Mod(h, q)
	block.ChameleonHash.Hash = h.Bytes()
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

func (ch *Chameleon) handleRedactTask(task *Task, myID crypto.ID) ([]byte, error) {
	block := ch.blockStore.LoadBlockByHeight(task.BlockHeight)
	redactBlock := block.Copy()
	old_msg := redactBlock.BlockDataHash()
	if task.TxIndex >= len(redactBlock.Body.Txs) {
		return nil, fmt.Errorf("you can only redact existed tx, origin_txs_num: %d, redact_tx_index: %d", len(redactBlock.Body.Txs), task.TxIndex)
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

	lss := &LeaderSchnorrSig{}
	e := new(big.Int).Sub(new(big.Int).SetBytes(old_msg), new(big.Int).SetBytes(new_msg))
	s := new(big.Int).Add(new(big.Int).Mul(ch.sk, e), ch.k)
	d := new(big.Int)
	alpha := new(big.Int).Set(redactBlock.ChameleonHash.Alpha)
	if s.Cmp(new(big.Int).SetInt64(0)) < 0 {
		inverseAlpha := calcInverseElem(alpha, q)
		_s := new(big.Int).Neg(s)
		d = d.Exp(inverseAlpha, _s, q)
		lss.Flag = true
	} else {
		d = d.Exp(alpha, s, q)
	}
	lss.S = s
	lss.D = d
	lss.BlockHeight = task.BlockHeight
	lss.TxIndex = task.TxIndex
	lss.NewTx = []byte(fmt.Sprintf("%x=%x", task.Key, task.Value))
	redactExist := redactHash(lss.BlockHeight, lss.TxIndex, lss.NewTx)
	if ch.bag[redactExist] == nil {
		ch.bag[redactExist] = new(redactBag)
	}
	ch.bag[redactExist].add(myID, lss.ToProto())
	bz := MustEncode(lss)
	return bz, nil
}

func (ch *Chameleon) verifyLeaderSchnorrSig(lss *LeaderSchnorrSig, peer *p2p.Peer, myID crypto.ID) ([]byte, error) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	redactExist := redactHash(lss.BlockHeight, lss.TxIndex, lss.NewTx)
	if len(ch.redacts) > 0 {
		if _, ok := ch.redacts[redactExist]; !ok {
			return nil, errors.New("redact task is not exist")
		}
	} else {
		ch.redacts[redactExist] = struct{}{}
	}
	block := ch.blockStore.LoadBlockByHeight(lss.BlockHeight)
	originBlockDataHash := block.BlockDataHash()

	redactBlock := block.Copy()
	redactBlock.Body.Txs[lss.TxIndex] = lss.NewTx
	redactBlockDataHash := redactBlock.BlockDataHash()

	e := new(big.Int).Sub(new(big.Int).SetBytes(redactBlockDataHash), new(big.Int).SetBytes(originBlockDataHash))
	_e := new(big.Int).Neg(e)

	if lss.Flag {
		lss.S.Neg(lss.S)
	}
	x_ := new(big.Int)
	if lss.S.Cmp(new(big.Int).SetInt64(0)) < 0 {
		_g := calcInverseElem(g, q)
		_s := new(big.Int).Neg(lss.S)
		x_.Exp(_g, _s, q)
	} else {
		x_.Exp(g, lss.S, q)
	}
	if e.Cmp(new(big.Int).SetInt64(0)) < 0 {
		_pk := calcInverseElem(ch.participants.ps[peer.NodeID()].pk, q)
		x_ = new(big.Int).Mul(x_, new(big.Int).Exp(_pk, _e, q))
	} else {
		x_ = new(big.Int).Mul(x_, new(big.Int).Exp(ch.participants.ps[peer.NodeID()].pk, e, q))
	}
	x_.Mod(x_, q)
	if x_.Cmp(ch.participants.ps[peer.NodeID()].x) == 0 {
		rss := &ReplicaSchnorrSig{}
		s := new(big.Int).Add(new(big.Int).Mul(ch.sk, _e), ch.k)
		d := new(big.Int)
		alpha := new(big.Int).Set(redactBlock.ChameleonHash.Alpha)
		if s.Cmp(new(big.Int).SetInt64(0)) < 0 {
			inverseAlpha := calcInverseElem(alpha, q)
			_s := new(big.Int).Neg(s)
			d = d.Exp(inverseAlpha, _s, q)
			rss.Flag = true
		} else {
			d = d.Exp(alpha, s, q)
		}
		rss.S = s
		rss.D = d
		rss.BlockHeight = lss.BlockHeight
		rss.TxIndex = lss.TxIndex
		rss.NewTx = lss.NewTx
		bz := MustEncode(rss)
		if ch.bag[redactExist] == nil {
			ch.bag[redactExist] = new(redactBag)
		}
		ch.bag[redactExist].add(peer.NodeID(), lss.ToProto())
		ch.bag[redactExist].add(myID, rss.ToProto())
		if len(ch.bag[redactExist].bag) == ch.n {
			return bz, ch.redact(redactExist, peer.NodeID())
		}
		return bz, nil
	} else {
		return nil, fmt.Errorf("leader %s sent wrong segment", peer.NodeID())
	}
}

func (ch *Chameleon) verifyReplicaSchnorrSig(rss *ReplicaSchnorrSig, peer *p2p.Peer) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	redactExist := redactHash(rss.BlockHeight, rss.TxIndex, rss.NewTx)
	if len(ch.redacts) > 0 {
		if _, ok := ch.redacts[redactExist]; !ok {
			return fmt.Errorf("peer %s may took wrong action when redact block", peer.NodeID())
		}
	} else {
		ch.redacts[redactExist] = struct{}{}
	}
	block := ch.blockStore.LoadBlockByHeight(rss.BlockHeight)
	originBlockDataHash := block.BlockDataHash()

	redactBlock := block.Copy()
	redactBlock.Body.Txs[rss.TxIndex] = rss.NewTx
	redactBlockDataHash := redactBlock.BlockDataHash()

	e := new(big.Int).Sub(new(big.Int).SetBytes(redactBlockDataHash), new(big.Int).SetBytes(originBlockDataHash))
	_e := new(big.Int).Neg(e)

	if rss.Flag {
		rss.S.Neg(rss.S)
	}
	x_ := new(big.Int)
	if rss.S.Cmp(new(big.Int).SetInt64(0)) < 0 {
		_g := calcInverseElem(g, q)
		_s := new(big.Int).Neg(rss.S)
		x_.Exp(_g, _s, q)
	} else {
		x_.Exp(g, rss.S, q)
	}
	if e.Cmp(new(big.Int).SetInt64(0)) < 0 {
		_pk := calcInverseElem(ch.participants.ps[peer.NodeID()].pk, q)
		x_ = new(big.Int).Mul(x_, new(big.Int).Exp(_pk, _e, q))
	} else {
		x_ = new(big.Int).Mul(x_, new(big.Int).Exp(ch.participants.ps[peer.NodeID()].pk, e, q))
	}
	x_.Mod(x_, q)
	if x_.Cmp(ch.participants.ps[peer.NodeID()].x) == 0 {
		if ch.bag[redactExist] == nil {
			ch.bag[redactExist] = new(redactBag)
		}
		ch.bag[redactExist].add(peer.NodeID(), rss.ToProto())
		if len(ch.bag[redactExist].bag) == ch.n {
			if err := ch.redact(redactExist, peer.NodeID()); err != nil {
				return err
			}
		}
		return nil
	} else {
		return fmt.Errorf("peer %s send wrong information", peer.NodeID())
	}
}

func (ch *Chameleon) redact(redactStr string, peerID crypto.ID) error {
	r := ch.bag[redactStr].bag[peerID]
	block := ch.blockStore.LoadBlockByHeight(r.BlockHeight)
	originBlockDataHash := block.BlockDataHash()
	block.Body.Txs[r.TxIndex] = r.Tx
	redactBlockDataHash := block.BlockDataHash()

	Alpha := new(big.Int).Set(ch.alphaExpK)
	for _, p := range ch.participants.ps {
		Alpha.Mul(Alpha, p.alphaExpK)
	}
	inverseAlpha := calcInverseElem(Alpha, q)
	e := new(big.Int).Sub(new(big.Int).SetBytes(originBlockDataHash), new(big.Int).SetBytes(redactBlockDataHash))
	_e := new(big.Int).Neg(e)

	c := new(big.Int).SetInt64(1)
	for _, seg := range ch.bag[redactStr].bag {
		di := new(big.Int).SetBytes(seg.D)
		c.Mul(c, di)
	}
	c.Mul(c, inverseAlpha)

	r1 := new(big.Int)
	if e.Cmp(new(big.Int).SetInt64(0)) < 0 {
		_alpha := calcInverseElem(block.ChameleonHash.Alpha, q)
		r1 = new(big.Int).Mul(block.ChameleonHash.GSigma, new(big.Int).Exp(_alpha, _e, q))
	} else {
		r1 = new(big.Int).Mul(block.ChameleonHash.GSigma, new(big.Int).Exp(block.ChameleonHash.Alpha, e, q))
	}
	r1.Mod(r1, q)
	block.ChameleonHash.GSigma.Set(r1)

	r2 := new(big.Int).Mul(block.ChameleonHash.HKSigma, c)
	r2.Mod(r2, q)
	block.ChameleonHash.HKSigma.Set(r2)

	rh := new(big.Int).Mul(block.ChameleonHash.GSigma, new(big.Int).Exp(block.ChameleonHash.Alpha, new(big.Int).SetBytes(redactBlockDataHash), q))
	rh.Mod(rh, q)
	if rh.Cmp(new(big.Int).SetBytes(block.ChameleonHash.Hash)) != 0 {
		return errors.New("redact failed")
	} else {
		ver := &FinalVer{
			Val:       new(big.Int).Exp(block.ChameleonHash.GSigma, ch.sk, q),
			RedactStr: redactStr,
			R2:        new(big.Int).Set(block.ChameleonHash.HKSigma),
		}
		ch.verMap[ch.id] = ver
		select {
		case ch.verCh <- ver:
		default:
			go func() { ch.verCh <- ver }()
		}
		ch.waitRedactBlock = block
		return nil
	}
}

func (ch *Chameleon) handleFinalVer(fv *FinalVer, peer *p2p.Peer) (bool, error) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.verMap[peer.NodeID()] = fv
	if len(ch.verMap) == ch.n {
		v := new(big.Int).SetInt64(1)
		for _, ver := range ch.verMap {
			v.Mul(v, ver.Val)
		}
		v.Mod(v, q)
		if v.Cmp(fv.R2) == 0 {
			if ch.waitRedactBlock != nil {
				ch.blockStore.SaveBlock(ch.waitRedactBlock)
				ch.redacts = make(map[string]struct{})
				ch.bag = make(map[string]*redactBag)
				ch.verMap = make(map[crypto.ID]*FinalVer)
			} else {
				return true, errors.New("no block to redact")
			}
		} else {
			return true, errors.New("wrong redact information")
		}
		ch.waitRedactBlock = nil
	}
	return false, nil
}
