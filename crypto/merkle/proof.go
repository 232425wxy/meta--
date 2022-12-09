package merkle

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/proto/pbcrypto"
	"math/bits"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 项目级全局函数

// ComputeMerkleRoot ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ComputeMerkleRoot 方法接受[][]byte类型的交易数据items，然后只计算这些交易数据组成的默克尔树的根哈希值并返回。
func ComputeMerkleRoot(items [][]byte) []byte {
	switch {
	case len(items) == 0:
		return nil
	case len(items) == 1:
		return leafHash(items[0])
	default:
		left := ComputeMerkleRoot(items[:getSplitPoint(uint64(len(items)))])
		right := ComputeMerkleRoot(items[getSplitPoint(uint64(len(items))):])
		return innerHash(left, right)
	}
}

// ProofsFromByteSlices ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ProofsFromByteSlices 方法接受一个区块的所有交易数据，用[][]byte类型的items存储这些交易数据，
// 该方法的目的就是根据这些交易数据构建默克尔树，然后返回每个交易数据的 Proof 和默克尔树的根哈希值。
func ProofsFromByteSlices(items [][]byte) (rootHash []byte, proofs []*Proof) {
	trails, root := trailsFromByteSlices(items)
	rootHash = root.Hash
	proofs = make([]*Proof, len(trails))
	for i, trail := range trails {
		proofs[i] = new(Proof)
		proofs[i].Index = uint64(i)
		proofs[i].Total = uint64(len(trails))
		proofs[i].LeafHash = trail.Hash
		proofs[i].Aunts = trail.flattenAunts()
	}
	return rootHash, proofs
}

// ProofFromProto ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ProofFromProto 方法接受一个 *pbcrypto.Proof 实例，然后将其转换为 *Proof。
func ProofFromProto(pb *pbcrypto.Proof) (*Proof, error) {
	if pb == nil {
		return nil, errors.New("nil Proof")
	}
	p := &Proof{
		Total:    pb.Total,
		Index:    pb.Index,
		LeafHash: pb.LeafHash,
		Aunts:    pb.Aunts,
	}
	return p, p.ValidateBasic()
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

// Proof ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Proof 结构体由四个字段组成，分别是：
//   - Total：
//   - Index：
//   - LeafHash：叶子节点的哈希值（某条交易数据的哈希值）
//   - Aunts：只要给定Aunts和LeafHash，我就可以求得默克尔树的根哈希，Aunts递归地存储了节点的aunt节点的哈希值
//
// Proof是一个用来验证交易数据是否被篡改的结构体。
type Proof struct {
	Total    uint64 `json:"total"`
	Index    uint64
	LeafHash []byte
	Aunts    [][]byte
}

// ComputeRootHash ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ComputeRootHash 方法根据叶子节点的aunts们的哈希值和自身的哈希值，计算默克尔树的根哈希值，
// 如果该叶子节点存储的aunts有误，或者该叶子节点是空的，则返回nil。
func (p *Proof) ComputeRootHash() []byte {
	var root []byte = p.LeafHash
	if len(p.Aunts) == 0 && p.Total == 1 {
		return p.LeafHash
	}
	for _, aunt := range p.Aunts {
		switch aunt[0] {
		case 'l':
			root = innerHash(aunt[1:], root)
		case 'r':
			root = innerHash(root, aunt[1:])
		default:
			return nil
		}
	}
	return root
}

// Verify ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Verify 方法接受两个参数，第一个参数是区块里存储的默克尔树的根哈希值，第二个参数是区块中的某条交易数据，
// 该方法先比较交易数据的哈希值是否与区块中默克尔树的对应叶子节点的哈希值一样，如果不一样，则表明区块里交易
// 数据的哈希值遭到篡改，一样的话继续比较默克尔树的根哈希值是否与给定的根哈希值一样，如果不一样，则说明区块
// 里的默克尔树的根哈希值遭到了篡改。
func (p *Proof) Verify(root, item []byte) error {
	h := leafHash(item)
	if !bytes.Equal(p.LeafHash, h) {
		return fmt.Errorf("invalid tx hash: wanted %x, got %X", h, p.LeafHash)
	}
	computedRoot := p.ComputeRootHash()
	if !bytes.Equal(computedRoot, root) {
		return fmt.Errorf("invalid merkle root hash, wanted %X, got %X", root, computedRoot)
	}
	return nil
}

// ValidateBasic ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ValidateBasic 方法用于实现 Message 接口，该接口定义在consensus包里。
func (p *Proof) ValidateBasic() error {
	if len(p.LeafHash) != sha256.Size32 {
		return fmt.Errorf("expected hash size to be %d, got %d", sha256.Size32, len(p.LeafHash))
	}
	if len(p.Aunts) > maxAunts {
		return fmt.Errorf("expected no more than %d aunts, got %d", maxAunts, len(p.Aunts))
	}
	return nil
}

// ToProto ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToProto 方法将 Proof 转换为protobuf里的 pbcrypto.Proof。
func (p *Proof) ToProto() *pbcrypto.Proof {
	if p == nil {
		return nil
	}
	return &pbcrypto.Proof{
		Total:    p.Total,
		Index:    p.Index,
		LeafHash: p.LeafHash,
		Aunts:    p.Aunts,
	}
}

// String ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// String 打印 Proof 的字符串格式。
func (p *Proof) String() string {
	return fmt.Sprintf("Proof{\n\tHash: %X\n\tAunts: %d\n}", p.LeafHash, len(p.Aunts))
}

// ProofNode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ProofNode 结构体由四个字段组成，分别是：
//   - Hash：
//   - Parent：在默克尔树里的父节点
//   - Left：在默克尔树里的左兄弟
//   - Right：在默克尔树里的右兄弟
type ProofNode struct {
	Hash   []byte
	Parent *ProofNode
	Left   *ProofNode
	Right  *ProofNode
}

func (pn *ProofNode) flattenAunts() [][]byte {
	var innerHashes [][]byte
	for pn != nil {
		switch {
		case pn.Left != nil:
			innerHashes = append(innerHashes, append([]byte{'l'}, pn.Left.Hash...))
		case pn.Right != nil:
			innerHashes = append(innerHashes, append([]byte{'r'}, pn.Right.Hash...))
		default:
			break
		}
		pn = pn.Parent
	}
	return innerHashes
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数和变量

// maxAunts ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// maxAunts 限制了默克尔树里每个叶子节点最多可以有多少个aunt，这也从侧面反映出一棵
// 默克尔树最高不能超过 maxAunts。
const maxAunts = 100

// trailsFromByteSlices ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// trailsFromByteSlices 方法接受一个[][]byte类型的切片items作为输入参数，items存储的是区块里的原始交易数据，
// 该方法就是根据给定的若干交易数据构造一颗完整的默克尔树，从左到右计算默克尔树左右叶子节点的哈希值，并将这些哈希值
// 做为第一个返回值返回，第二个返回值返回的是这棵默克尔树的根哈希值。
func trailsFromByteSlices(items [][]byte) (trails []*ProofNode, root *ProofNode) {
	switch len(items) {
	case 0:
		return []*ProofNode{}, &ProofNode{Hash: emptyHash(), Left: nil, Right: nil, Parent: nil}
	case 1:
		trail := &ProofNode{Hash: leafHash(items[0]), Left: nil, Right: nil, Parent: nil}
		return []*ProofNode{trail}, trail
	default:
		split := getSplitPoint(uint64(len(items)))
		lefts, leftRoot := trailsFromByteSlices(items[:split])
		rights, rightRoot := trailsFromByteSlices(items[split:])
		root = &ProofNode{Hash: innerHash(leftRoot.Hash, rightRoot.Hash), Left: nil, Right: nil}
		leftRoot.Right = rightRoot
		rightRoot.Left = leftRoot
		leftRoot.Parent = root
		rightRoot.Parent = root
		trails = append(lefts, rights...)
		return trails, root
	}
}

// getSplitPoint ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// getSplitPoint 方法接受一个整数length，然后计算小于length的最大的2的n次幂，下面给出几个例子：
//
//	输入：4	输出：2
//	输入：8	输出：4
//	输入：9	输出：8
//	输入：20	输出：16
//
// 这个方法在把默克尔树的所有叶子节点分成左右两份时被调用，用来确定在何处进行分割。
func getSplitPoint(length uint64) uint64 {
	if length < 1 {
		panic("trying to split a tree with size < 1")
	}
	bitsLen := bits.Len64(length)
	k := uint64(1 << (bitsLen - 1))
	if k == length {
		k >>= 1
	}
	return k
}
