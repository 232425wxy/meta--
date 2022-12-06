package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadAndSave(t *testing.T) {
	path := "nodeKey.json"
	key, err := LoadOrGenNodeKey(path)
	assert.Nil(t, err)
	assert.Nil(t, key.SaveAs(path))
}

func TestLoad(t *testing.T) {
	path := "nodeKey.json"
	key, err := LoadNodeKey(path)
	assert.Nil(t, err)
	t.Log(key)
}
