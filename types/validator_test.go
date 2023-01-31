package types

import (
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/proto/pbabci"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestValidatorSet_Update(t *testing.T) {
	validators := make([]*Validator, 4)
	for i := 0; i < 4; i++ {
		privateKey, _ := bls12.GeneratePrivateKey()
		validators[i] = NewValidator(privateKey.PublicKey(), 10)
	}
	validator1 := validators[1]
	updates := make([]*pbabci.ValidatorUpdate, 2)
	updates[0] = &pbabci.ValidatorUpdate{
		BLS12PublicKey: validator1.PublicKey.ToProto(),
		Power:          3,
	}
	privateKey, _ := bls12.GeneratePrivateKey()
	updates[1] = &pbabci.ValidatorUpdate{
		BLS12PublicKey: privateKey.PublicKey().ToProto(),
		Power:          20,
	}
	set := NewValidatorSet(validators)

	assert.Equal(t, 4, len(set.Validators))

	set.Update(updates)
	assert.Equal(t, 5, len(set.Validators))

	for _, validator := range set.Validators {
		if validator.ID == validator1.ID {
			assert.Equal(t, int64(3), validator.VotingPower)
		}
	}
}

func del(arr *[]int, size int) {
	*arr = append((*arr)[:size], (*arr)[size+1:]...)
}

func TestDel(t *testing.T) {
	arr := make([]int, 5)
	for i := 0; i < 5; i++ {
		arr[i] = i * 8
	}
	del(&arr, 4)
	for i := 0; i < len(arr); i++ {
		t.Log(arr[i])
	}
}

func TestSortVals(t *testing.T) {
	vals := make([]*Validator, 4)
	for i := 0; i < len(vals); i++ {
		vals[i] = &Validator{}
	}
	vals[0].ID = crypto.ID("c")
	vals[1].ID = crypto.ID("d")
	vals[2].ID = crypto.ID("a")
	vals[3].ID = crypto.ID("b")
	sort.Sort(Validators(vals))
	for i := 0; i < len(vals); i++ {
		t.Log(i, ":", vals[i].ID)
	}
}
