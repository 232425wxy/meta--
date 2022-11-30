package crypto

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestID_Bytes(t *testing.T) {
	var id ID = "333"
	t.Log(id.ToBytes())
}

func TestIDSet(t *testing.T) {
	set := NewIDSet(0)

	set.AddID("1234567890")
	set.AddID("1234567890")
	assert.Equal(t, 1, set.Size())

	set.AddID("0987654321")
	set.AddID("0987654321")
	assert.Equal(t, 2, set.Size())

	set.RemoveID("0987654321")
	set.RemoveID("0987654321")
	assert.Equal(t, 1, set.Size())

	set.RemoveID("1234567890")
	set.RemoveID("1234567890")
	assert.Equal(t, 0, set.Size())
}

func TestFormatW(t *testing.T) {
	err := errors.New(fmt.Sprintf("ni ke zhen shuai ya! %s", "tom"))
	fmt.Println(fmt.Errorf("error: %q", err))
}
