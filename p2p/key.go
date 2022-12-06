package p2p

import "github.com/232425wxy/meta--/crypto"

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
