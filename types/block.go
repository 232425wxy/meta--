package types

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/crypto/merkle"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"math/big"
	"time"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 区块

type ChameleonHash struct {
	R1 *big.Int
	R2 *big.Int

	Alpha *big.Int
	Hash  []byte
}

func (ch *ChameleonHash) ToProto() *pbtypes.ChameleonHash {
	if ch == nil {
		return nil
	}
	return &pbtypes.ChameleonHash{
		GSigma:  ch.R1.Bytes(),
		HKSigma: ch.R2.Bytes(),
		Alpha:   ch.Alpha.Bytes(),
		Hash:    ch.Hash,
	}
}

func ChameleonHashFromProto(pb *pbtypes.ChameleonHash) *ChameleonHash {
	if pb == nil {
		return nil
	}
	return &ChameleonHash{
		R1:    new(big.Int).SetBytes(pb.GSigma),
		R2:    new(big.Int).SetBytes(pb.HKSigma),
		Alpha: new(big.Int).SetBytes(pb.Alpha),
		Hash:  pb.Hash,
	}
}

type Block struct {
	Header        *Header        `json:"header"`
	Body          *Data          `json:"body"`
	ChameleonHash *ChameleonHash `json:"chameleon_hash"`
}

func (b *Block) Copy() *Block {
	if b == nil {
		return nil
	}
	cp := &Block{
		Header: &Header{
			PreviousBlockHash: b.Header.PreviousBlockHash,
			BlockDataHash:     b.Header.BlockDataHash,
			Height:            b.Header.Height,
			Timestamp:         b.Header.Timestamp,
			Proposer:          b.Header.Proposer,
		},
		Body: &Data{
			RootHash: b.Body.RootHash,
			Txs:      make(Txs, len(b.Body.Txs)),
		},
		ChameleonHash: &ChameleonHash{
			R1:    new(big.Int).Set(b.ChameleonHash.R1),
			R2:    new(big.Int).Set(b.ChameleonHash.R2),
			Alpha: new(big.Int).Set(b.ChameleonHash.Alpha),
			Hash:  b.ChameleonHash.Hash,
		},
	}
	for i, tx := range b.Body.Txs {
		cp.Body.Txs[i] = tx
	}

	return cp
}

func (b *Block) ValidateBasic() error {
	if b == nil {
		return errors.New("nil block")
	}
	if err := b.Body.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

// BlockDataHash
// 计算区块的哈希值
// TODO 将来换成变色龙哈希
func (b *Block) BlockDataHash() []byte {
	h := sha256.New()
	h.Write(b.Header.PreviousBlockHash)
	h.Write([]byte(fmt.Sprintf("%d", b.Header.Height)))
	//h.Write([]byte(b.Header.Timestamp.String()))
	h.Write([]byte(b.Header.Proposer))
	_txs := make([][]byte, len(b.Body.Txs))
	for i, tx := range b.Body.Txs {
		_txs[i] = tx
	}
	b.Body.RootHash = merkle.ComputeMerkleRoot(_txs)
	h.Write(b.Body.RootHash)
	b.Header.BlockDataHash = h.Sum(nil)
	return b.Header.BlockDataHash
}

func (b *Block) ToProto() *pbtypes.Block {
	// 不包括对当前区块的投票决定
	if b == nil {
		return nil
	}
	pb := &pbtypes.Block{
		Header:        b.Header.ToProto(),
		Body:          b.Body.ToProto(),
		ChameleonHash: b.ChameleonHash.ToProto(),
	}
	return pb
}

func BlockFromProto(pb *pbtypes.Block) *Block {
	if pb == nil {
		return nil
	}
	return &Block{
		Header:        HeaderFromProto(pb.Header),
		Body:          DataFromProto(pb.Body),
		ChameleonHash: ChameleonHashFromProto(pb.ChameleonHash),
	}
}

func (b *Block) String() string {
	if b == nil {
		return "Block{nil}"
	}
	h := sha256.New()
	h.Write(b.Header.PreviousBlockHash)
	previousBlockHash := h.Sum(nil)
	h.Reset()
	h.Write(b.ChameleonHash.Hash)
	hash := h.Sum(nil)
	str := fmt.Sprintf("Block{\n\tHeader{\n\t\tPreviousBlockHash: %x\n\t\tBlockDataHash: %x\n\t\tHeight: %d\n\t\tTimestamp: %s\n\t\tProposer: %s\n\t}\n\tBody{\n\t\tRootHash: %x\n\t\tTxs: %s\n\t}\n\tChameleonHash{\n\t\tR1: %x\n\t\tR2: %x\n\t\tAlpha: %x\n\t\tHash: %x\n\t}\n}",
		previousBlockHash, b.Header.BlockDataHash, b.Header.Height, b.Header.Timestamp.Format(time.RFC3339), b.Header.Proposer, b.Body.RootHash, b.Body.Txs, b.ChameleonHash.R1.Bytes(), b.ChameleonHash.R2.Bytes(), b.ChameleonHash.Alpha.Bytes(), hash)
	return str
}

type BlockHeight struct {
	Height int64 `json:"height"`
}

func (bh *BlockHeight) ToProto() *pbtypes.BlockHeight {
	if bh == nil {
		return nil
	}
	return &pbtypes.BlockHeight{Height: bh.Height}
}

func BlockHeightFromProto(pb *pbtypes.BlockHeight) *BlockHeight {
	if pb == nil {
		return nil
	}
	return &BlockHeight{Height: pb.Height}
}

type CommitBlock struct {
	Height             int64                     `json:"height"`
	Hash               []byte                    `json:"hash"`
	AggregateSignature *bls12.AggregateSignature `json:"aggregate_signature"`
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 区块头

type Header struct {
	PreviousBlockHash []byte    `json:"previous_block_hash"`
	BlockDataHash     []byte    `json:"block_data_hash"`
	Height            int64     `json:"height"`
	Timestamp         time.Time `json:"timestamp"`
	Proposer          crypto.ID `json:"proposer"`
}

func (h *Header) ToProto() *pbtypes.Header {
	if h == nil {
		return nil
	}
	return &pbtypes.Header{
		PreviousBlockHash: h.PreviousBlockHash,
		BlockDataHash:     h.BlockDataHash,
		Height:            h.Height,
		Timestamp:         h.Timestamp,
		Proposer:          string(h.Proposer),
	}
}

func HeaderFromProto(pb *pbtypes.Header) *Header {
	if pb == nil {
		return nil
	}
	return &Header{
		PreviousBlockHash: pb.PreviousBlockHash,
		BlockDataHash:     pb.BlockDataHash,
		Height:            pb.Height,
		Timestamp:         pb.Timestamp.Local(),
		Proposer:          crypto.ID(pb.Proposer),
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 区块体

type Data struct {
	RootHash []byte `json:"root_hash"`
	Txs      Txs    `json:"txs"`
}

func (d *Data) ToProto() *pbtypes.Data {
	if d == nil {
		return nil
	}
	_txs := make([][]byte, len(d.Txs))
	for i, tx := range d.Txs {
		_txs[i] = tx
	}
	return &pbtypes.Data{
		RootHash: d.RootHash,
		Txs:      _txs,
	}
}

func DataFromProto(pb *pbtypes.Data) *Data {
	if pb == nil {
		return nil
	}
	txs := make(Txs, len(pb.Txs))
	for i := 0; i < len(pb.Txs); i++ {
		txs[i] = pb.Txs[i]
	}
	return &Data{
		RootHash: pb.RootHash,
		Txs:      txs,
	}
}

// ValidateBasic ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// ValidateBasic 方法验证区块体部分的交易数据大小不能超过1MB。
func (d *Data) ValidateBasic() error {
	size := 0
	for _, tx := range d.Txs {
		size += len(tx)
	}
	if size > 1024*1024 {
		return fmt.Errorf("exceed data limit: %d > %d", size, 1024*1024)
	}
	return nil
}
