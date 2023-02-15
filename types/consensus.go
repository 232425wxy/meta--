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
	if len(block.ChameleonHash.Hash) == 0 {
		panic("block's hash is empty")
	}
	hash := make([]byte, len(block.ChameleonHash.Hash))
	copy(hash[:], block.ChameleonHash.Hash)
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

func NewPrepareVote(height int64, hash []byte, privateKey *bls12.PrivateKey) *PrepareVote {
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
	ValueHash          []byte                       `json:"value_hash"`
	Timestamp          time.Time                    `json:"timestamp"`
	AggregateSignature *bls12.AggregateSignature    `json:"aggregate_signature"` // 这个签名是对PrepareVote消息的聚合签名
}

func NewPreCommit(agg *bls12.AggregateSignature, hash []byte, id crypto.ID, height int64) *PreCommit {
	return &PreCommit{
		Type:               pbtypes.PreCommitType,
		ID:                 id,
		Height:             height,
		ValueHash:          hash,
		Timestamp:          time.Now(),
		AggregateSignature: agg,
	}
}

func (pc *PreCommit) ToProto() *pbtypes.PreCommit {
	if pc == nil {
		return nil
	}
	return &pbtypes.PreCommit{
		Type:               pc.Type,
		ID:                 string(pc.ID),
		Height:             pc.Height,
		ValueHash:          pc.ValueHash[:],
		Timestamp:          pc.Timestamp,
		AggregateSignature: pc.AggregateSignature.ToProto(),
	}
}

func PreCommitFromProto(pb *pbtypes.PreCommit) *PreCommit {
	if pb == nil {
		return nil
	}
	hash := make([]byte, len(pb.ValueHash))
	copy(hash[:], pb.ValueHash)
	return &PreCommit{
		Type:               pb.Type,
		ID:                 crypto.ID(pb.ID),
		Height:             pb.Height,
		ValueHash:          hash,
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

func NewPreCommitVote(height int64, hash []byte, privateKey *bls12.PrivateKey) *PreCommitVote {
	vote := &PreCommitVote{Vote: &Vote{
		VoteType:  pbtypes.PreCommitVoteType,
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
	ValueHash          []byte                       `json:"value_hash"`
	Timestamp          time.Time                    `json:"timestamp"`
	AggregateSignature *bls12.AggregateSignature    `json:"aggregate_signature"`
}

func NewCommit(agg *bls12.AggregateSignature, hash []byte, id crypto.ID, height int64) *Commit {
	return &Commit{
		Type:               pbtypes.CommitType,
		ID:                 id,
		Height:             height,
		ValueHash:          hash,
		Timestamp:          time.Now(),
		AggregateSignature: agg,
	}
}

func (c *Commit) ToProto() *pbtypes.Commit {
	if c == nil {
		return nil
	}
	return &pbtypes.Commit{
		Type:               c.Type,
		ID:                 string(c.ID),
		Height:             c.Height,
		ValueHash:          c.ValueHash[:],
		Timestamp:          c.Timestamp,
		AggregateSignature: c.AggregateSignature.ToProto(),
	}
}

func CommitFromProto(pb *pbtypes.Commit) *Commit {
	if pb == nil {
		return nil
	}
	hash := make([]byte, len(pb.ValueHash))
	copy(hash[:], pb.ValueHash)
	return &Commit{
		Type:               pb.Type,
		ID:                 crypto.ID(pb.ID),
		Height:             pb.Height,
		ValueHash:          hash,
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

func NewCommitVote(height int64, hash []byte, privateKey *bls12.PrivateKey) *CommitVote {
	vote := &CommitVote{Vote: &Vote{
		VoteType:  pbtypes.CommitVoteType,
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
	ValueHash          []byte                       `json:"value_hash"`
	Timestamp          time.Time                    `json:"timestamp"`
	AggregateSignature *bls12.AggregateSignature
}

func NewDecide(agg *bls12.AggregateSignature, hash []byte, id crypto.ID, height int64) *Decide {
	return &Decide{
		Type:               pbtypes.DecideType,
		ID:                 id,
		Height:             height,
		ValueHash:          hash,
		Timestamp:          time.Now(),
		AggregateSignature: agg,
	}
}

func (d *Decide) ToProto() *pbtypes.Decide {
	if d == nil {
		return nil
	}
	return &pbtypes.Decide{
		Type:               d.Type,
		ID:                 string(d.ID),
		Height:             d.Height,
		ValueHash:          d.ValueHash[:],
		Timestamp:          d.Timestamp,
		AggregateSignature: d.AggregateSignature.ToProto(),
	}
}

func DecideFromProto(pb *pbtypes.Decide) *Decide {
	if pb == nil {
		return nil
	}
	hash := make([]byte, len(pb.ValueHash))
	copy(hash[:], pb.ValueHash)
	return &Decide{
		Type:               pb.Type,
		ID:                 crypto.ID(pb.ID),
		Height:             pb.Height,
		ValueHash:          hash,
		Timestamp:          pb.Timestamp,
		AggregateSignature: bls12.AggregateSignatureFromProto(pb.AggregateSignature),
	}
}

func (d *Decide) ValidateBasic() error {
	return nil
}

func GeneratePrepareVoteValueHash(blockHash []byte) []byte {
	value := append([]byte("PrepareVote-"), blockHash...)
	h := sha256.Sum(value)
	return h[:]
}

func GeneratePreCommitValueHash(blockHash []byte) []byte {
	return GeneratePrepareVoteValueHash(blockHash)
}

func GeneratePreCommitVoteValueHash(blockHash []byte) []byte {
	value := append([]byte("PreCommitVote-"), blockHash...)
	h := sha256.Sum(value)
	return h[:]
}

func GenerateCommitValueHash(blockHash []byte) []byte {
	return GeneratePreCommitVoteValueHash(blockHash)
}

func GenerateCommitVoteValueHash(blockHash []byte) []byte {
	value := append([]byte("CommitVote-"), blockHash...)
	h := sha256.Sum(value)
	return h[:]
}

func GenerateDecideValueHash(blockHash []byte) []byte {
	return GenerateCommitVoteValueHash(blockHash)
}
