package types

import (
	"errors"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"github.com/cosmos/gogoproto/proto"
	"time"
)

type NextView struct {
	Type   pbtypes.ConsensusMessageType `json:"type"`
	ID     crypto.ID                    `json:"ID"`
	Height int64                        `json:"height"`
}

func (nv *NextView) ToProto() *pbtypes.NextView {
	if nv == nil {
		return nil
	}
	return &pbtypes.NextView{
		Type:   nv.Type,
		ID:     string(nv.ID),
		Height: nv.Height,
	}
}

func NextViewFromProto(pb *pbtypes.NextView) *NextView {
	if pb == nil {
		return nil
	}
	return &NextView{
		Type:   pb.Type,
		ID:     crypto.ID(pb.ID),
		Height: pb.Height,
	}
}

func (nv *NextView) ValidateBasic() error {
	return nil
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

func NewPrepare(height int64, block *Block, id crypto.ID, privateKey *bls12.PrivateKey) *Prepare {
	p := &Prepare{
		Type:      pbtypes.PrepareType,
		ID:        id,
		Height:    height,
		Block:     block,
		Timestamp: time.Now(),
	}
	if len(block.Header.Hash) == 0 {
		panic("block's hash is empty")
	}
	hash := sha256.Hash{}
	copy(hash[:], block.Header.Hash)
	var err error
	p.Signature, err = privateKey.Sign(hash)
	if err != nil {
		panic(err)
	}
	return p
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
		Type:      pb.Type,
		ID:        crypto.ID(pb.ID),
		Height:    pb.Height,
		Block:     BlockFromProto(pb.Block),
		Timestamp: pb.Timestamp,
		Signature: bls12.SignatureFromProto(pb.Signature),
	}
}

func (p *Prepare) Hash() sha256.Hash {
	pb := p.ToProto()
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return sha256.Sum(bz)
}

type PrepareVote struct {
	Vote *Vote `json:"vote"`
}

func NewPrepareVote(height int64, hash sha256.Hash, privateKey *bls12.PrivateKey) *PrepareVote {
	vote := &PrepareVote{Vote: &Vote{
		VoteType:  pbtypes.PrepareVoteType,
		Height:    height,
		ValueHash: hash,
		Timestamp: time.Now(),
	}}
	var err error
	vote.Vote.Signature, err = privateKey.Sign(vote.Vote.ValueHash)
	if err != nil {
		panic(err)
	}
	return vote
}

func (pv *PrepareVote) ToProto() *pbtypes.PrepareVote {
	if pv == nil {
		return nil
	}
	return &pbtypes.PrepareVote{Vote: pv.Vote.ToProto()}
}

func PrepareVoteFromProto(pb *pbtypes.PrepareVote) *PrepareVote {
	if pb == nil {
		return nil
	}
	return &PrepareVote{Vote: VoteFromProto(pb.Vote)}
}

func (pv *PrepareVote) ValidateBasic() error {
	return nil
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

func (pc *PreCommit) ToProto() *pbtypes.PreCommit {
	if pc == nil {
		return nil
	}
	return &pbtypes.PreCommit{
		Type:               pc.Type,
		ID:                 string(pc.ID),
		Height:             pc.Height,
		PrepareHash:        pc.PrepareHash[:],
		Timestamp:          pc.Timestamp,
		AggregateSignature: pc.AggregateSignature.ToProto(),
	}
}

func PreCommitFromProto(pb *pbtypes.PreCommit) *PreCommit {
	if pb == nil {
		return nil
	}
	hash := sha256.Hash{}
	copy(hash[:], pb.PrepareHash)
	return &PreCommit{
		Type:               pb.Type,
		ID:                 crypto.ID(pb.ID),
		Height:             pb.Height,
		PrepareHash:        hash,
		Timestamp:          pb.Timestamp,
		AggregateSignature: bls12.AggregateSignatureFromProto(pb.AggregateSignature),
	}
}

func (pc *PreCommit) ValidateBasic() error {
	return nil
}

type PreCommitVote struct {
	Vote *Vote `json:"vote"`
}

func (pcv *PreCommitVote) ToProto() *pbtypes.PreCommitVote {
	if pcv == nil {
		return nil
	}
	return &pbtypes.PreCommitVote{Vote: pcv.Vote.ToProto()}
}

func PreCommitVoteFromProto(pb *pbtypes.PreCommitVote) *PreCommitVote {
	if pb == nil {
		return nil
	}
	return &PreCommitVote{Vote: VoteFromProto(pb.Vote)}
}

func (pcv *PreCommitVote) ValidateBasic() error {
	return nil
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

func (c *Commit) ToProto() *pbtypes.Commit {
	if c == nil {
		return nil
	}
	return &pbtypes.Commit{
		Type:               c.Type,
		ID:                 string(c.ID),
		Height:             c.Height,
		PreCommitHash:      c.PreCommitHash[:],
		Timestamp:          c.Timestamp,
		AggregateSignature: c.AggregateSignature.ToProto(),
	}
}

func CommitFromProto(pb *pbtypes.Commit) *Commit {
	if pb == nil {
		return nil
	}
	hash := sha256.Hash{}
	copy(hash[:], pb.PreCommitHash)
	return &Commit{
		Type:               pb.Type,
		ID:                 crypto.ID(pb.ID),
		Height:             pb.Height,
		PreCommitHash:      hash,
		Timestamp:          pb.Timestamp,
		AggregateSignature: bls12.AggregateSignatureFromProto(pb.AggregateSignature),
	}
}

func (c *Commit) ValidateBasic() error {
	return nil
}

type CommitVote struct {
	Vote *Vote `json:"vote"`
}

func (cv *CommitVote) ToProto() *pbtypes.CommitVote {
	if cv == nil {
		return nil
	}
	return &pbtypes.CommitVote{Vote: cv.Vote.ToProto()}
}

func CommitVoteFromProto(pb *pbtypes.CommitVote) *CommitVote {
	if pb == nil {
		return nil
	}
	return &CommitVote{Vote: VoteFromProto(pb.Vote)}
}

func (cv *CommitVote) ValidateBasic() error {
	return nil
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

func (d *Decide) ToProto() *pbtypes.Decide {
	if d == nil {
		return nil
	}
	return &pbtypes.Decide{
		Type:               d.Type,
		ID:                 string(d.ID),
		Height:             d.Height,
		CommitHash:         d.CommitHash[:],
		Timestamp:          d.Timestamp,
		AggregateSignature: d.AggregateSignature.ToProto(),
	}
}

func DecideFromProto(pb *pbtypes.Decide) *Decide {
	if pb == nil {
		return nil
	}
	hash := sha256.Hash{}
	copy(hash[:], pb.CommitHash)
	return &Decide{
		Type:               pb.Type,
		ID:                 crypto.ID(pb.ID),
		Height:             pb.Height,
		CommitHash:         hash,
		Timestamp:          pb.Timestamp,
		AggregateSignature: bls12.AggregateSignatureFromProto(pb.AggregateSignature),
	}
}

func (d *Decide) ValidateBasic() error {
	return nil
}
