package bls12

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	private, err := GeneratePrivateKey()
	assert.Nil(t, err)

	public := private.Public()

	t.Log("private key bytes:", private.ToBytes())
	t.Log("private key bytes:", private.ToBytes())
	t.Log("private key bytes:", private.ToBytes())
	t.Log("public key bytes:", public.ToBytes())
	t.Log("public key bytes:", public.ToBytes())
	t.Log("public key bytes:", public.ToBytes())
}
