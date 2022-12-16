package database

import "bytes"

type item struct {
	key   []byte
	value []byte
}

func (i *item) Less(other *item) bool {
	return bytes.Compare(i.key, other.key) == -1
}
