package p2p

import (
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type node struct {
	info *NodeInfo
	key  *NodeKey
}

func create2Nodes(t *testing.T) (*node, *node) {
	privateKeyA, errA := bls12.GeneratePrivateKey()
	assert.Nil(t, errA)
	privateKeyB, errB := bls12.GeneratePrivateKey()
	assert.Nil(t, errB)
	nodeA := testNodeInfo(privateKeyA)
	nodeB := testNodeInfo(privateKeyB)
	keyA := testNodeKey(privateKeyA)
	keyB := testNodeKey(privateKeyB)
	return &node{info: nodeA, key: keyA}, &node{info: nodeB, key: keyB}
}

func create2Transports(t *testing.T) (*Transport, *Transport) {
	nodeA, nodeB := create2Nodes(t)
	transportA := createTransport(nodeA.key.PrivateKey)
	transportB := createTransport(nodeB.key.PrivateKey)
	return transportA, transportB
}

func create2Switches(t *testing.T) (*Switch, *Switch) {
	transportA, transportB := create2Transports(t)
	cfgA := config.DefaultP2PConfig()
	cfgA.PongTimeout = 45 * time.Millisecond
	cfgA.PingInterval = 90 * time.Millisecond
	cfgB := config.DefaultP2PConfig()
	cfgB.PongTimeout = 45 * time.Millisecond
	cfgB.PingInterval = 90 * time.Millisecond
	sw1 := NewSwitch(cfgA, transportA, metricsA)
	sw2 := NewSwitch(cfgB, transportB, metricsB)
	return sw1, sw2
}
