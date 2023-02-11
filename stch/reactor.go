package stch

import (
	"github.com/232425wxy/meta--/p2p"
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
			fnX := r.ch.CalculateFnXForPeer(msg, r.Switch.NodeInfo().NodeID, src.NodeID())
			r.sendFnXToPeer(fnX, src)
		case *FnX:
			if ok := r.ch.handleFnX(src, msg); ok {
				//r.Logger.Error("广播")
				r.ch.calculateSK(g, q)
				r.broadcastPKToPeer()
			}
		case *PublicKeySeg:
			if ok := r.ch.handlePublicKeySeg(src, msg); ok {
				// 收集齐了其他节点的公钥，可以计算变色龙哈希函数的公钥了

			}

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
