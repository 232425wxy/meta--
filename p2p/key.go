package p2p

import (
	"fmt"
	mos "github.com/232425wxy/meta--/common/os"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	mjson "github.com/232425wxy/meta--/json"
)

// NodeKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NodeKey 结构体里存储着一个BLS12-381的私钥。
type NodeKey struct {
	PrivateKey crypto.PrivateKey `json:"privateKey"`
}

// GetID ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// GetID 根据BLS12-381密钥信息获取对应的ID：去公钥的前10个字节，将这10个字节编码成16进制的字符串，
// 以此字符串作为节点的ID。
func (key *NodeKey) GetID() crypto.ID {
	return key.PrivateKey.PublicKey().ToID()
}

// PublicKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// PublicKey 获取BLS12-381私钥所对应的公钥。
func (key *NodeKey) PublicKey() crypto.PublicKey {
	return key.PrivateKey.PublicKey()
}

// SaveAs ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// SaveAs 给定存储节点密钥的地址，将节点密钥存储到那里面。
func (key *NodeKey) SaveAs(filePath string) error {
	bz, err := mjson.Encode(key)
	if err != nil {
		return err
	}
	return mos.WriteFile(filePath, bz, 0600)
}

// String ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// String 返回节点密钥的字符串形式：NodeKey{PrivateKey:"BLS12-381 PRIVATE KEY":{33184469658132716532202857962421420469965768660734559330213063713395516800091}}
func (key *NodeKey) String() string {
	return fmt.Sprintf("NodeKey{PrivateKey:%v}", key.PrivateKey.String())
}

// LoadOrGenNodeKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// LoadOrGenNodeKey 给定存储节点密钥的文件路径，如果该文件存在，就从文件中读取节点密钥，
// 否则就新生成节点的密钥。
func LoadOrGenNodeKey(filePath string) (*NodeKey, error) {
	if mos.FileExists(filePath) {
		return LoadNodeKey(filePath)
	}
	key, err := bls12.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	nodeKey := &NodeKey{PrivateKey: key}
	return nodeKey, nil
}

// LoadNodeKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// LoadNodeKey 给定存储节点密钥的文件路径，从中读取节点密钥信息。
func LoadNodeKey(filePath string) (*NodeKey, error) {
	bz, err := mos.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	nodeKey := new(NodeKey)
	err = mjson.Decode(bz, nodeKey)
	if err != nil {
		return nil, err
	}
	return nodeKey, nil
}
