package consensus

import (
	"fmt"
	"github.com/232425wxy/meta--/consensus/state"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/types"
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
	r.subscribeEvents()
	if !r.waitSync {
		if err := r.core.Start(); err != nil {
			return err
		}
	}
	go r.gossipRoutine()
	return r.BaseService.Start()
}

func (r *Reactor) InitPeer(peer *p2p.Peer) *p2p.Peer {
	stat := NewPeerState()
	peer.Set(types.PeerStateKey, stat)
	return peer
}

func (r *Reactor) AddPeer(peer *p2p.Peer) {

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
			ID:                  p2p.ReplicaNextViewChannel,
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
		{
			ID:                  p2p.ReplicaStateChannel,
			Priority:            3,
			SendQueueCapacity:   100,
			RecvBufferCapacity:  1024 * 1024,
			RecvMessageCapacity: 1024 * 1024,
		},
	}
}

func (r *Reactor) Receive(chID byte, src *p2p.Peer, bz []byte) {
	ps, ok := src.Get(types.PeerStateKey).(*PeerState)
	if !ok {
		panic(fmt.Sprintf("peer %v has no state", src.NodeID()))
	}

	switch chID {
	case p2p.ReplicaNextViewChannel, p2p.LeaderProposeChannel, p2p.ReplicaVoteChannel:
		msg := MustDecode(bz)
		info := MessageInfo{Msg: msg, NodeID: src.NodeID()}
		r.core.sendExternalMessage(info)
	case p2p.ReplicaStateChannel:
		msg := events.MustDecode(bz)
		switch msg := msg.(type) {
		case *events.EventDataNewStep:
			ps.SetHeight(msg.Height)
			ps.SetRound(msg.Round)
			ps.SetStep(Step(msg.Step))
			//r.Logger.Trace("收到了其他节点的状态信息", "状态", msg)
		}
	default:
		r.Logger.Error("unknown message channel", "channel", fmt.Sprintf("%x", chID))
	}

}

func (r *Reactor) SwitchToConsensus(stat *state.State) {
	//if stat.LastBlockHeight <= r.core.state.LastBlockHeight {
	//	r.core.newStep()
	//}

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
	if err := r.core.eventSwitch.AddListenerWithEvent(subscriber, events.EventNewStep,
		func(data events.EventData) {
			r.broadcastNewStep(data.(*events.EventDataNewStep))
		}); err != nil {
		r.Logger.Warn("failed to add listener for events", "err", err)
	}
}

func (r *Reactor) unsubscribeEvents() {
	subscriber := "consensus-reactor"
	r.core.eventSwitch.RemoveListener(subscriber)
}

func (r *Reactor) gossipRoutine() {
	logger := r.Logger.New()

	for {
		if r.core.isLeader() {
			select {
			case prepare := <-r.core.stepInfo.prepare:
				for _, p := range r.Switch.Peers().Peers() {
					// 用PeerState来保证只会给节点发送一次主节点提出的共识消息
					ps := p.Data.Get(types.PeerStateKey).(*PeerState)
					if (r.core.stepInfo.step == PrepareStep || r.core.stepInfo.step == PrepareVoteStep) && !ps.HasPrepare(prepare) {
						msg := MustEncode(prepare)
						if ok := p.Send(p2p.LeaderProposeChannel, msg); ok {
							ps.SetPrepare(prepare)
							//logger.Info("leader is me, send Prepare message", "to", peer.NodeID())
						} else {
							logger.Error("failed to send Prepare message", "to", p.NodeID())
							return
						}
					}
				}

			case preCommit := <-r.core.stepInfo.preCommit:
				for _, p := range r.Switch.Peers().Peers() {
					// 用PeerState来保证只会给节点发送一次主节点提出的共识消息
					ps := p.Data.Get(types.PeerStateKey).(*PeerState)
					if (r.core.stepInfo.step == PreCommitStep || r.core.stepInfo.step == PreCommitVoteStep) && !ps.HasPreCommit(preCommit) {
						msg := MustEncode(preCommit)
						if ok := p.Send(p2p.LeaderProposeChannel, msg); ok {
							ps.SetPreCommit(preCommit)
							//logger.Info("leader is me, send PreCommit message", "to", peer.NodeID())
						} else {
							logger.Error("failed to send PreCommit message", "to", p.NodeID())
							return
						}
					}
				}

			case commit := <-r.core.stepInfo.commit:
				for _, p := range r.Switch.Peers().Peers() {
					// 用PeerState来保证只会给节点发送一次主节点提出的共识消息
					ps := p.Data.Get(types.PeerStateKey).(*PeerState)
					if (r.core.stepInfo.step == CommitStep || r.core.stepInfo.step == CommitVoteStep) && !ps.HasCommit(commit) {
						msg := MustEncode(commit)
						if ok := p.Send(p2p.LeaderProposeChannel, msg); ok {
							ps.SetCommit(commit)
							//logger.Info("leader is me, send Commit message", "to", peer.NodeID())
						} else {
							logger.Error("failed to send Commit message", "to", p.NodeID())
							return
						}
					}
				}

			case decide := <-r.core.stepInfo.decide:
				for _, p := range r.Switch.Peers().Peers() {
					// 用PeerState来保证只会给节点发送一次主节点提出的共识消息
					ps := p.Data.Get(types.PeerStateKey).(*PeerState)
					if r.core.stepInfo.step == DecideStep && !ps.HasDecide(decide) {
						msg := MustEncode(decide)
						if ok := p.Send(p2p.LeaderProposeChannel, msg); ok {
							ps.SetDecide(decide)
							//logger.Trace("leader is me, send Decide message", "to", p.NodeID())
						} else {
							logger.Error("failed to send Decide message", "to", p.NodeID())
							return
						}
					}
				}

			default:

			}
		}

		// 只给leader节点发送投票信息
		select {
		case vote := <-r.core.prepareVotesQueue:
			msg := MustEncode(vote)
			_ = r.Switch.SendToPeer(p2p.ReplicaVoteChannel, r.core.state.Validators.GetLeader(r.core.stepInfo.round).ID, msg)
			//logger.Info("send PrepareVote to leader", "me", r.core.publicKey.ToID(), "leader", peer.NodeID())
		case vote := <-r.core.preCommitVotesQueue:
			msg := MustEncode(vote)
			_ = r.Switch.SendToPeer(p2p.ReplicaVoteChannel, r.core.state.Validators.GetLeader(r.core.stepInfo.round).ID, msg)
			//logger.Info("send PreCommitVote to leader", "me", r.core.publicKey.ToID(), "leader", peer.NodeID())
		case vote := <-r.core.commitVotesQueue:
			msg := MustEncode(vote)
			_ = r.Switch.SendToPeer(p2p.ReplicaVoteChannel, r.core.state.Validators.GetLeader(r.core.stepInfo.round).ID, msg)
			//logger.Info("send CommitVote to leader", "me", r.core.publicKey.ToID(), "leader", peer.NodeID())
		default:

		}
	}
}

func (r *Reactor) sendNextViewToLeader(view *types.NextView) {
	bz := MustEncode(view)
	_ = r.Switch.SendToPeer(p2p.ReplicaNextViewChannel, r.core.state.Validators.GetLeader(r.core.stepInfo.round).ID, bz)
}

func (r *Reactor) broadcastNewStep(step *events.EventDataNewStep) {
	bz := events.MustEncode(step)
	r.Switch.Broadcast(p2p.ReplicaStateChannel, bz)
}

func (r *Reactor) State() *state.State {
	return r.core.state
}
