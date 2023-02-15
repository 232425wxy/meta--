package types

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"time"
)

// Vote 对不同共识阶段里的消息进行投票，投票需要节点的签名，对投票中的ValueHash进行签名。
type Vote struct {
	VoteType  pbtypes.VoteType
	Height    int64
	ValueHash []byte
	Timestamp time.Time
	Signature *bls12.Signature
}

func NewVote(typ pbtypes.VoteType, height int64, valueHash []byte, privateKey *bls12.PrivateKey) *Vote {
	v := &Vote{
		VoteType:  typ,
		Height:    height,
		ValueHash: valueHash,
		Timestamp: time.Now(),
	}
	v.sign(privateKey)
	return v
}

func (v *Vote) Verify() error {
	if v == nil {
		return fmt.Errorf("vote is nil")
	}
	publicKey := bls12.GetBLSPublicKeyFromLib(v.Signature.Signer())
	if publicKey == nil {
		return fmt.Errorf("cannot find the public key corresponding to %v", v.Signature.Signer())
	}
	ok := publicKey.Verify(v.Signature, v.ValueHash)
	if ok {
		return nil
	}
	return fmt.Errorf("invalid signature for %s", v.VoteType)
}

func (v *Vote) sign(privateKey *bls12.PrivateKey) {
	if len(v.ValueHash[:]) == 0 {
		panic("cannot sign nil message")
	}
	sig, err := privateKey.Sign(v.ValueHash)
	if err != nil {
		panic(err)
	}
	v.Signature = sig
}

func (v *Vote) ValidateBasic() error {
	if v.Height < 0 {
		return errors.New("negative height")
	}
	return nil
}

func (v *Vote) ToProto() *pbtypes.Vote {
	if v == nil {
		return nil
	}
	return &pbtypes.Vote{
		VoteType:  v.VoteType,
		Height:    v.Height,
		ValueHash: v.ValueHash[:],
		Timestamp: v.Timestamp,
		Signature: v.Signature.ToProto(),
	}
}

func VoteFromProto(pb *pbtypes.Vote) *Vote {
	if pb == nil {
		return nil
	}
	hash := make([]byte, len(pb.ValueHash))
	copy(hash[:], pb.ValueHash)
	return &Vote{
		VoteType:  pb.VoteType,
		Height:    pb.Height,
		ValueHash: hash,
		Timestamp: pb.Timestamp,
		Signature: bls12.SignatureFromProto(pb.Signature),
	}
}
