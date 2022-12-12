package p2p

import (
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/log"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
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

func testReactor(channels []*ChannelDescriptor) *TestReactor {
	tr := &TestReactor{
		channels:     channels,
		msgsReceived: make(map[byte][]PeerMessage),
	}
	tr.BaseReactor = *NewBaseReactor("TestReactor")
	logger := log.New("module", "test/reactor")
	logger.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	tr.SetLogger(logger)
	return tr
}

func createPeer(t *testing.T, addr *NetAddress, config ConnectionConfig) *Peer {
	chdescs := []*ChannelDescriptor{
		{ID: 0x01, Priority: 1},
	}
	reactorByCh := map[byte]Reactor{0x01: testReactor(chdescs)}
	privateKey, err := bls12.GeneratePrivateKey()
	assert.Nil(t, t)
}
