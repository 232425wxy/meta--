package types

import (
	"errors"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"time"
)

type NextView struct {
	Height int64     `json:"height"`
	ID     crypto.ID `json:"ID"`
}

type Prepare struct {
	ID        crypto.ID        `json:"ID"`
	Height    int64            `json:"height"`
	Block     *Block           `json:"block"`
	Timestamp time.Time        `json:"timestamp"`
	Signature *bls12.Signature `json:"signature"`
}

func NewPrepare(height int64, block *Block) *Prepare {
	return &Prepare{
		Height:    height,
		Block:     block,
		Timestamp: time.Now(),
	}
}

func (p *Prepare) ValidateBasic() error {
	if p.Height < 0 {
		return errors.New("negative height")
	}
	return nil
}

func (p *Prepare) ToProto() *pbtypes.Prepare {
	return &pbtypes.Prepare{
		Height:    p.Height,
		Block:     p.Block.ToProto(),
		Timestamp: p.Timestamp,
		Signature: p.Signature.ToProto(),
	}
}

type PrepareVote struct {
	Vote *Vote `json:"vote"`
}

/**********************************************************************************************************************/

type PreCommit struct {
	ID                 crypto.ID                 `json:"ID"`
	Height             int64                     `json:"height"`
	BlockHash          []byte                    `json:"block_hash"`
	Timestamp          time.Time                 `json:"timestamp"`
	AggregateSignature *bls12.AggregateSignature `json:"aggregate_signature"`
}

type PreCommitVote struct {
	Vote *Vote `json:"vote"`
}

/**********************************************************************************************************************/

type Commit struct {
	Vote *Vote `json:"vote"`
}

type CommitVote struct {
}

type Decide struct {
}
