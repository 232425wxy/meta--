package stch

import (
	"crypto/rand"
	"math/big"
)

// GenerateKAndX 生成节点的秘密值k和身份标识x，k是小于q的一个随机值，x = g^k mod q
func GenerateKAndX() (*big.Int, *big.Int) {
	k, err := rand.Int(rand.Reader, q)
	if err != nil {
		panic(err)
	}
	x := new(big.Int).Exp(g, k, q)
	return k, x
}

// GeneratePolynomialItem 随机生成多项式项的系数。
func GeneratePolynomialItem() *big.Int {
	item, err := rand.Int(rand.Reader, q)
	if err != nil {
		panic(err)
	}
	return item
}
