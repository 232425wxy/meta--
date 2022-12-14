package types

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/crypto/merkle"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/proto/pbtypes"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义交易结构体

// Tx ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Tx 区块中的交易数据，由任意的字节切片构成。
type Tx []byte

// Hash ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Hash 计算并返回交易数据的sha256哈希值。
func (t Tx) Hash() []byte {
	h := sha256.Sum(t)
	return h[:]
}

func (t Tx) String() string {
	return fmt.Sprintf("Tx{%x}", []byte(t))
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义交易集合

// Txs ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Txs 定义了由若干个交易构成的集合，可以作为区块中的交易字段。
type Txs []Tx

// MerkleRootHash ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// MerkleRootHash 方法计算若干个交易的默克尔根哈希值。
func (txs Txs) MerkleRootHash() []byte {
	hashes := make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		hashes[i] = txs[i].Hash()
	}
	return merkle.ComputeMerkleRoot(hashes)
}

// Proof ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Proof 给定一个区块里的所有交易数据，然后计算这群交易的默克尔根哈希值，以及每个交易的 merkle.Proof，
// 返回指定交易的 TxProof。
func (txs Txs) Proof(index int) TxProof {
	length := len(txs)
	hashes := make([][]byte, length)
	for i := 0; i < length; i++ {
		hashes[i] = txs[i].Hash()
	}
	root, proofs := merkle.ProofsFromByteSlices(hashes)
	return TxProof{
		MerkleRootHash: root,
		Data:           txs[index],
		Proof:          *proofs[index],
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义交易在默克尔树中的证明结构体

type TxProof struct {
	MerkleRootHash []byte       `json:"merkle_root_hash"`
	Data           Tx           `json:"data"`
	Proof          merkle.Proof `json:"proof"`
}

func (tp TxProof) Validate(rootHash []byte) error {
	if !bytes.Equal(tp.MerkleRootHash, rootHash) {
		return errors.New("proof matches different merkle root hash")
	}
	if err := tp.Proof.Verify(tp.MerkleRootHash, tp.Data.Hash()); err != nil {
		return err
	}
	return nil
}

func (tp TxProof) ToProto() pbtypes.TxProof {
	return pbtypes.TxProof{
		MerkleRootHash: tp.MerkleRootHash,
		Data:           tp.Data,
		Proof:          tp.Proof.ToProto(),
	}
}

func TxProofFromProto(pb pbtypes.TxProof) (TxProof, error) {
	proof, err := merkle.ProofFromProto(pb.Proof)
	if err != nil {
		return TxProof{}, err
	}
	return TxProof{
		MerkleRootHash: pb.MerkleRootHash,
		Data:           pb.Data,
		Proof:          *proof,
	}, nil
}

// ComputeProtoSizeForTxs ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ComputeProtoSizeForTxs 方法将一众交易数据包装成区块中的交易字段，包括交易数据的默克尔根哈希，然后计算
// 交易字段的大小并返回。
func ComputeProtoSizeForTxs(txs []Tx) int64 {
	data := Data{Txs: txs}
	pbData := data.ToProto()
	return int64(pbData.Size())
}
