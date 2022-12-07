package p2p

import (
	"github.com/232425wxy/meta--/common/hexbytes"
	"github.com/232425wxy/meta--/crypto"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义常量

const (
	maxNodeInfoSize = 1024 * 10 // 一个p2p节点的信息大小最多不能超过10KB
	maxNumChannels  = 16        // 一个p2p节点最多可以管理16个信道
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义p2p网络中节点的基本信息

type NodeInfo struct {
	NodeID     crypto.ID         `json:"nodeID"`
	ListenAddr string            `json:"listenAddr"` // 监听的网络地址，从该地址获取新连接
	ChainID    string            // 区块链的ID号，类似于以太坊中的网络号
	Channels   hexbytes.HexBytes // 节点管理的所有信道
	RPCAddress string            // rpc通信地址
	TxIndex    string
}

// ID ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ID 返回节点的ID号。
func (node NodeInfo) ID() crypto.ID {
	return node.NodeID
}

// Validate ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Validate 验证节点的基本信息是否合法。
func (node NodeInfo) Validate() error {

}
