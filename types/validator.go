package types

import (
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
)

type Validator struct {
	ID               crypto.ID        `json:"ID"`
	PublicKey        *bls12.PublicKey `json:"public_key"`
	VotingPower      int64            `json:"voting_power"`
	ProposerPriority int64            `json:"proposer_priority"`
}

func NewValidator(publicKey *bls12.PublicKey, votingPower int64) *Validator {
	return &Validator{
		ID:               publicKey.ToID(),
		PublicKey:        publicKey,
		VotingPower:      votingPower,
		ProposerPriority: 0,
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// validator.proto 集合

type ValidatorSet struct {
	Validators       []*Validator
	Proposer         *Validator
	TotalVotingPower int64
}

func (set *ValidatorSet) Copy() *ValidatorSet {
	cpy := &ValidatorSet{
		Validators:       make([]*Validator, len(set.Validators)),
		Proposer:         set.Proposer,
		TotalVotingPower: set.TotalVotingPower,
	}
	copy(cpy.Validators, set.Validators)
	return cpy
}

func NewValidatorSet(validators []*Validator) *ValidatorSet {
	set := &ValidatorSet{Validators: validators}
	for _, validator := range validators {
		set.TotalVotingPower += validator.VotingPower
	}
	return set
}
