package types

import (
	"errors"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"time"
)

type Proposal struct {
	Height      int64            `json:"height"`
	SimpleBlock SimpleBlock      `json:"simple_block"`
	Timestamp   time.Time        `json:"timestamp"`
	Signature   *bls12.Signature `json:"signature"`
}

func NewProposal(height int64, simple SimpleBlock) *Proposal {
	return &Proposal{
		Height:      height,
		SimpleBlock: simple,
		Timestamp:   time.Now(),
	}
}

func (p *Proposal) ValidateBasic() error {
	if p.Height < 0 {
		return errors.New("negative height")
	}
	return nil
}

func (p *Proposal) ToProto() *pbtypes.Proposal {
	return &pbtypes.Proposal{
		Height:      p.Height,
		SimpleBlock: p.SimpleBlock.ToProto(),
		Timestamp:   p.Timestamp,
		Signature:   p.Signature.ToProto(),
	}
}
