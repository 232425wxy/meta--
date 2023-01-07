package types

import (
	"errors"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"time"
)

type NextView struct {
	Type   pbtypes.ConsensusMessageType `json:"type"`
	ID     crypto.ID                    `json:"ID"`
	Height int64                        `json:"height"`
}

/**********************************************************************************************************************/

type Prepare struct {
	Type      pbtypes.ConsensusMessageType `json:"type"`
	ID        crypto.ID                    `json:"ID"`
	Height    int64                        `json:"height"`
	Block     *Block                       `json:"block"`
	Timestamp time.Time                    `json:"timestamp"`
	Signature *bls12.Signature             `json:"signature"`
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

func PrepareFromProto(pb *pbtypes.Prepare) *Prepare {
	if pb == nil {
		return nil
	}
	return &Prepare{
		Type:      pbty,
		ID:        "",
		Height:    0,
		Block:     nil,
		Timestamp: time.Time{},
		Signature: nil,
	}
}

type PrepareVote struct {
	Vote *Vote `json:"vote"`
}

/**********************************************************************************************************************/

type PreCommit struct {
	Type               pbtypes.ConsensusMessageType `json:"type"`
	ID                 crypto.ID                    `json:"ID"`
	Height             int64                        `json:"height"`
	PrepareHash        sha256.Hash                  `json:"prepare_hash"` // 这个字段的值等于 Hash("Prepare"||ValueHash)
	Timestamp          time.Time                    `json:"timestamp"`
	AggregateSignature *bls12.AggregateSignature    `json:"aggregate_signature"`
}

type PreCommitVote struct {
	Vote *Vote `json:"vote"`
}

/**********************************************************************************************************************/

type Commit struct {
	Type               pbtypes.ConsensusMessageType `json:"type"`
	ID                 crypto.ID                    `json:"ID"`
	Height             int64                        `json:"height"`
	PreCommitHash      sha256.Hash                  `json:"block_hash"`
	Timestamp          time.Time                    `json:"timestamp"`
	AggregateSignature *bls12.AggregateSignature    `json:"aggregate_signature"`
}

type CommitVote struct {
	Vote *Vote `json:"vote"`
}

/**********************************************************************************************************************/

type Decide struct {
	Type               pbtypes.ConsensusMessageType `json:"type"`
	ID                 crypto.ID                    `json:"ID"`
	Height             int64                        `json:"height"`
	CommitHash         sha256.Hash                  `json:"commit_hash"`
	Timestamp          time.Time                    `json:"timestamp"`
	AggregateSignature *bls12.AggregateSignature
}
