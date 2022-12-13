package p2p

import (
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/stretchr/testify/assert"
	"testing"
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
