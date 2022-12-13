package p2p

import (
	"fmt"
	"github.com/232425wxy/meta--/common/hexbytes"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func getFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func testNodeInfo(privateKey *bls12.PrivateKey) *NodeInfo {
	id := privateKey.PublicKey().ToID()
	channels := hexbytes.HexBytes{0x01}
	return &NodeInfo{
		NodeID:     id,
		ListenAddr: fmt.Sprintf("127.0.0.1:%d", getFreePort()),
		ChainID:    "100",
		Channels:   channels,
		RPCAddress: fmt.Sprintf("127.0.0.1:%d", getFreePort()),
		TxIndex:    "on",
	}
}

func testNodeKey(privateKey *bls12.PrivateKey) *NodeKey {
	return &NodeKey{PrivateKey: privateKey}
}

func testConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		SendRate:                5120000,
		RecvRate:                5120000,
		MaxPacketMsgPayloadSize: 1024,
		FlushDur:                50 * time.Millisecond,
		PingInterval:            90 * time.Millisecond,
		PongTimeout:             45 * time.Millisecond,
	}
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
	config := testConnectionConfig()
	transport := NewTransport(nodeInfo, nodeKey, config)
	return transport
}

func TestTransport_Listen(t *testing.T) {
	privateKey, err := bls12.GeneratePrivateKey()
	assert.Nil(t, err)
	transport := createTransport(privateKey)
	addr := testNetAddress(transport.nodeInfo.NodeID, transport.nodeInfo.ListenAddr)
	err = transport.Listen(addr)
	assert.Nil(t, err)
	err = transport.Close()
	assert.Nil(t, err)
}
