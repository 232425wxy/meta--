package sha256

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSum(t *testing.T) {
	text := []byte("hello, world")
	sum32 := Sum(text)
	sum20 := Sum20(text)
	assert.Equal(t, sum32[:20], sum20[:])
}
