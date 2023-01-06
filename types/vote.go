package types

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"github.com/cosmos/gogoproto/proto"
	"time"
)

type Vote struct {
	VoteType  pbtypes.VoteType
	Height    int64
	BlockHash []byte
	Timestamp time.Time
	Voter     crypto.ID
	Signature *bls12.Signature
}

func (v *Vote) ToSignBytes() []byte {
	pb := &pbtypes.Vote{
		VoteType:  v.VoteType,
		Height:    v.Height,
		BlockHash: v.BlockHash,
		Timestamp: v.Timestamp,
		Voter:     string(v.Voter),
	}
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return bz
}

func (v *Vote) Verify() error {
	publicKey := bls12.GetBLSPublicKeyFromLib(v.Voter)
	if publicKey == nil {
		return fmt.Errorf("cannot find the public key corresponding to %v", v.Voter)
	}
	hashBytes := sha256.Sum(v.ToSignBytes())
	ok := publicKey.Verify(v.Signature, hashBytes)
	if ok {
		return nil
	}
	return fmt.Errorf("invalid signature for %s", v.VoteType)
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
		BlockHash: v.BlockHash,
		Timestamp: v.Timestamp,
		Voter:     string(v.Voter),
		Signature: v.Signature.ToProto(),
	}
}

func VoteFromProto(pb *pbtypes.Vote) *Vote {
	return &Vote{
		VoteType:  pb.VoteType,
		Height:    pb.Height,
		BlockHash: pb.BlockHash,
		Timestamp: pb.Timestamp,
		Voter:     crypto.ID(pb.Voter),
		Signature: bls12.SignatureFromProto(pb.Signature),
	}
}
