package stch

import "github.com/232425wxy/meta--/p2p"

type Reactor struct {
	p2p.BaseReactor
	ch *Chameleon
}

func NewReactor(ch *Chameleon) *Reactor {
	return &Reactor{
		BaseReactor: *p2p.NewBaseReactor("STCH"),
		ch:          ch,
	}
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

func (r *Reactor) Receive(chID byte, src *p2p.Peer, msg []byte) {
	switch chID {
	case p2p.STCHChannel:
		r.Logger.Error("收到了节点的变色龙信息", "from", src.NodeID(), "content", msg)
	}
}

func (r *Reactor) sendXToPeer(peer *p2p.Peer) {
	x := r.ch.GetX().Bytes()
	peer.Send(p2p.STCHChannel, x)
}
