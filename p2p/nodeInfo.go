package p2p

import (
	"bytes"
	"fmt"
	"github.com/232425wxy/meta--/common/hexbytes"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/proto/pbp2p"
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
	PublicKey   []byte            `json:"public_key"`
	NodeID      crypto.ID         `json:"nodeID"`
	ListenAddr  string            `json:"listenAddr"` // 监听的网络地址，从该地址获取新连接
	Channels    hexbytes.HexBytes // 节点管理的所有信道
	RPCAddress  string            // rpc通信地址
	TxIndex     string
	CryptoBLS12 *bls12.CryptoBLS12
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
	// 首先就是验证节点监听的地址是否正确，监听地址一般都是："0.0.0.0:端口号"。
	_, err := NewNetAddressString(IDAddressString(node.ID(), node.ListenAddr))
	if err != nil {
		return err
	}
	if len(node.Channels) > maxNumChannels {
		return fmt.Errorf("node %q has to many channels %q", node.ID(), node.Channels)
	}
	channels := make(map[byte]struct{})
	for _, ch := range node.Channels {
		_, ok := channels[ch]
		if ok {
			return fmt.Errorf("node %q has duplicate channel %q", node.ID(), ch)
		}
		channels[ch] = struct{}{}
	}
	if node.TxIndex != "on" && node.TxIndex != "off" {
		return fmt.Errorf("node %q should make TxIndex \"on\" or \"off\", but he made it %q", node.ID(), node.TxIndex)
	}
	return nil
}

// CompatibleWith ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// CompatibleWith 检测两个节点是否兼容。
func (node NodeInfo) CompatibleWith(other NodeInfo) error {
	if !node.Channels.CompatibleWith(other.Channels) {
		return fmt.Errorf("we have completely different channels, i have %q, but the other side has %q", node.Channels, other.Channels)
	}
	return nil
}

// NetAddress ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NetAddress 通过节点的ID和监听地址，返回节点的 NetAddress 实例。
func (node NodeInfo) NetAddress() (*NetAddress, error) {
	return NewNetAddressString(IDAddressString(node.ID(), node.ListenAddr))
}

// HasChannel ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// HasChannel 判断节点是否含有指定的信道。
func (node NodeInfo) HasChannel(ch byte) bool {
	return bytes.Contains(node.Channels, []byte{ch})
}

// ToProto ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToProto 将自定义的 NodeInfo 转换为protobuf形式。
func (node NodeInfo) ToProto() *pbp2p.NodeInfo {
	return &pbp2p.NodeInfo{
		PublicKey:  node.PublicKey,
		NodeID:     string(node.NodeID),
		ListenAddr: node.ListenAddr,
		Channels:   node.Channels,
		RPCAddress: node.TxIndex,
		TxIndex:    node.TxIndex,
	}
}

// NodeInfoFromProto ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NodeInfoFromProto 将protobuf形式的NodeInfo转换成自定义的 NodeInfo。
func NodeInfoFromProto(pbInfo *pbp2p.NodeInfo) *NodeInfo {
	return &NodeInfo{
		PublicKey:  pbInfo.PublicKey,
		NodeID:     crypto.ID(pbInfo.NodeID),
		ListenAddr: pbInfo.ListenAddr,
		Channels:   pbInfo.Channels,
		RPCAddress: pbInfo.RPCAddress,
		TxIndex:    pbInfo.TxIndex,
	}
}
