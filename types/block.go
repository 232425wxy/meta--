package types

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"time"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 区块

type Block struct {
	Header *Header `json:"header"`
	Body   *Data   `json:"body"`
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

func (b *Block) Hash() []byte {
	h := sha256.New()
	h.Write(b.Header.PreviousBlockHash)
	h.Write([]byte(fmt.Sprintf("%d", b.Header.Height)))
	h.Write([]byte(b.Header.Timestamp.String()))
	h.Write([]byte(b.Header.Proposer))
	h.Write(b.Body.RootHash)
	b.Header.Hash = h.Sum(nil)
	return b.Header.Hash
}

func (b *Block) ToProto() *pbtypes.Block {
	// 不包括对当前区块的投票决定
	if b == nil {
		return nil
	}
	pb := &pbtypes.Block{
		Header: b.Header.ToProto(),
		Body:   b.Body.ToProto(),
	}
	return pb
}

func BlockFromProto(pb *pbtypes.Block) *Block {
	if pb == nil {
		return nil
	}
	return &Block{
		Header: HeaderFromProto(pb.Header),
		Body:   DataFromProto(pb.Body),
	}
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
	Hash              []byte    `json:"hash"` // 当前区块哈希
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
		Hash:              h.Hash,
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
		Hash:              pb.Hash,
		Height:            pb.Height,
		Timestamp:         pb.Timestamp,
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
		copy(_txs[i], tx)
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

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 共识投票

type Decision struct {
	Signature *bls12.AggregateSignature
}

func DecisionFromProto(pb *pbtypes.Decision) *Decision {
	if pb == nil {
		return nil
	}
	return &Decision{Signature: bls12.AggregateSignatureFromProto(pb.Signature)}
}
