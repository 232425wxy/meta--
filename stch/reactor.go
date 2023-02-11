package stch

import (
	"github.com/232425wxy/meta--/p2p"
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
			r.Logger.Error("收到了公钥", "received", r.ch.received, "from", msg.From)
			r.ch.received++
			r.ch._secret.Add(r.ch._secret, msg.A0)
			r.ch._secret.Mod(r.ch._secret, q)
			r.ch.secret.Add(r.ch.secret, msg.SK)
			r.ch.secret.Mod(r.ch.secret, q)
			if r.ch.received == 3 {
				go func() {
					time.Sleep(time.Second)
					r.ch._secret.Add(r.ch._secret, r.ch.fn.items[0])
					r.ch._secret.Mod(r.ch._secret, q)
					r.ch.secret.Add(r.ch.secret, r.ch.sk)
					r.ch.secret.Mod(r.ch.secret, q)
					r.Logger.Trace("恢复密钥", "完整私钥", r.ch._secret.String(), "计算私钥", r.ch.secret.String())
				}()
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
		SK:        r.ch.sk,
		A0:        r.ch.fn.items[0],
	}
	bz := MustEncode(pks)
	r.Switch.Broadcast(p2p.STCHChannel, bz)
}
