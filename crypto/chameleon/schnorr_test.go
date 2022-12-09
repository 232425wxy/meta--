package chameleon

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchnorr(t *testing.T) {
	public, private := GenerateKey()

	signature, err := private.Sign([]byte("hi"))
	assert.Nil(t, err)

	ret := public.Verify(signature)
	assert.True(t, ret)
}
