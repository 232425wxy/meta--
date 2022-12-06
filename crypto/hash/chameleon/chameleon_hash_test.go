package chameleon

import (
	"math/big"
	"testing"
)

func TestRun(t *testing.T) {
	run()
}

func TestParams(t *testing.T) {
	for i := 0; i < 100; i++ {
		k := randGen(Q)
		t.Log("k:", k)
		x := new(big.Int).Exp(G, k, Q)
		t.Log("x:", x)
	}
}
