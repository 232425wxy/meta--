package consensus

import (
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

func (r *Reactor) Receive(chID byte, src *p2p.Peer, msg []byte) {

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
