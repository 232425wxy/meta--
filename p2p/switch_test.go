package p2p

import (
	"fmt"
	"github.com/232425wxy/meta--/common/hexbytes"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

type node struct {
	info *NodeInfo
	key  *NodeKey
}

func testNodeInfo2(privateKey *bls12.PrivateKey) *NodeInfo {
	id := privateKey.PublicKey().ToID()
	channels := hexbytes.HexBytes{0x01}
	listen := fmt.Sprintf("127.0.0.1:%d", getFreePort())
	fmt.Printf("节点%s监听地址：%s\n", id, listen)
	return &NodeInfo{
		NodeID:     id,
		ListenAddr: listen,
		ChainID:    "100",
		Channels:   channels,
		RPCAddress: fmt.Sprintf("127.0.0.1:%d", getFreePort()),
		TxIndex:    "on",
	}
}

func createTransport2(n *node) *Transport {
	cfg := config.DefaultP2PConfig()
	cfg.PongTimeout = 2 * time.Second
	cfg.PingInterval = 4 * time.Second
	transport := NewTransport(n.info, n.key, cfg)
	return transport
}

func create2Nodes(t *testing.T) (*node, *node) {
	privateKeyA, errA := bls12.GeneratePrivateKey()
	assert.Nil(t, errA)
	privateKeyB, errB := bls12.GeneratePrivateKey()
	assert.Nil(t, errB)
	nodeA := testNodeInfo2(privateKeyA)
	nodeB := testNodeInfo2(privateKeyB)
	keyA := testNodeKey(privateKeyA)
	keyB := testNodeKey(privateKeyB)
	return &node{info: nodeA, key: keyA}, &node{info: nodeB, key: keyB}
}

func create2Transports(t *testing.T) (*Transport, *Transport) {
	nodeA, nodeB := create2Nodes(t)
	transportA := createTransport2(nodeA)
	transportB := createTransport2(nodeB)
	addr1 := testNetAddress(nodeA.info.ID(), nodeA.info.ListenAddr)
	addr2 := testNetAddress(nodeB.info.ID(), nodeB.info.ListenAddr)
	err := transportA.Listen(addr1)
	assert.Nil(t, err)
	err = transportB.Listen(addr2)
	assert.Nil(t, err)
	return transportA, transportB
}

func create2Switches(t *testing.T) (*Switch, *Switch) {
	transportA, transportB := create2Transports(t)
	sw1 := NewSwitch(transportA, metricsA)
	sw2 := NewSwitch(transportB, metricsB)
	logger1 := log.New("Switch", "1")
	logger1.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	logger2 := log.New("Switch", "2")
	logger2.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	sw1.SetLogger(logger1)
	sw2.SetLogger(logger2)
	return sw1, sw2
}

func TestSwitchBasic(t *testing.T) {
	log.PrintOrigins(true)
	sw1, sw2 := create2Switches(t)
	addr1 := sw1.NetAddress()

	err := sw1.Start()
	assert.Nil(t, err)
	err = sw2.Start()
	assert.Nil(t, err)

	pass := make(chan struct{})
	go func() {
		for {
			if sw1.peers.Size() > 0 {
				close(pass)
				return
			}
		}
	}()
	err = sw2.DialPeerWithAddress(addr1)
	assert.Nil(t, err)
	<-pass

	assert.Equal(t, 1, sw1.peers.Size())
	assert.Equal(t, 1, sw2.peers.Size())

	assert.True(t, sw2.peers.HasPeer(sw1.NodeInfo().NodeID))
	assert.Nil(t, sw1.Stop())
	assert.Nil(t, sw2.Stop())

	assert.Equal(t, 0, sw1.peers.Size())
	assert.Equal(t, 0, sw2.peers.Size())
}

func TestSwitch_Broadcast(t *testing.T) {
	sw1, sw2 := create2Switches(t)
	reactor1 := &TestReactor{
		BaseReactor: *NewBaseReactor("Switch-1"),
		msgReceived: make(map[byte][]PeerMsg),
		counter:     0,
	}
	sw1.AddReactor("test reactor", reactor1)
	reactor2 := &TestReactor{
		BaseReactor: *NewBaseReactor("Switch-2"),
		msgReceived: make(map[byte][]PeerMsg),
		counter:     0,
	}
	sw2.AddReactor("test reactor", reactor2)
	assert.Nil(t, sw1.Start())
	assert.Nil(t, sw2.Start())

	connected := make(chan struct{})

	go func() {
		assert.Nil(t, sw1.DialPeerWithAddress(sw2.NetAddress()))
	}()

	go func() {
		for {
			if sw2.peers.HasPeer(sw1.NodeInfo().ID()) {
				close(connected)
				return
			}
		}
	}()

	<-connected

	sw1.Broadcast(0x01, []byte("hello, world"))

	for {
		if reactor2.counter == 1 {
			break
		}
	}

	assert.Equal(t, reactor2.msgReceived[0x01][0].peerID, sw1.NodeInfo().NodeID)
	assert.Equal(t, reactor2.msgReceived[0x01][0].msg, []byte("hello, world"))

	assert.Nil(t, sw1.Stop())
	assert.Nil(t, sw2.Stop())
}

func TestSwitch_SetAddrBook(t *testing.T) {
	sw1, sw2 := create2Switches(t)
	addrBook1 := NewAddrBook("addrbook1.json")
	addrBook2 := NewAddrBook("addrbook2.json")
	sw1.SetAddrBook(addrBook1)
	sw2.SetAddrBook(addrBook2)
	assert.Nil(t, sw1.Start())
	assert.Nil(t, sw2.Start())

	connected := make(chan struct{})

	go func() {
		assert.Nil(t, sw1.DialPeerWithAddress(sw2.NetAddress()))
	}()

	go func() {
		for {
			if sw2.peers.HasPeer(sw1.NodeInfo().ID()) {
				close(connected)
				return
			}
		}
	}()

	<-connected

	time.Sleep(time.Second * 10)

	assert.Nil(t, sw1.Stop())
	assert.Nil(t, sw2.Stop())
}
