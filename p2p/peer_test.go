package p2p

import (
	"github.com/232425wxy/meta--/crypto"
	"sync"
)

type PeerMessage struct {
	PeerID  crypto.ID
	Bytes   []byte
	Counter int
}

type TestReactor struct {
	BaseReactor
	mu           sync.RWMutex
	channels     []*ChannelDescriptor
	msgCounter   int
	msgsReceived map[byte][]PeerMessage
}

func testReactor(channels []*ChannelDescriptor)

func createPeer(addr *NetAddress, config ConnectionConfig) (*Peer, error) {
	chdescs := []*ChannelDescriptor{
		{ID: 0x01, Priority: 1},
	}

}
