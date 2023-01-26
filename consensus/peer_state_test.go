package consensus

import (
	"github.com/232425wxy/meta--/common/cmap"
	"github.com/232425wxy/meta--/p2p"
	"testing"
)

func TestPeerData(t *testing.T) {
	peer := &p2p.Peer{Data: cmap.NewCap()}
	ps := &PeerState{
		Height: 12,
		Round:  2,
		Step:   PreCommitStep,
	}
	peer.Data.Set(PeerStateKey, ps)

	_ps := peer.Data.Get(PeerStateKey).(*PeerState)
	t.Log("ps:", ps)
	t.Log("_ps:", _ps)

	_ps.Step = DecideStep

	t.Log("ps:", ps)
}
