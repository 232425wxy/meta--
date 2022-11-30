package bls12

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12/bls12381"
	"math/big"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 定义项目级全局函数

// GeneratePrivateKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// GeneratePrivateKey 根据定义的椭圆曲线G1群的阶 curveOrder 随机生成一个数作为私钥。
func GeneratePrivateKey() (*PrivateKey, error) {
	key, err := rand.Int(rand.Reader, curveOrder)
	if err != nil {
		return nil, fmt.Errorf("bls12: failed to generate private key: %q", err)
	}
	return &PrivateKey{key: key}, nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义项目级全局变量：公私钥对

// PublicKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// PublicKey 是bls12-381的公钥。
type PublicKey struct {
	key *bls12381.PointG1
}

// ToBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToBytes 将公钥序列化成字节切片。
func (pub *PublicKey) ToBytes() []byte {
	return bls12381.NewG1().ToCompressed(pub.key)
}

// FromBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FromBytes 给定一个公钥的字节切片，对其进行反序列化，得到公钥对象。
func (pub *PublicKey) FromBytes(bz []byte) (err error) {
	pub.key, err = bls12381.NewG1().FromCompressed(bz)
	if err != nil {
		return fmt.Errorf("bls12: failed to decompress public key: %q", err)
	}
	return nil
}

// PrivateKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// PrivateKey 是bls12-381的私钥，实际上私钥用 *big.Int 表示。
type PrivateKey struct {
	key *big.Int
}

// ToBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToBytes 返回私钥的字节切片内容，其实就是返回 *big.Int 的字节切片内容。
func (private *PrivateKey) ToBytes() []byte {
	return private.key.Bytes()
}

// FromBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FromBytes 根据给定的字节切片，将其转换成私钥，其实就是将字节切片转换为 *big.Int。
func (private *PrivateKey) FromBytes(bz []byte) error {
	private.key = new(big.Int)
	private.key.SetBytes(bz)
	return nil
}

// Public ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Public 返回与当前私钥关联的公钥。
func (private *PrivateKey) Public() crypto.PublicKey {
	key := &bls12381.PointG1{}
	return &PublicKey{key: bls12381.NewG1().MulScalarBig(key, &bls12381.G1One, private.key)}
}

// Signature ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Signature 是一个bls12-381的签名。
type Signature struct {
	signer crypto.ID
	sig    *bls12381.PointG2
}

// ToBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToBytes 将签名转换成字节切片形式并返回。
func (s *Signature) ToBytes() []byte {
	var id [4]byte
	copy(id[:], s.signer.ToBytes())
	return append(id[:], bls12381.NewG2().ToCompressed(s.sig)...)
}

// FromBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FromBytes 接受签名的字节切片形式的内容，然后将其转换为 Signature 对象。
func (s *Signature) FromBytes(bz []byte) (err error) {
	s.signer = crypto.ID(binary.LittleEndian.Uint32(bz))
	s.sig, err = bls12381.NewG2().FromCompressed(bz[4:])
	if err != nil {
		return fmt.Errorf("bls12: failed to decompress signature: %q", err)
	}
	return nil
}

// Signer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Signer 返回签名者的id号。
func (s *Signature) Signer() crypto.ID {
	return s.signer
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 项目级全局常量

const (
	// PrivateKeyFileType ♏ |作者：吴翔宇| 🍁 |日期：2022/11/30|
	//
	// PrivateKeyFileType PEM格式的私钥。
	PrivateKeyFileType = "BLS12-381 PRIVATE KEY"

	// PublicKeyFileType ♏ |作者：吴翔宇| 🍁 |日期：2022/11/30|
	//
	// PublicKeyFileType PEM格式的公钥。
	PublicKeyFileType = "BLS12-381 PUBLIC KEY"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级全局变量

var domain = []byte("BLS12-381-SIG:REDACTABLE-BLOCKCHAIN")

// curveOrder ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// curveOrder 椭圆曲线G1的阶。
var curveOrder, _ = new(big.Int).SetString("73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001", 16)
