package p2p

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/232425wxy/meta--/common/hexbytes"
	config2 "github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/log"
	"github.com/stretchr/testify/assert"
)

type TestReactor struct {
	BaseReactor
	msgReceived map[byte][]PeerMsg
	counter     int
}

func (tr *TestReactor) Receive(chID byte, peer *Peer, msg []byte) {
	peerMsg := PeerMsg{
		peerID: peer.NodeID(),
		msg:    msg,
	}
	if tr.msgReceived[chID] == nil {
		tr.msgReceived[chID] = make([]PeerMsg, 0)
	}
	tr.msgReceived[chID] = append(tr.msgReceived[chID], peerMsg)
	tr.counter++
}

func (tr *TestReactor) GetChannels() []*ChannelDescriptor {
	return []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: defaultSendQueueCapacity, RecvMessageCapacity: defaultRecvMessageCapacity, RecvBufferCapacity: defaultRecvBufferCapacity},
	}
}

type PeerMsg struct {
	peerID crypto.ID
	msg    []byte
}

var (
	channelDescs []*ChannelDescriptor
	reactorByChA map[byte]Reactor
	reactorByChB map[byte]Reactor
	metricsA     *Metrics
	metricsB     *Metrics
	onPeerErrorA func(p *Peer, err error)
	onPeerErrorB func(p *Peer, err error)
	peerCfgA     peerConfig
	peerCfgB     peerConfig
	loggerA      log.Logger
	loggerB      log.Logger
)

func init() {
	channelDescs = []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: defaultSendQueueCapacity, RecvMessageCapacity: defaultRecvMessageCapacity, RecvBufferCapacity: defaultRecvBufferCapacity},
	}
	reactorA := new(TestReactor)
	reactorA.BaseReactor = *NewBaseReactor("test/reactorA")
	reactorA.msgReceived = make(map[byte][]PeerMsg, 0)
	reactorByChA = map[byte]Reactor{0x01: reactorA}
	metricsA = P2PMetrics()
	onPeerErrorA = func(p *Peer, err error) {}
	peerCfgA = peerConfig{
		chDescs:      channelDescs,
		onPeerError:  onPeerErrorA,
		reactorsByCh: reactorByChA,
		metrics:      metricsA,
	}

	reactorB := new(TestReactor)
	reactorB.BaseReactor = *NewBaseReactor("test/reactorA")
	reactorB.msgReceived = make(map[byte][]PeerMsg, 0)
	reactorByChB = map[byte]Reactor{0x01: reactorB}
	metricsB = P2PMetrics()
	onPeerErrorB = func(p *Peer, err error) {}
	peerCfgB = peerConfig{
		chDescs:      channelDescs,
		onPeerError:  onPeerErrorB,
		reactorsByCh: reactorByChB,
		metrics:      metricsB,
	}

	loggerA = log.New("peer", "A")
	loggerA.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	loggerB = log.New("peer", "B")
	loggerB.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
}

func getFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = l.Close()
	}()
	return l.Addr().(*net.TCPAddr).Port
}

func testNodeInfo(privateKey *bls12.PrivateKey) *NodeInfo {
	id := privateKey.PublicKey().ToID()
	channels := hexbytes.HexBytes{0x01}
	listen := fmt.Sprintf("127.0.0.1:%d", getFreePort())
	fmt.Printf("节点%s监听地址：%s\n", id, listen)
	return &NodeInfo{
		PublicKey: privateKey.PublicKey().ToBytes(),
		NodeID:     id,
		ListenAddr: listen,
		Channels:   channels,
		RPCAddress: fmt.Sprintf("127.0.0.1:%d", getFreePort()),
		TxIndex:    "on",
	}
}

func testNodeKey(privateKey *bls12.PrivateKey) *NodeKey {
	return &NodeKey{PrivateKey: privateKey}
}

func testNetAddress(id crypto.ID, addr string) *NetAddress {
	netAddr, err := NewNetAddressString(fmt.Sprintf("%s@%s", id, addr))
	if err != nil {
		panic(err)
	}
	return netAddr
}

func createTransport(privateKey *bls12.PrivateKey) *Transport {
	nodeInfo := testNodeInfo(privateKey)
	nodeKey := testNodeKey(privateKey)
	config := config2.DefaultP2PConfig()
	address := testNetAddress(nodeInfo.NodeID, nodeInfo.ListenAddr)
	transport := NewTransport(address, nodeInfo, nodeKey, config)
	return transport
}

func TestTransport_Listen(t *testing.T) {
	privateKey, err := bls12.GeneratePrivateKey()
	assert.Nil(t, err)
	transport := createTransport(privateKey)
	err = transport.Listen()
	assert.Nil(t, err)
	err = transport.Close()
	assert.Nil(t, err)
}

func TestTransportWaitForNewConn(t *testing.T) {
	privateKeyA, err := bls12.GeneratePrivateKey()
	assert.Nil(t, err)
	transportA := createTransport(privateKeyA)
	err = transportA.Listen()
	assert.Nil(t, err)
	defer func() {
		_ = transportA.Close()
	}()

	privateKeyB, err := bls12.GeneratePrivateKey()
	assert.Nil(t, err)
	transportB := createTransport(privateKeyB)
	addrB := testNetAddress(transportB.nodeInfo.NodeID, transportB.nodeInfo.ListenAddr)
	err = transportB.Listen()
	assert.Nil(t, err)
	defer func() {
		_ = transportB.Close()
	}()

	p, err := transportA.Dial(addrB, peerCfgA)
	assert.Nil(t, err)
	assert.Equal(t, transportB.nodeInfo.NodeID, p.NodeID())
	assert.Equal(t, transportB.nodeInfo.Channels, p.nodeInfo.Channels)
}

func TestPeerToPeer(t *testing.T) {
	log.PrintOrigins(true)
	privateKeyA, err := bls12.GeneratePrivateKey()
	assert.Nil(t, err)
	transportA := createTransport(privateKeyA)
	addrA := testNetAddress(transportA.nodeInfo.NodeID, transportA.nodeInfo.ListenAddr)
	err = transportA.Listen()
	assert.Nil(t, err)
	defer func() {
		_ = transportA.Close()
	}()

	privateKeyB, err := bls12.GeneratePrivateKey()
	assert.Nil(t, err)
	transportB := createTransport(privateKeyB)
	addrB := testNetAddress(transportB.nodeInfo.NodeID, transportB.nodeInfo.ListenAddr)
	err = transportB.Listen()
	assert.Nil(t, err)
	defer func() {
		_ = transportB.Close()
	}()

	peerC := make(chan *Peer)

	go func() {
		for {
			p, err := transportB.Accept(peerCfgB)
			if err != nil {
				t.Log(">>>>>>>>>", err)
				return
			}
			peerC <- p
		}
	}()

	peerB, err := transportA.Dial(addrB, peerCfgA)
	peerA := <-peerC

	peerA_, _ := transportB.Dial(addrA, peerCfgB)

	t.Log(peerA.NodeID())
	t.Log(peerA_.NodeID())

	peerA.SetLogger(loggerA)
	peerB.SetLogger(loggerB)

	assert.Nil(t, peerB.Start())
	assert.Nil(t, peerA.Start())

	sent := make(chan struct{})
	msg := []byte("hello")
	go func() {
		res := peerA.Send(0x01, msg)
		assert.True(t, res)
		close(sent)
	}()
	<-sent
	time.Sleep(time.Millisecond * 1000)
	assert.Equal(t, 1, reactorByChA[0x01].(*TestReactor).counter)
	assert.Equal(t, msg, reactorByChA[0x01].(*TestReactor).msgReceived[0x01][reactorByChA[0x01].(*TestReactor).counter-1].msg)
	assert.Equal(t, peerB.nodeInfo.NodeID, reactorByChA[0x01].(*TestReactor).msgReceived[0x01][reactorByChA[0x01].(*TestReactor).counter-1].peerID)
	time.Sleep(time.Second * 1)
}
