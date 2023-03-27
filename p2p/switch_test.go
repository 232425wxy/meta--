package p2p

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/232425wxy/meta--/common/hexbytes"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/log"
	"github.com/stretchr/testify/assert"
)

type node struct {
	info    *NodeInfo
	key     *NodeKey
	address *NetAddress
}

func testNodeInfo2(privateKey *bls12.PrivateKey) *NodeInfo {
	id := privateKey.PublicKey().ToID()
	channels := hexbytes.HexBytes{0x01}
	listen := fmt.Sprintf("127.0.0.1:%d", getFreePort())
	fmt.Printf("节点%s监听地址：%s\n", id, listen)
	return &NodeInfo{
		NodeID:     id,
		ListenAddr: listen,
		Channels:   channels,
		RPCAddress: fmt.Sprintf("127.0.0.1:%d", getFreePort()),
		TxIndex:    "on",
	}
}

func createTransport2(n *node) *Transport {
	cfg := config.DefaultP2PConfig()
	cfg.PongTimeout = 2 * time.Second
	cfg.PingInterval = 4 * time.Second
	transport := NewTransport(n.address, n.info, n.key, cfg)
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
	addressA := testNetAddress(nodeA.NodeID, nodeA.ListenAddr)
	addressB := testNetAddress(nodeB.NodeID, nodeB.ListenAddr)
	return &node{info: nodeA, key: keyA, address: addressA}, &node{info: nodeB, key: keyB, address: addressB}
}

func create2Transports(t *testing.T) (*Transport, *Transport) {
	nodeA, nodeB := create2Nodes(t)
	transportA := createTransport2(nodeA)
	transportB := createTransport2(nodeB)
	err := transportA.Listen()
	assert.Nil(t, err)
	err = transportB.Listen()
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
	sw1.SetAddrBook(NewAddrBook("addrbook1.json"))
	sw2.SetAddrBook(NewAddrBook("addrbook2.json"))
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

	sw2.Broadcast(0x01, []byte("hello, world"))

	for {
		if reactor1.counter == 1 {
			break
		}
	}

	assert.Equal(t, reactor1.msgReceived[0x01][0].peerID, sw2.NodeInfo().NodeID)
	assert.Equal(t, reactor1.msgReceived[0x01][0].msg, []byte("hello, world"))

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

func Test4Nodes(t *testing.T) {
	log.PrintOrigins(true)
	sw1, sw2 := create2Switches(t)
	sw3, sw4 := create2Switches(t)

	sw1.SetAddrBook(NewAddrBook("addrbook1.json"))
	sw2.SetAddrBook(NewAddrBook("addrbook2.json"))
	sw3.SetAddrBook(NewAddrBook("addrbook3.json"))
	sw4.SetAddrBook(NewAddrBook("addrbook4.json"))

	reactor1 := &TestReactor{
		BaseReactor: *NewBaseReactor("Switch-1"),
		msgReceived: make(map[byte][]PeerMsg),
		counter:     0,
	}
	reactor2 := &TestReactor{
		BaseReactor: *NewBaseReactor("Switch-2"),
		msgReceived: make(map[byte][]PeerMsg),
		counter:     0,
	}
	reactor3 := &TestReactor{
		BaseReactor: *NewBaseReactor("Switch-1"),
		msgReceived: make(map[byte][]PeerMsg),
		counter:     0,
	}
	reactor4 := &TestReactor{
		BaseReactor: *NewBaseReactor("Switch-2"),
		msgReceived: make(map[byte][]PeerMsg),
		counter:     0,
	}
	sw1.AddReactor("test reactor", reactor1)
	sw2.AddReactor("test reactor", reactor2)
	sw3.AddReactor("test reactor", reactor3)
	sw4.AddReactor("test reactor", reactor4)

	assert.Nil(t, sw1.Start())
	assert.Nil(t, sw2.Start())
	assert.Nil(t, sw3.Start())
	assert.Nil(t, sw4.Start())

	addrs := []*NetAddress{sw1.NetAddress(), sw2.NetAddress(), sw3.NetAddress(), sw4.NetAddress()}

	connect := func(sw *Switch) {
		for _, addr := range addrs {
			_ = sw.DialPeerWithAddress(addr)
		}
	}

	go connect(sw1)
	go connect(sw2)
	go connect(sw3)
	go connect(sw4)

	connected := make(chan struct{})

	go func() {
		for {
			if sw1.peers.Size() == 3 && sw2.peers.Size() == 3 && sw3.peers.Size() == 3 && sw4.peers.Size() == 3 {
				close(connected)
				return
			}
		}
	}()

	<-connected

	time.Sleep(time.Millisecond * 200)

	sw1.Broadcast(0x01, []byte("hello, greeting from node1"))
	time.Sleep(time.Millisecond * 200)
	assert.Equal(t, 0, reactor1.counter)
	assert.Equal(t, 1, reactor2.counter)
	assert.Equal(t, 1, reactor3.counter)
	assert.Equal(t, 1, reactor4.counter)

	sw2.Broadcast(0x01, []byte("hello, greeting from node2"))
	time.Sleep(time.Millisecond * 200)
	assert.Equal(t, 1, reactor1.counter)
	assert.Equal(t, 1, reactor2.counter)
	assert.Equal(t, 2, reactor3.counter)
	assert.Equal(t, 2, reactor4.counter)

	time.Sleep(4 * time.Second)

	assert.Equal(t, []byte("hello, greeting from node1"), reactor2.msgReceived[0x01][0].msg)
	assert.Equal(t, sw1.NodeInfo().NodeID, reactor2.msgReceived[0x01][0].peerID)

	assert.Nil(t, sw1.Stop())
	assert.Nil(t, sw2.Stop())
	assert.Nil(t, sw3.Stop())
	assert.Nil(t, sw4.Stop())
}
