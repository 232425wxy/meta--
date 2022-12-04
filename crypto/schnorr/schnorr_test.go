package schnorr

import (
	"fmt"
	"math/big"
	"testing"
)

func TestNiYuan(t *testing.T) {
	s := new(big.Int).SetInt64(7)
	n := new(big.Int).SetInt64(11)
	x := new(big.Int).SetInt64(0)
	y := new(big.Int).SetInt64(0)
	d := exgcd(s, n, x, y)
	m := mod_reverse(s, n)
	fmt.Println(d.String())
	fmt.Println(m.String())
}

func TestSub(t *testing.T) {
	a := new(big.Int).SetInt64(0)
	b := new(big.Int).SetInt64(1)
	res := sub(a, b)
	n := new(big.Int)
	n.Set(res)
	t.Log(n.String())
}
