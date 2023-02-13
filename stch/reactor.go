package stch

import (
	"fmt"
	"github.com/232425wxy/meta--/p2p"
	"math/big"
	"time"
)

type Reactor struct {
	p2p.BaseReactor
	ch        *Chameleon
	receivedX int
}

func NewReactor(ch *Chameleon) *Reactor {
	return &Reactor{
		BaseReactor: *p2p.NewBaseReactor("STCH"),
		ch:          ch,
		receivedX:   0,
	}
}

func (r *Reactor) Start() error {
	go r.processRedactTaskRoutine()
	return r.BaseService.Start()
}

func (r *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  p2p.STCHChannel,
			Priority:            10,
			SendQueueCapacity:   100,
			RecvBufferCapacity:  1024 * 1024 * 10,
			RecvMessageCapacity: 1024 * 1024,
		},
	}
}

func (r *Reactor) AddPeer(peer *p2p.Peer) {
	r.sendXToPeer(peer)
}

func (r *Reactor) InitPeer(peer *p2p.Peer) *p2p.Peer {
	return peer
}

func (r *Reactor) Receive(chID byte, src *p2p.Peer, bz []byte) {

	switch chID {
	case p2p.STCHChannel:
		msg := MustDecode(bz)
		switch msg := msg.(type) {
		case *IdentityX:
			if err := r.ch.handleIdentityX(src, msg); err != nil {
				r.Logger.Error("failed to handle IdentityX message", "err", err)
				return
			}
			fnX := r.ch.calculateFnXForPeer(msg, r.Switch.NodeInfo().NodeID, src.NodeID())
			r.sendFnXToPeer(fnX, src)
		case *FnX:
			if ok := r.ch.handleFnX(src, msg); ok {
				r.ch.calculateSK(g, q)
				r.broadcastPKToPeer()
			}
		case *PublicKeySeg:
			if ok := r.ch.handlePublicKeySeg(src, msg); ok {
				// 收集齐了其他节点的公钥
				if r.ch.pk != nil {
					r.ch.calculateHKAndCID(q)
					r.Logger.Error("计算变色龙公钥", "hk", r.ch.hk.String(), "cid", r.ch.cid.String(), "alpha", r.ch.alpha.String())
				} else {
					// 自己的公钥还没制作出来的情况下，需要等待自己的公钥制作出来后再生成变色龙公钥
					go func() {
						for {
							if r.ch.pk != nil {
								r.ch.calculateHKAndCID(q)
								r.Logger.Error("计算变色龙公钥", "hk", r.ch.hk.String(), "cid", r.ch.cid.String(), "alpha", r.ch.alpha.String())
								return
							}
							time.Sleep(time.Millisecond * 10)
						}
					}()
				}
			}
		case *SchnorrSig:
			r.Logger.Error("来任务了，需要修改区块链了")
		}
	}
}

func (r *Reactor) sendXToPeer(peer *p2p.Peer) {
	identityX := &IdentityX{
		X:  r.ch.GetX(),
		ID: r.Switch.NodeInfo().ID(),
	}
	bz := MustEncode(identityX)
	peer.Send(p2p.STCHChannel, bz)
}

func (r *Reactor) sendFnXToPeer(fnX *FnX, peer *p2p.Peer) {
	bz := MustEncode(fnX)
	peer.Send(p2p.STCHChannel, bz)
}

func (r *Reactor) broadcastPKToPeer() {
	pks := &PublicKeySeg{
		From:      r.Switch.NodeInfo().ID(),
		PublicKey: r.ch.pk,
	}
	bz := MustEncode(pks)
	r.Switch.Broadcast(p2p.STCHChannel, bz)
}

func (r *Reactor) Chameleon() *Chameleon {
	return r.ch
}

func (r *Reactor) processRedactTaskRoutine() {
	for {
		if r.ch.redactAvailable {
			select {
			case task := <-r.ch.redactTaskChan:
				r.handleRedactTask(task)
			}
		}
	}
}

func (r *Reactor) handleRedactTask(task *Task) {
	block := task.Block
	fmt.Println(task, block)
	if task.TxIndex >= len(block.Body.Txs)+1 {
		r.Logger.Error("you can only redact existed txs or append tx", "origin_txs_num", len(block.Body.Txs), "redact_tx_index", task.TxIndex)
		return
	}
	if task.TxIndex == len(block.Body.Txs) {
		block.Body.Txs = append(block.Body.Txs, []byte(fmt.Sprintf("%v=%v", task.Key, task.Value)))
	}
	if task.TxIndex < len(block.Body.Txs) {
		tx := []byte(fmt.Sprintf("%v=%v", task.Key, task.Value))
		block.Body.Txs[task.TxIndex] = tx
	}
	old_msg := block.Header.BlockDataHash
	new_msg := block.BlockDataHash()

	e := new(big.Int).Sub(new(big.Int).SetBytes(old_msg), new(big.Int).SetBytes(new_msg))
	s := new(big.Int).Mul(r.ch.sk, e)
	s.Add(s, r.ch.k)
	d := new(big.Int)
	alpha := new(big.Int).Set(block.ChameleonHash.Alpha)
	if s.Cmp(new(big.Int).SetInt64(0)) < 0 {
		inverseAlpha := calcInverseElem(alpha, q)
		s.Neg(s)
		d = d.Exp(inverseAlpha, s, q)
	} else {
		d = d.Exp(alpha, s, q)
	}
	ss := &SchnorrSig{
		S: s,
		D: d,
	}
	bz := MustEncode(ss)
	r.Switch.Broadcast(p2p.STCHChannel, bz)
}
