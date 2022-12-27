package database

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func key(i int) []byte {
	res := make([]byte, 0)
	for i >= 0 {
		if i >= 26 {
			res = append(res, 'z')
			i -= 26
		} else {
			res = append(res, 'a'+byte(i))
			i = i - i - 1
		}
	}
	return res
}

func TestMemDBIterator(t *testing.T) {
	mem := NewMemDB()

	for i := 0; i < 100; i++ {
		k := key(i)
		v := []byte(fmt.Sprintf("block:%d", i))
		err := mem.Set(k, v)
		assert.Nil(t, err)
	}
	t.Log(mem.Stats())

	start := []byte{'d'}
	end := []byte("zzd")
	iter := newMemDBIterator(mem, start, end, true)
	for iter.Valid() {
		t.Log(string(iter.Key()), "->", string(iter.Value()))
		iter.Next()
	}
}

func test(ch chan int) {
	go func() {
		for i := 0; i < 100; i++ {
			select {
			case ch <- i:

			}
		}
		close(ch)
	}()
}

func TestName(t *testing.T) {
	ch := make(chan int, 4)
	test(ch)
	//close(ch)
	for i := range ch {
		t.Log(i)
	}
}
