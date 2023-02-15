package bls12

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12/bls12381"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/json"
	"github.com/232425wxy/meta--/proto/pbcrypto"
	"go.uber.org/multierr"
	"math/big"
	"sync"
)

func init() {
	lib = new(pubKeyLib)
	lib.keys = make(map[crypto.ID]*PublicKey)

	json.RegisterType(&PublicKey{}, PublicKeyFileType)
	json.RegisterType(&PrivateKey{}, PrivateKeyFileType)
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
		return nil, fmt.Errorf("bls12: failed to generate private Key: %q", err)
	}
	return &PrivateKey{Key: key}, nil
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
		sig:          s,
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
		return fmt.Errorf("bls12: add public Key failed: %q", err)
	}
	id := public.ToID()
	lib.keys[id] = public
	return nil
}

// GetBLSPublicKeyFromLib ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// GetBLSPublicKeyFromLib 给定一个节点的ID，从库里获取该节点的公钥。
func GetBLSPublicKeyFromLib(id crypto.ID) *PublicKey {
	lib.mu.RLock()
	defer lib.mu.RUnlock()
	if key, ok := lib.keys[id]; ok {
		return key
	}
	return nil
}

// PublicKeyFromProto ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// PublicKeyFromProto 将protobuf形式的公钥转换为我们自定义的公钥。
func PublicKeyFromProto(pb *pbcrypto.BLS12PublicKey) *PublicKey {
	pub := new(PublicKey)
	err := pub.FromBytes(pb.Key)
	if err != nil {
		panic(err)
	}
	return pub
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义项目级全局变量：公私钥对

// PublicKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// PublicKey 是bls12-381的公钥。
type PublicKey struct {
	Key *bls12381.PointG1
}

// Verify ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Verify 验证签名。
func (pub *PublicKey) Verify(sig *Signature, h []byte) bool {
	p, err := bls12381.NewG2().HashToCurve(h[:], domain)
	if err != nil {
		return false
	}
	engine := bls12381.NewEngine()
	engine.AddPairInv(&bls12381.G1One, sig.sig)
	engine.AddPair(pub.Key, p)
	return engine.Result().IsOne()
}

// ToID ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToID 将节点的公钥转换成节点的ID。
func (pub *PublicKey) ToID() crypto.ID {
	bz := pub.ToBytes()[:TruncatePublicKeyLength]
	id := crypto.ID(hex.EncodeToString(bz))
	return id
}

// ToBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToBytes 将公钥序列化成字节切片。
func (pub *PublicKey) ToBytes() []byte {
	return bls12381.NewG1().ToCompressed(pub.Key)
}

// FromBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FromBytes 给定一个公钥的字节切片，对其进行反序列化，得到公钥对象。
func (pub *PublicKey) FromBytes(bz []byte) (err error) {
	pub.Key, err = bls12381.NewG1().FromCompressed(bz)
	if err != nil {
		return fmt.Errorf("bls12: failed to decompress public Key: %q", err)
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

// ToProto ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToProto 将公钥转换为protobuf形式。
func (pub *PublicKey) ToProto() *pbcrypto.BLS12PublicKey {
	key := new(pbcrypto.BLS12PublicKey)
	key.Key = pub.ToBytes()
	return key
}

// PrivateKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// PrivateKey 是bls12-381的私钥，实际上私钥用 *big.Int 表示。
type PrivateKey struct {
	Key *big.Int
}

// Sign ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Sign 生成签名消息。
func (private *PrivateKey) Sign(h []byte) (sig *Signature, err error) {
	p, err := bls12381.NewG2().HashToCurve(h[:], domain)
	if err != nil {
		return nil, fmt.Errorf("bls12: hash to curve failed: %q", err)
	}
	bls12381.NewG2().MulScalarBig(p, p, private.Key)
	return &Signature{signer: private.PublicKey().ToID(), sig: p}, nil
}

// ToBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToBytes 返回私钥的字节切片内容，其实就是返回 *big.Int 的字节切片内容。
func (private *PrivateKey) ToBytes() []byte {
	return private.Key.Bytes()
}

// FromBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FromBytes 根据给定的字节切片，将其转换成私钥，其实就是将字节切片转换为 *big.Int。
func (private *PrivateKey) FromBytes(bz []byte) error {
	private.Key = new(big.Int)
	private.Key.SetBytes(bz)
	return nil
}

// PublicKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// PublicKey 返回与当前私钥关联的公钥。
func (private *PrivateKey) PublicKey() *PublicKey {
	key := &bls12381.PointG1{}
	return &PublicKey{Key: bls12381.NewG1().MulScalarBig(key, &bls12381.G1One, private.Key)}
}

// Type ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Type 返回私钥类型："BLS12-381 PRIVATE KEY"。
func (private *PrivateKey) Type() string {
	return "BLS12-381 PRIVATE KEY"
}

// String ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// String 返回BLS12-381私钥的字符串格式："BLS12-381 PRIVATE KEY":{33184469658132716532202857962421420469965768660734559330213063713395516800091}
func (private *PrivateKey) String() string {
	return fmt.Sprintf(`"%s":{%s}`, private.Type(), private.Key)
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
	var id [TruncatePublicKeyLength]byte
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
	s.signer, err = crypto.FromBytesToID(bz[:TruncatePublicKeyLength])
	if err != nil {
		return err
	}
	s.sig, err = bls12381.NewG2().FromCompressed(bz[TruncatePublicKeyLength:])
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

func (s *Signature) ToProto() *pbcrypto.Signature {
	sig := bls12381.NewG2().ToCompressed(s.sig)
	return &pbcrypto.Signature{
		Signer: string(s.signer),
		Sig:    sig,
	}
}

func SignatureFromProto(pb *pbcrypto.Signature) *Signature {
	sig, err := bls12381.NewG2().FromCompressed(pb.Sig)
	if err != nil {
		panic(err)
	}
	return &Signature{
		signer: crypto.ID(pb.Signer),
		sig:    sig,
	}
}

// AggregateSignature ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// AggregateSignature 是bls12-381的聚合签名。
type AggregateSignature struct {
	sig          *bls12381.PointG2
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
	bz := bls12381.NewG2().ToCompressed(agg.sig)
	return bz
}

func (agg *AggregateSignature) FromBytes(bz []byte) (err error) {
	agg.sig, err = bls12381.NewG2().FromCompressed(bz)
	return err
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

func (agg *AggregateSignature) ToProto() *pbcrypto.AggregateSignature {
	if agg == nil {
		return nil
	}
	pb := &pbcrypto.AggregateSignature{
		Participants: make([]string, 0),
	}
	pb.Sig = agg.ToBytes()
	for _, participant := range agg.participants.IDs {
		pb.Participants = append(pb.Participants, string(participant))
	}
	return pb
}

func AggregateSignatureFromProto(pb *pbcrypto.AggregateSignature) *AggregateSignature {
	if pb == nil {
		return nil
	}
	agg := &AggregateSignature{
		participants: crypto.NewIDSet(0),
	}
	err := agg.FromBytes(pb.Sig)
	if err != nil {
		panic(err)
	}
	for _, participant := range pb.Participants {
		agg.participants.AddID(crypto.ID(participant))
	}
	return agg
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
	public := private.PublicKey()

	cb.private = private
	cb.public = public
	cb.id = public.ToID()
	err := AddBLSPublicKey(public.ToBytes())
	if err != nil {
		panic(err)
	}
}

// Sign ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Sign 对一个长度为256比特的哈希值进行签名。
func (cb *CryptoBLS12) Sign(h []byte) (*Signature, error) {
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
	sig := &bls12381.PointG2{}
	var participants = crypto.NewIDSet(0)
	for id, s := range signatures {
		g2.Add(sig, sig, s.sig)
		participants.AddID(id)
	}
	return &AggregateSignature{sig: sig, participants: participants}
}

// Verify ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Verify 给定一个签名，签名中包含签名者的ID，根据这个ID去找到这个签名者的公钥，然后验证这个签名是否合法。
func (cb *CryptoBLS12) Verify(sig *Signature, h []byte) bool {
	signerPubKey := GetBLSPublicKeyFromLib(sig.Signer())
	if signerPubKey == nil {
		return false
	}
	return signerPubKey.Verify(sig, h)
}

// VerifyThresholdSignature ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// VerifyThresholdSignature 验证聚合签名。
func (cb *CryptoBLS12) VerifyThresholdSignature(signature *AggregateSignature, h []byte) bool {
	pubKeys := make([]*PublicKey, 0)
	for _, participant := range signature.Participants().IDs {
		pubKey := GetBLSPublicKeyFromLib(participant)
		if pubKey != nil {
			pubKeys = append(pubKeys, pubKey)
		}
	}
	ps, err := bls12381.NewG2().HashToCurve(h[:], domain)
	if err != nil {
		return false
	}

	//if len(pubKeys) < quorumSize {
	//	return false
	//}
	engine := bls12381.NewEngine()
	engine.AddPairInv(&bls12381.G1One, signature.sig)
	for _, key := range pubKeys {
		engine.AddPair(key.Key, ps)
	}
	return engine.Result().IsOne()
}

// VerifyThresholdSignatureForMessageSet ♏ |作者：吴翔宇| 🍁 |日期：2022/11/30|
//
// VerifyThresholdSignatureForMessageSet 根据给定的聚合签名和不同消息的哈希值，验证聚合签名是否合法。
func (cb *CryptoBLS12) VerifyThresholdSignatureForMessageSet(signature *AggregateSignature, hashes map[crypto.ID]sha256.Hash, quorumSize int) bool {
	hashSet := make(map[sha256.Hash]struct{})
	engine := bls12381.NewEngine()
	engine.AddPairInv(&bls12381.G1One, signature.sig)
	for id, hash := range hashes {
		if _, ok := hashSet[hash]; ok {
			continue
		}
		hashSet[hash] = struct{}{}
		pubKey := GetBLSPublicKeyFromLib(id)
		if pubKey == nil {
			return false
		}
		p2, err := bls12381.NewG2().HashToCurve(hash[:], domain)
		if err != nil {
			return false
		}
		engine.AddPair(pubKey.Key, p2)
	}
	if !engine.Result().IsOne() {
		return false
	}
	return len(hashSet) >= quorumSize
}

// CreateThresholdSignature ♏ |作者：吴翔宇| 🍁 |日期：2022/11/30|
//
// CreateThresholdSignature 根据给定的部分签名创建聚合签名。
func (cb *CryptoBLS12) CreateThresholdSignature(partialSignatures []*Signature) (_ *AggregateSignature, err error) {
	sigs := make(map[crypto.ID]*Signature, len(partialSignatures))
	for _, sig := range partialSignatures {
		if _, ok := sigs[sig.Signer()]; ok {
			err = multierr.Append(err, fmt.Errorf("bls12: duplicate partial signature from ID: %q", sig.Signer()))
			continue
		}
		sigs[sig.Signer()] = sig
	}
	return cb.aggregateSignatures(sigs), nil
}

// CreateThresholdSignatureForMessageSet ♏ |作者：吴翔宇| 🍁 |日期：2022/11/30|
//
// CreateThresholdSignatureForMessageSet 将若干个为不同消息签名的签名聚合成聚合签名。
func (cb *CryptoBLS12) CreateThresholdSignatureForMessageSet(partialSignatures []*Signature) (*AggregateSignature, error) {
	return cb.CreateThresholdSignature(partialSignatures)

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

	// TruncatePublicKeyLength ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
	// ---------------------------------------------------------
	// TruncatePublicKeyLength 代表的是一个长度，这个长度是指要截取公钥字节的长度，在利用公钥生成节点ID时有用。
	TruncatePublicKeyLength = 10
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
type pubKeyLib struct {
	mu   sync.RWMutex
	keys map[crypto.ID]*PublicKey
}

var lib *pubKeyLib

// curveOrder ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// curveOrder 椭圆曲线G1的阶。
var curveOrder, _ = new(big.Int).SetString("73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001", 16)
