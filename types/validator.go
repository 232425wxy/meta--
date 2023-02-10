package types

import (
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/proto/pbabci"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"sort"
)

type Validator struct {
	ID             crypto.ID        `json:"ID"`
	PublicKey      *bls12.PublicKey `json:"public_key"`
	VotingPower    int64            `json:"voting_power"`
	LeaderPriority int64            `json:"leader_priority"`
}

func NewValidator(publicKey *bls12.PublicKey, votingPower int64) *Validator {
	return &Validator{
		ID:             publicKey.ToID(),
		PublicKey:      publicKey,
		VotingPower:    votingPower,
		LeaderPriority: 0,
	}
}

func (v *Validator) ToProto() *pbtypes.Validator {
	if v == nil {
		return nil
	}
	return &pbtypes.Validator{
		ID:             string(v.ID),
		PublicKey:      v.PublicKey.ToProto(),
		VotingPower:    v.VotingPower,
		LeaderPriority: v.LeaderPriority,
	}
}

func ValidatorFromProto(pb *pbtypes.Validator) *Validator {
	if pb == nil {
		return nil
	}
	return &Validator{
		ID:             crypto.ID(pb.ID),
		PublicKey:      bls12.PublicKeyFromProto(pb.PublicKey),
		VotingPower:    pb.VotingPower,
		LeaderPriority: pb.LeaderPriority,
	}
}

type Validators []*Validator

func (vals Validators) Len() int {
	return len(vals)
}

func (vals Validators) Less(i, j int) bool {
	return vals[i].ID < vals[j].ID
}

func (vals Validators) Swap(i, j int) {
	vals[i], vals[j] = vals[j], vals[i]
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// validator.proto 集合

type ValidatorSet struct {
	Validators       []*Validator `json:"validators"`
	Leader           *Validator   `json:"leader"`
	TotalVotingPower int64        `json:"total_voting_power"`
}

func (set *ValidatorSet) GetLeader(round int16) *Validator {
	index := int(round) % len(set.Validators)
	set.Leader = set.Validators[index]
	return set.Leader
}

func (set *ValidatorSet) Copy() *ValidatorSet {
	cpy := &ValidatorSet{
		Validators:       make([]*Validator, len(set.Validators)),
		Leader:           set.Leader,
		TotalVotingPower: set.TotalVotingPower,
	}
	copy(cpy.Validators, set.Validators)
	return cpy
}

func NewValidatorSet(validators []*Validator) *ValidatorSet {
	sort.Sort(Validators(validators))
	set := &ValidatorSet{Validators: validators}
	for _, validator := range validators {
		set.TotalVotingPower += validator.VotingPower
	}
	return set
}

func (set *ValidatorSet) GetValidatorByID(id crypto.ID) *Validator {
	for _, val := range set.Validators {
		if val.ID == id {
			return val
		}
	}
	return nil
}

func (set *ValidatorSet) PowerMajor23() int64 {
	set.TotalVotingPower = 0
	for _, val := range set.Validators {
		set.TotalVotingPower += val.VotingPower
	}
	major23 := set.TotalVotingPower * 2
	major23 /= 3
	return major23
}

func (set *ValidatorSet) PowerMajorFull() int64 {
	set.TotalVotingPower = 0
	for _, val := range set.Validators {
		set.TotalVotingPower += val.VotingPower
	}
	return set.TotalVotingPower
}

func (set *ValidatorSet) Update(validatorUpdates []*pbabci.ValidatorUpdate) {
	for _, update := range validatorUpdates {
		var exists bool = false
		publicKey := bls12.PublicKeyFromProto(update.BLS12PublicKey)
		for i, validator := range set.Validators {
			if publicKey.ToID() == validator.ID {
				exists = true
				validator.VotingPower = update.Power
				if update.Power <= 0 {
					set.Validators = append(set.Validators[:i], set.Validators[i+1:]...)
					if publicKey.ToID() == set.Leader.ID {
						set.Leader = set.Validators[0]
					}
				}
			}
		}
		if !exists {
			set.Validators = append(set.Validators, &Validator{
				ID:             publicKey.ToID(),
				PublicKey:      publicKey,
				VotingPower:    update.Power,
				LeaderPriority: 10,
			})
		}
	}
}

func (set *ValidatorSet) ToProto() *pbtypes.ValidatorSet {
	if set == nil {
		return nil
	}
	validators := make([]*pbtypes.Validator, 0)
	for _, validator := range set.Validators {
		validators = append(validators, validator.ToProto())
	}
	return &pbtypes.ValidatorSet{
		Validators:       validators,
		Leader:           set.Leader.ToProto(),
		TotalVotingPower: set.TotalVotingPower,
	}
}

func ValidatorSetFromProto(pb *pbtypes.ValidatorSet) *ValidatorSet {
	if pb == nil {
		return nil
	}
	validators := make([]*Validator, 0)
	for _, validator := range pb.Validators {
		validators = append(validators, ValidatorFromProto(validator))
	}
	return &ValidatorSet{
		Validators:       validators,
		Leader:           ValidatorFromProto(pb.Leader),
		TotalVotingPower: pb.TotalVotingPower,
	}
}
