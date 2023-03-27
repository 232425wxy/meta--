package bls12

import (
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	private, err := GeneratePrivateKey()
	assert.Nil(t, err)

	public := private.PublicKey()

	t.Log("private Key bytes:", private.ToBytes())
	t.Log("private Key bytes length:", len(private.ToBytes()))
	t.Log("public Key bytes:", public.ToBytes())
	t.Log("public Key bytes length:", len(public.ToBytes()))

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

	sig, err := bc.Sign(h[:])
	assert.Nil(t, err)
	t.Log("signature:", sig.ToBytes())
	t.Log("signer:", sig.Signer())
}

func TestSignAndVerify(t *testing.T) {
	private, err := GeneratePrivateKey()
	assert.Nil(t, err)
	public := private.PublicKey()

	msg := []byte("welcome to china!")
	h := sha256.Sum(msg)

	sig, err := private.Sign(h[:])
	assert.Nil(t, err)

	b := public.Verify(sig, h[:])
	assert.True(t, b)
}

func TestThreshold(t *testing.T) {
	private1, err := GeneratePrivateKey()
	public1 := private1.PublicKey()
	assert.Nil(t, err)

	private2, err := GeneratePrivateKey()
	public2 := private2.PublicKey()
	assert.Nil(t, err)

	private3, err := GeneratePrivateKey()
	public3 := private3.PublicKey()
	assert.Nil(t, err)

	private4, err := GeneratePrivateKey()
	public4 := private4.PublicKey()
	assert.Nil(t, err)

	err = AddBLSPublicKey(public1.ToBytes())
	assert.Nil(t, err)
	err = AddBLSPublicKey(public2.ToBytes())
	assert.Nil(t, err)
	err = AddBLSPublicKey(public3.ToBytes())
	assert.Nil(t, err)
	err = AddBLSPublicKey(public4.ToBytes())
	assert.Nil(t, err)

	msg := []byte("Let's test threshold signature!")
	h := sha256.Sum(msg)

	msg2 := []byte("Let's test threshold signature.")
	h2 := sha256.Sum(msg2)
	_ = h2

	sig1, err := private1.Sign(h[:])
	assert.Nil(t, err)

	sig2, err := private2.Sign(h[:])
	assert.Nil(t, err)

	sig3, err := private3.Sign(h[:])
	assert.Nil(t, err)

	sig4, err := private4.Sign(h[:])
	assert.Nil(t, err)

	sigs := []*Signature{sig1, sig2, sig3, sig4}

	cb := NewCryptoBLS12()
	thresholdSig, err := cb.CreateThresholdSignature(sigs)
	assert.Nil(t, err)

	assert.True(t, cb.VerifyThresholdSignature(thresholdSig, h[:]))
}

func TestAnyThreshold(t *testing.T) {
	private1, err := GeneratePrivateKey()
	public1 := private1.PublicKey()
	assert.Nil(t, err)

	private2, err := GeneratePrivateKey()
	public2 := private2.PublicKey()
	assert.Nil(t, err)

	private3, err := GeneratePrivateKey()
	public3 := private3.PublicKey()
	assert.Nil(t, err)

	private4, err := GeneratePrivateKey()
	public4 := private4.PublicKey()
	assert.Nil(t, err)

	private5, err := GeneratePrivateKey()
	public5 := private5.PublicKey()
	assert.Nil(t, err)

	err = AddBLSPublicKey(public1.ToBytes())
	assert.Nil(t, err)
	err = AddBLSPublicKey(public2.ToBytes())
	assert.Nil(t, err)
	err = AddBLSPublicKey(public3.ToBytes())
	assert.Nil(t, err)
	err = AddBLSPublicKey(public4.ToBytes())
	assert.Nil(t, err)
	err = AddBLSPublicKey(public5.ToBytes())
	assert.Nil(t, err)

	msg := []byte("blockchain")
	h := sha256.Sum(msg)

	sig1, _ := private1.Sign(h[:])
	sig2, _ := private2.Sign(h[:])
	sig3, _ := private3.Sign(h[:])
	sig4, _ := private4.Sign(h[:])
	sig5, _ := private5.Sign(h[:])
	_ = sig5

	sig1234 := []*Signature{sig1, sig2, sig3, sig4}

	cb := NewCryptoBLS12()
	thresholdSig1234, err := cb.CreateThresholdSignature(sig1234)

	assert.True(t, cb.VerifyThresholdSignature(thresholdSig1234, h[:]))
}
