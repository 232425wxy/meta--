package consensus

import (
	"fmt"
	"github.com/232425wxy/meta--/events"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
	"sync"
)

const ConsensusChannel = byte(0x20)

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
	case ProposalChannel:

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

func (r *Reactor) gossipDataRoutine(peer *p2p.Peer) {

}

func (r *Reactor) sendNextViewToLeader(view *types.NextView) {
	pb := view.ToProto()
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	r.Switch.SendToPeer(ConsensusChannel, r.core.state.Validators.GetLeader().ID, bz)
}
