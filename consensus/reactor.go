package consensus

import (
	"fmt"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
	"sync"
)

type Reactor struct {
	p2p.BaseReactor
	core *Core
	mu   sync.RWMutex
}

func (r *Reactor) Receive(chID byte, src *p2p.Peer, bz []byte) {
	msg, err := MustDecode(bz)
	if err != nil {
		r.Switch.StopPeerForError(src, err)
		r.Logger.Error("failed to decode message", "peer", src.NodeID(), "err", err)
		return
	}
	r.Logger.Debug("receive message", "peer", src.NodeID(), "message", msg)
	ps, ok := src.Get(PeerStateKey).(*PeerState)
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
	for {
		if r.core.isLeader() {
			switch {
			case r.core.stepInfo.prepare != nil && r.core.stepInfo.step == PrepareStep:
				msg := MustEncode(r.core.stepInfo.prepare)
				peer.Send(p2p.LeaderProposeChannel, msg)
				logger.Info("leader is me, send Prepare message", "to", peer.NodeID())
			case r.core.stepInfo.preCommit != nil && r.core.stepInfo.step == PreCommitStep:
				msg := MustEncode(r.core.stepInfo.preCommit)
				peer.Send(p2p.LeaderProposeChannel, msg)
				logger.Info("leader is me, send PreCommit message", "to", peer.NodeID())
			}
		}

		if peer.NodeID() == r.core.state.Validators.GetLeader().ID {
			// 只给leader节点发送投票信息
			select {
			case vote := <-r.core.prepareVotesQueue:
				msg := MustEncode(vote)
				peer.Send(p2p.ReplicaVoteChannel, msg)
				logger.Info("replica is me, send PrepareVote to leader", "me", r.core.publicKey.ToID(), "leader", peer.NodeID())

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
	r.Switch.SendToPeer(p2p.ReplicaStateChannel, r.core.state.Validators.GetLeader().ID, bz)
}
