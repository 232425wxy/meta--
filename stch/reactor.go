package stch

import (
	"math/big"
	"time"

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
	go r.processRedactTaskRoutine()
	go r.waitForFinalVer()
	go r.processFormerRSS()
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
				r.Logger.Error("Failed to handle IdentityX message", "err", err)
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
					r.brodacastAlphaExpKAndHK()
					r.Logger.Info("Distributed chameleon hash function initialization complete", "hk", r.ch.hk.String()[:10], "cid", r.ch.cid.String()[:10], "alpha", r.ch.alpha.String())
				} else {
					// 自己的公钥还没制作出来的情况下，需要等待自己的公钥制作出来后再生成变色龙公钥
					go func() {
						for {
							if r.ch.pk != nil {
								r.ch.calculateHKAndCID(q)
								r.brodacastAlphaExpKAndHK()
								r.Logger.Error("Distributed chameleon hash function initialization complete", "hk", r.ch.hk.String()[:10], "cid", r.ch.cid.String()[:10], "alpha", r.ch.alpha.String())
								return
							}
							time.Sleep(time.Millisecond * 10)
						}
					}()
				}
			}
		case *AlphaExpKAndHK:
			if err := r.ch.handleAlphaExpKAndHK(msg, src); err != nil {
				r.Logger.Error("Failed to handle AlphaExpKAndHK message", "err", err)
			}
		case *LeaderSchnorrSig:
			r.Logger.Debug("Receive new redact mission from leader", "leader", src.NodeID())
			data, err := r.ch.verifyLeaderSchnorrSig(msg, src, r.Switch.NodeInfo().ID())
			if len(data) > 0 && err == nil {
				r.Switch.Broadcast(p2p.STCHChannel, data)
			} else if err != nil {
				r.Logger.Error("The private key slice information sent by the leader is incorrect", "leader", src.NodeID())
			}
		case *ReplicaSchnorrSig:
			r.Logger.Debug("Receive segment of threshold key", "from", src.NodeID())
			if err := r.ch.verifyReplicaSchnorrSig(msg, src.NodeID()); err != nil {
				r.Logger.Error("Failed to handle replica schnorr signature", "err", err)
			}
		case *RandomVerification:
			r.Logger.Debug("Receive new randomness of new block", "from", src.NodeID())
			err := r.ch.handleRandomVerification(msg, src.NodeID())
			if err != nil {
				r.Logger.Error("Failed to handle verification of new randomness", "err", err)
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

func (r *Reactor) Chameleon() *Chameleon {
	return r.ch
}

func (r *Reactor) brodacastAlphaExpKAndHK() {
	ah := &AlphaExpKAndHK{
		AlphaExpK: new(big.Int).Set(r.ch.alphaExpK),
		HK:        new(big.Int).Set(r.ch.hk),
	}
	bz := MustEncode(ah)
	r.Switch.Broadcast(p2p.STCHChannel, bz)
}

func (r *Reactor) processRedactTaskRoutine() {
	for {
		if r.ch.redactAvailable {
			select {
			case task := <-r.ch.redactTaskChan:
				r.ch.redactAvailable = false
				r.Logger.Debug("A new redact mission arrives")
				data, err := r.ch.handleRedactTask(task, r.Switch.NodeInfo().ID())
				if err != nil {
					r.Logger.Error("failed to handle generateNewRandomness task", "err", err)
				} else {
					r.Switch.Broadcast(p2p.STCHChannel, data)
				}
			}
		}
	}
}

func (r *Reactor) waitForFinalVer() {
	for {
		select {
		case rv := <-r.ch.redactSteps.randomChan:
			bz := MustEncode(rv)
			r.Switch.Broadcast(p2p.STCHChannel, bz)
		}
	}
}

func (r *Reactor) processFormerRSS() {
	for {
		if len(r.ch.redactSteps.redactMission) > 0 {
			select {
			case rss := <-r.ch.redactSteps.rssChan:
				if err := r.ch.verifyReplicaSchnorrSig(rss.rss, rss.id); err != nil {
					r.Logger.Error("Failed to handle replica schnorr signature", "err", err)
				}
			}
		}
	}
}
