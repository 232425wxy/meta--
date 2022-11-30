package bls12

import (
	"github.com/232425wxy/meta--/crypto/hash/sha256"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	private, err := GeneratePrivateKey()
	assert.Nil(t, err)

	public := private.Public()

	t.Log("private key bytes:", private.ToBytes())
	t.Log("private key bytes:", len(private.ToBytes()))
	t.Log("public key bytes:", public.ToBytes())
	t.Log("public key bytes:", len(public.ToBytes()))

	id := public.ToID()
	t.Log(id, len(id))
	t.Log(id.ToBytes(), len(id.ToBytes()))
}

func TestBls12Crypto_Sign(t *testing.T) {
	private, err := GeneratePrivateKey()
	assert.Nil(t, err)
	bc := NewCryptoBLS12()
	bc.Init(private)

	msg := []byte("Welcome to China!")
	h := sha256.Sum(msg)

	sig, err := bc.Sign(h)
	assert.Nil(t, err)
	t.Log("signature:", sig.ToBytes())
	t.Log("signer:", sig.Signer())
}

func TestSignAndVerify(t *testing.T) {
	private, err := GeneratePrivateKey()
	assert.Nil(t, err)
	public := private.Public()

	msg := []byte("welcome to china!")
	h := sha256.Sum(msg)

	sig, err := private.Sign(h)
	assert.Nil(t, err)

	b := public.Verify(sig, h)
	assert.True(t, b)
}

func TestThreshold(t *testing.T) {
	private1, err := GeneratePrivateKey()
	assert.Nil(t, err)

	private2, err := GeneratePrivateKey()
	assert.Nil(t, err)

	private3, err := GeneratePrivateKey()
	assert.Nil(t, err)

	private4, err := GeneratePrivateKey()
	assert.Nil(t, err)

	msg := []byte("Let's test threshold signature!")
	h := sha256.Sum(msg)

	sig1 := private1.Sign()
}