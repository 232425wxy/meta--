package consensus

import (
	"fmt"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/state"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
	"sync"
)

type Reactor struct {
	p2p.BaseReactor
	core     *Core
	waitSync bool
	mu       sync.RWMutex
}

func NewReactor(core *Core) *Reactor {
	r := &Reactor{core: core, waitSync: true}
	r.BaseReactor = *p2p.NewBaseReactor("Consensus")
	return r
}

func (r *Reactor) Start() error {
	if !r.waitSync {
		if err := r.core.Start(); err != nil {
			return err
		}
	}
	return r.BaseService.Start()
}

func (r *Reactor) InitPeer(peer *p2p.Peer) *p2p.Peer {
	stat := NewPeerState()
	peer.Set(types.PeerStateKey, stat)
	return peer
}

func (r *Reactor) AddPeer(peer *p2p.Peer) {
	go r.gossipRoutine(peer)
}

func (r *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  p2p.LeaderProposeChannel,
			Priority:            8,
			SendQueueCapacity:   100,
			RecvBufferCapacity:  1024 * 1024 * 10,
			RecvMessageCapacity: 1024 * 1024,
		},
		{
			ID:                  p2p.ReplicaStateChannel,
			Priority:            10,
			SendQueueCapacity:   100,
			RecvBufferCapacity:  1024 * 1024 * 10,
			RecvMessageCapacity: 1024 * 1024,
		},
		{
			ID:                  p2p.ReplicaVoteChannel,
			Priority:            5,
			SendQueueCapacity:   100,
			RecvBufferCapacity:  1024 * 1024 * 10,
			RecvMessageCapacity: 1024 * 1024,
		},
	}
}

func (r *Reactor) Receive(chID byte, src *p2p.Peer, bz []byte) {
	msg := MustDecode(bz)
	r.Logger.Debug("receive message", "peer", src.NodeID(), "message", msg)
	ps, ok := src.Get(types.PeerStateKey).(*PeerState)
	if !ok {
		panic(fmt.Sprintf("peer %v has no state", src.NodeID()))
	}
	info := MessageInfo{Msg: msg, NodeID: src.NodeID()}
	switch chID {
	case p2p.ReplicaStateChannel:
		r.core.sendExternalMessage(info)
		r.Logger.Info("receive next view message", "from", src.NodeID(), "height", ps.Height)
	case p2p.LeaderProposeChannel:
		r.core.sendExternalMessage(info)
		switch m := msg.(type) {
		case *types.Prepare:
			r.Logger.Info("receive Prepare message", "leader", src.NodeID(), "height", m.Height)
		case *types.PreCommit:
			r.Logger.Info("receive PreCommit message", "leader", src.NodeID(), "height", m.Height)
		case *types.Commit:
			r.Logger.Info("receive Commit message", "leader", src.NodeID(), "height", m.Height)
		case *types.Decide:
			r.Logger.Info("receive Decide message", "leader", src.NodeID(), "height", m.Height)
		}
	case p2p.ReplicaVoteChannel:
		r.core.sendExternalMessage(info)
		switch m := msg.(type) {
		case *types.PrepareVote:
			r.Logger.Info("receive PrepareVote message", "replica", src.NodeID(), "height", m.Vote.Height)
		case *types.PreCommitVote:
			r.Logger.Info("receive PreCommitVote message", "replica", src.NodeID(), "height", m.Vote.Height)
		case *types.CommitVote:
			r.Logger.Info("receive CommitVote message", "replica", src.NodeID(), "height", m.Vote.Height)
		}
	default:
		r.Logger.Error("unknown message channel", "channel", fmt.Sprintf("%x", chID))
	}

}

func (r *Reactor) SwitchToConsensus(stat *state.State) {
	if stat.LastBlockHeight <= r.core.state.LastBlockHeight {
		r.core.newStep()
	}

	r.mu.Lock()
	r.waitSync = false
	r.mu.Unlock()

	if err := r.core.Start(); err != nil {
		panic(err)
	}
}

func (r *Reactor) subscribeEvents() {
	subscriber := "consensus-reactor"
	if err := r.core.eventSwitch.AddListenerWithEvent(subscriber, events.EventNextView,
		func(data events.EventData) {
			r.sendNextViewToLeader(data.(*types.NextView))
		}); err != nil {
		r.Logger.Warn("failed to add listener for events", "err", err)
	}
}

func (r *Reactor) unsubscribeEvents() {
	subscriber := "consensus-reactor"
	r.core.eventSwitch.RemoveListener(subscriber)
}

func (r *Reactor) gossipRoutine(peer *p2p.Peer) {
	logger := r.Logger.New("peer", peer.NodeID())
	// 用PeerState来保证只会给节点发送一次主节点提出的共识消息
	ps := peer.Data.Get(types.PeerStateKey).(*PeerState)
	for {
		if r.core.isLeader() {
			switch {
			case r.core.stepInfo.prepare != nil && r.core.stepInfo.step == PrepareStep && !ps.HasPrepare(r.core.stepInfo.prepare):
				msg := MustEncode(r.core.stepInfo.prepare)
				if ok := peer.Send(p2p.LeaderProposeChannel, msg); ok {
					ps.SetPrepare(r.core.stepInfo.prepare)
					logger.Info("leader is me, send Prepare message", "to", peer.NodeID())
				}
			case r.core.stepInfo.preCommit != nil && r.core.stepInfo.step == PreCommitStep && !ps.HasPreCommit(r.core.stepInfo.preCommit):
				msg := MustEncode(r.core.stepInfo.preCommit)
				if ok := peer.Send(p2p.LeaderProposeChannel, msg); ok {
					ps.SetPreCommit(r.core.stepInfo.preCommit)
					logger.Info("leader is me, send PreCommit message", "to", peer.NodeID())
				}
			case r.core.stepInfo.commit != nil && r.core.stepInfo.step == CommitStep && !ps.HasCommit(r.core.stepInfo.commit):
				msg := MustEncode(r.core.stepInfo.commit)
				if ok := peer.Send(p2p.LeaderProposeChannel, msg); ok {
					ps.SetCommit(r.core.stepInfo.commit)
					logger.Info("leader is me, send Commit message", "to", peer.NodeID())
				}
			case r.core.stepInfo.decide != nil && r.core.stepInfo.step == DecideStep && !ps.HasDecide(r.core.stepInfo.decide):
				msg := MustEncode(r.core.stepInfo.decide)
				if ok := peer.Send(p2p.LeaderProposeChannel, msg); ok {
					ps.SetDecide(r.core.stepInfo.decide)
					logger.Info("leader is me, send Decide message", "to", peer.NodeID())
				}
			}
		}

		if peer.NodeID() == r.core.state.Validators.GetLeader(r.core.stepInfo.height).ID {
			// 只给leader节点发送投票信息
			select {
			case vote := <-r.core.prepareVotesQueue:
				msg := MustEncode(vote)
				peer.Send(p2p.ReplicaVoteChannel, msg)
				logger.Info("replica is me, send PrepareVote to leader", "me", r.core.publicKey.ToID(), "leader", peer.NodeID())
			case vote := <-r.core.preCommitVotesQueue:
				msg := MustEncode(vote)
				peer.Send(p2p.ReplicaVoteChannel, msg)
				logger.Info("replica is me, send PreCommitVote to leader", "me", r.core.publicKey.ToID(), "leader", peer.NodeID())
			case vote := <-r.core.commitVotesQueue:
				msg := MustEncode(vote)
				peer.Send(p2p.ReplicaVoteChannel, msg)
				logger.Info("replica is me, send CommitVote to leader", "me", r.core.publicKey.ToID(), "leader", peer.NodeID())
			}
		}
	}
}

func (r *Reactor) sendNextViewToLeader(view *types.NextView) {
	pb := view.ToProto()
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	r.Switch.SendToPeer(p2p.ReplicaStateChannel, r.core.state.Validators.GetLeader(r.core.stepInfo.height).ID, bz)
}
