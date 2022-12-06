package os

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileExists(t *testing.T) {
	notExists := "../go.mod"
	exists := "../../go.mod"

	assert.False(t, FileExists(notExists))
	assert.True(t, FileExists(exists))
}
