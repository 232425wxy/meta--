package bls12

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12/bls12381"
	"github.com/232425wxy/meta--/crypto/hash/sha256"
	"go.uber.org/multierr"
	"math/big"
	"sync"
)

func init() {
	lib = new(pubLeyLib)
	lib.keys = make(map[crypto.ID]*PublicKey)
}

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

// RestoreAggregateSignature ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// RestoreAggregateSignature 用于恢复一个聚合签名，该方法不能用于创建一个新的聚合签名。
func RestoreAggregateSignature(sig []byte, participants *crypto.IDSet) (*AggregateSignature, error) {
	s, err := bls12381.NewG2().FromCompressed(sig)
	if err != nil {
		return nil, fmt.Errorf("bls12: failed to restore aggregate signature: %q", err)
	}
	return &AggregateSignature{
		sig:          *s,
		participants: participants,
	}, nil
}

// AddBLSPublicKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// AddBLSPublicKey 给定一个节点的公钥（字节切片形式），将该公钥添加到库里。
func AddBLSPublicKey(bz []byte) error {
	lib.mu.Lock()
	defer lib.mu.Unlock()
	public := new(PublicKey)
	err := public.FromBytes(bz)
	if err != nil {
		return fmt.Errorf("bls12: add public key failed: %q", err)
	}
	id := public.ToID()
	lib.keys[id] = public
	return nil
}

// GetBLSPublicKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// GetBLSPublicKey 给定一个节点的ID，从库里获取该节点的公钥。
func GetBLSPublicKey(id crypto.ID) *PublicKey {
	lib.mu.RLock()
	defer lib.mu.RUnlock()
	if key, ok := lib.keys[id]; ok {
		return key
	}
	return nil
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

// Verify ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Verify 验证签名。
func (pub *PublicKey) Verify(sig crypto.Signature, h sha256.Hash) bool {
	s, ok := sig.(*Signature)
	if !ok {
		panic(fmt.Sprintf("bls12: need bls12-381 signature, but got %q", sig.Type()))
	}
	p, err := bls12381.NewG2().HashToCurve(h[:], domain)
	if err != nil {
		return false
	}
	engine := bls12381.NewEngine()
	engine.AddPairInv(&bls12381.G1One, s.sig)
	engine.AddPair(pub.key, p)
	return engine.Result().IsOne()
}

// ToID ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToID 将节点的公钥转换成节点的ID。
func (pub *PublicKey) ToID() crypto.ID {
	bz := pub.ToBytes()[:crypto.TruncatePublicKeyLength]
	id := crypto.ID(hex.EncodeToString(bz))
	return id
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

// Type ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Type 返回公钥类型："BLS12-381 PUBLIC KEY"。
func (pub *PublicKey) Type() string {
	return "BLS12-381 PUBLIC KEY"
}

// PrivateKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// PrivateKey 是bls12-381的私钥，实际上私钥用 *big.Int 表示。
type PrivateKey struct {
	key *big.Int
}

// Sign ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Sign 生成签名消息。
func (private *PrivateKey) Sign(h sha256.Hash) (sig crypto.Signature, err error) {
	p, err := bls12381.NewG2().HashToCurve(h[:], domain)
	if err != nil {
		return nil, fmt.Errorf("bls12: hash to curve failed: %q", err)
	}
	bls12381.NewG2().MulScalarBig(p, p, private.key)
	return &Signature{signer: private.Public().ToID(), sig: p}, nil
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
func (private *PrivateKey) Public() *PublicKey {
	key := &bls12381.PointG1{}
	return &PublicKey{key: bls12381.NewG1().MulScalarBig(key, &bls12381.G1One, private.key)}
}

// Type ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Type 返回私钥类型："BLS12-381 PRIVATE KEY"。
func (private *PrivateKey) Type() string {
	return "BLS12-381 PRIVATE KEY"
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
	var id [crypto.TruncatePublicKeyLength]byte
	bz := s.signer.ToBytes()
	copy(id[:], bz)
	return append(id[:], bls12381.NewG2().ToCompressed(s.sig)...)
}

// FromBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FromBytes 接受签名的字节切片形式的内容，然后将其转换为 Signature 对象。
func (s *Signature) FromBytes(bz []byte) (err error) {
	s.signer, err = crypto.FromBytesToID(bz[:crypto.TruncatePublicKeyLength])
	if err != nil {
		return err
	}
	s.sig, err = bls12381.NewG2().FromCompressed(bz[crypto.TruncatePublicKeyLength:])
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

// Type ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Type 返回签名的类型："BLS12-381 SIGNATURE"。
func (s *Signature) Type() string {
	return "BLS12-381 SIGNATURE"
}

// AggregateSignature ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// AggregateSignature 是bls12-381的聚合签名。
type AggregateSignature struct {
	sig          bls12381.PointG2
	participants *crypto.IDSet
}

// ToBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToBytes 返回聚合签名的字节切片形式。
func (agg *AggregateSignature) ToBytes() []byte {
	if agg == nil {
		return nil
	}
	bz := bls12381.NewG2().ToCompressed(&agg.sig)
	return bz
}

// Participants ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Participants 返回参与门限签名的节点集合。
func (agg *AggregateSignature) Participants() *crypto.IDSet {
	set := crypto.NewIDSet(agg.participants.Size())
	copy(set.IDs, agg.participants.IDs)
	return set
}

// Type ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Type 返回聚合签名的类型："BLS12-381 THRESHOLD SIGNATURE"。
func (agg *AggregateSignature) Type() string {
	return "BLS12-381 THRESHOLD SIGNATURE"
}

// CryptoBLS12 ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// CryptoBLS12 实现了bls12-381聚合签名的的签名和验证功能。
type CryptoBLS12 struct {
	private *PrivateKey
	public  *PublicKey
	id      crypto.ID
}

// NewCryptoBLS12 ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewCryptoBLS12 创建一个新的 *CryptoBLS12，现在它的公私钥还是空的，需要调用 Init 方法来对它
// 进行初始化。
func NewCryptoBLS12() *CryptoBLS12 {
	return &CryptoBLS12{}
}

// Init ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Init 初始化，给 *blsCrypto 设置私钥和节点ID。
func (cb *CryptoBLS12) Init(private *PrivateKey) {
	public := private.Public()

	cb.private = private
	cb.public = public
	cb.id = public.ToID()
}

// Sign ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Sign 对一个长度为256比特的哈希值进行签名。
func (cb *CryptoBLS12) Sign(h sha256.Hash) (crypto.Signature, error) {
	sig, err := cb.private.Sign(h)
	return sig, err
}

// aggregateSignatures ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// aggregateSignatures 将一众签名聚合到一起。
func (cb *CryptoBLS12) aggregateSignatures(signatures map[crypto.ID]*Signature) *AggregateSignature {
	if len(signatures) == 0 {
		return nil
	}
	g2 := bls12381.NewG2()
	sig := bls12381.PointG2{}
	var participants = crypto.NewIDSet(0)
	for id, s := range signatures {
		g2.Add(&sig, &sig, s.sig)
		participants.AddID(id)
	}
	return &AggregateSignature{sig: sig, participants: participants}
}

// Verify ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Verify 给定一个签名，签名中包含签名者的ID，根据这个ID去找到这个签名者的公钥，然后验证这个签名是否合法。
func (cb *CryptoBLS12) Verify(sig crypto.Signature, h [32]byte) bool {
	signerPubKey := GetBLSPublicKey(sig.Signer())
	if signerPubKey == nil {
		return false
	}
	return signerPubKey.Verify(sig, h)
}

// VerifyThresholdSignature ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//  ---------------------------------------------------------
// VerifyThresholdSignature 验证聚合签名。
func (cb *CryptoBLS12) VerifyThresholdSignature(signature crypto.ThresholdSignature, h sha256.Hash, quorumSize int) bool {
	sig, ok := signature.(*AggregateSignature)
	if !ok {
		panic(fmt.Sprintf("bls12: need bls12-381 threshold signature, but got %q", signature.Type()))
	}
	pubKeys := make([]*PublicKey, 0)
	for _, participant := range sig.Participants().IDs {
		pubKey := GetBLSPublicKey(participant)
		if pubKey != nil {
			pubKeys = append(pubKeys, pubKey)
		}
	}
	ps, err := bls12381.NewG2().HashToCurve(h[:], domain)
	if err != nil {
		return false
	}
	if len(pubKeys) < quorumSize {
		return false
	}
	engine := bls12381.NewEngine()
	engine.AddPairInv(&bls12381.G1One, &sig.sig)
	for _, key := range pubKeys {
		engine.AddPair(key.key, ps)
	}
	return engine.Result().IsOne()
}

// VerifyThresholdSignatureForMessageSet ♏ |作者：吴翔宇| 🍁 |日期：2022/11/30|
//
// VerifyThresholdSignatureForMessageSet 根据给定的聚合签名和不同消息的哈希值，验证聚合签名是否合法。
func (cb *CryptoBLS12) VerifyThresholdSignatureForMessageSet(signature crypto.ThresholdSignature, hashes map[crypto.ID]sha256.Hash, quorumSize int) bool {
	sig, ok := signature.(*AggregateSignature)
	if !ok {
		panic(fmt.Sprintf("bls12: need bls12-381 threshold signature, but got %q", signature.Type()))
	}
	hashSet := make(map[sha256.Hash]struct{})
	engine := bls12381.NewEngine()
	engine.AddPairInv(&bls12381.G1One, &sig.sig)
	for id, hash := range hashes {
		if _, ok = hashSet[hash]; ok {
			continue
		}
		hashSet[hash] = struct{}{}
		pubKey := GetBLSPublicKey(id)
		if pubKey == nil {
			return false
		}
		p2, err := bls12381.NewG2().HashToCurve(hash[:], domain)
		if err != nil {
			return false
		}
		engine.AddPair(pubKey.key, p2)
	}
	if !engine.Result().IsOne() {
		return false
	}
	return len(hashSet) >= quorumSize
}

// CreateThresholdSignature ♏ |作者：吴翔宇| 🍁 |日期：2022/11/30|
//
// CreateThresholdSignature 根据给定的部分签名创建聚合签名。
func (cb *CryptoBLS12) CreateThresholdSignature(partialSignatures []crypto.Signature, _ sha256.Hash, quorumSize int) (_ crypto.ThresholdSignature, err error) {
	if len(partialSignatures) < quorumSize {
		return nil, fmt.Errorf("bls12: not reach quorum size: %q", quorumSize)
	}
	sigs := make(map[crypto.ID]*Signature, len(partialSignatures))
	for _, sig := range partialSignatures {
		if _, ok := sigs[sig.Signer()]; ok {
			err = multierr.Append(err, fmt.Errorf("bls12: duplicate partial signature from ID: %q", sig.Signer()))
			continue
		}
		s, ok := sig.(*Signature)
		if !ok {
			err = multierr.Append(err, fmt.Errorf("bls12: need bls12-381 signature, but got %q from ID: %q", sig.Type(), sig.Signer()))
			continue
		}
		sigs[sig.Signer()] = s
	}
	if len(sigs) < quorumSize {
		return nil, multierr.Combine(err, fmt.Errorf("bls12: not reach quorum size: %q, only got %q", quorumSize, len(sigs)))
	}
	return cb.aggregateSignatures(sigs), nil
}

// CreateThresholdSignatureForMessageSet ♏ |作者：吴翔宇| 🍁 |日期：2022/11/30|
//
// CreateThresholdSignatureForMessageSet 将若干个为不同消息签名的签名聚合成聚合签名。
func (cb *CryptoBLS12) CreateThresholdSignatureForMessageSet(partialSignatures []crypto.Signature, hashes map[crypto.ID]sha256.Hash, quorumSize int) (crypto.ThresholdSignature, error) {
	return cb.CreateThresholdSignature(partialSignatures, sha256.Hash{}, quorumSize)
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

// domain ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// domain 在生成bls12-381签名和验证签名时被使用。
var domain = []byte("BLS12-381-SIG:REDACTABLE-BLOCKCHAIN")

// pubKeyLib ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// pubKeyLib 存储系统中其他节点的公钥库。
type pubLeyLib struct {
	mu   sync.RWMutex
	keys map[crypto.ID]*PublicKey
}

var lib *pubLeyLib

// curveOrder ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// curveOrder 椭圆曲线G1的阶。
var curveOrder, _ = new(big.Int).SetString("73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001", 16)
