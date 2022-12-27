package types

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/crypto/sha256"
	"time"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 区块

type SimpleBlock struct {
	Hash []byte
}

type Block struct {
	LastBlock SimpleBlock `json:"lastBlock"` // 上个区块的哈希值
	Header    Header      `json:"header"`
	Data      Data        `json:"data"`
	Decision  Decision    `json:"decision"`
}

func (b *Block) ValidateBasic() error {
	if b == nil {
		return errors.New("nil block")
	}
	if err := b.Data.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func (b *Block) Hash() []byte {
	h := sha256.New()
	h.Write(b.LastBlock.Hash)
	h.Write([]byte(fmt.Sprintf("%d", b.Header.Height)))
	h.Write([]byte(b.Header.Timestamp.String()))
	h.Write([]byte(b.Header.Proposer))
	h.Write(b.Data.RootHash)
	b.Header.Hash = h.Sum(nil)
	return b.Header.Hash
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 区块头

type Header struct {
	Hash      []byte    `json:"hash"` // 当前区块哈希
	Height    int64     `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	Proposer  crypto.ID `json:"proposer"`
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 区块体

type Data struct {
	RootHash []byte `json:"rootHash"`
	Txs      Txs    `json:"txs"`
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
