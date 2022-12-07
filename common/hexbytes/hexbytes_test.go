package hexbytes

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHexBytes_MarshalJSON(t *testing.T) {
	var bz HexBytes = []byte("hello, world")
	marshal, err := bz.MarshalJSON()
	assert.Nil(t, err)
	t.Log(string(marshal))
}

func TestHexBytes_UnmarshalJSON(t *testing.T) {
	var bz HexBytes = []byte("hello, world")
	marshal, err := bz.MarshalJSON()
	assert.Nil(t, err)
	dst := &HexBytes{}
	err = dst.UnmarshalJSON(marshal)
	assert.Nil(t, err)
	t.Log(string(*dst))
}

func TestHexBytes_CompatibleWith(t *testing.T) {
	var h1 HexBytes = []byte{1, 2, 3, 4, 5}
	var h2 HexBytes = []byte{9, 8, 7}
	var h3 HexBytes = []byte{9, 8, 7, 6, 5}

	assert.False(t, h1.CompatibleWith(h2))
	assert.True(t, h3.CompatibleWith(h1))
}
