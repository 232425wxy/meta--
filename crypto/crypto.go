package crypto

type ThresholdSignature interface {
	Type() string
	ToBytes() []byte
	Participants() *IDSet
}

type Signature interface {
	Type() string
	ToBytes() []byte
	Signer() ID
}

type PublicKey interface {
	Type() string
	ToBytes() []byte
	FromBytes(bz []byte) error
	Verify(sig Signature, h [32]byte) bool
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 项目级全局常量

// TruncatePublicKeyLength ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
// ---------------------------------------------------------
// TruncatePublicKeyLength 代表的是一个长度，这个长度是指要截取公钥字节的长度，在利用公钥生成节点ID时有用。
const TruncatePublicKeyLength = 10
