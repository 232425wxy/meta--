package cmap

import (
	"fmt"
	"github.com/232425wxy/meta--/common/rand"
	"testing"
)

func BenchmarkCMap(b *testing.B) {
	cm := NewCap()
	go func() {
		for i := 0; i < b.N*2; i++ {
			r := rand.Intn(i + 100)
			if cm.Has(fmt.Sprintf("key-%d", r)) {
				b.Log(fmt.Sprintf("cmap has %v:%v", fmt.Sprintf("key-%d", r), cm.Get(fmt.Sprintf("key-%d", r))))
			}
		}
	}()
	for i := 0; i < b.N; i++ {
		cm.Set(fmt.Sprintf("key-%d", i), i)
	}
}
