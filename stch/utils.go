package stch

import (
	"crypto/rand"
	"fmt"
	"github.com/232425wxy/meta--/crypto/sha256"
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

func exgcd(a, b, x, y *big.Int) *big.Int {
	var d *big.Int
	if b.Cmp(new(big.Int).SetInt64(0)) == 0 {
		x.SetInt64(1)
		y.SetInt64(0)
		return new(big.Int).Set(a)
	}
	m := mod(a, b)
	d = exgcd(b, m, y, x)
	di := div(a, b)
	di.Mul(di, x)
	y.Sub(y, di)
	return new(big.Int).Set(d)
}

// ax + by = 1，求a mod b 的逆元
func calcInverseElem(a, b *big.Int) *big.Int {
	var d, x, y *big.Int
	x = new(big.Int)
	y = new(big.Int)
	d = exgcd(a, b, x, y)
	if d.Cmp(new(big.Int).SetInt64(1)) == 0 {
		xmod := mod(x, b)
		if xmod.Cmp(new(big.Int).SetInt64(0)) == -1 || xmod.Cmp(new(big.Int).SetInt64(0)) == 0 {
			return xmod.Add(xmod, b)
		} else {
			return xmod
		}
	} else {
		return new(big.Int).SetInt64(-1)
	}
}

func mod(a, b *big.Int) *big.Int {
	res := new(big.Int)
	return res.Mod(a, b)
}

func div(a, b *big.Int) *big.Int {
	res := new(big.Int)
	return res.Div(a, b)
}

func redactHash(blockHeight int64, txIndex int, newTx []byte) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d:%d:%x", blockHeight, txIndex, newTx)))
	val := h.Sum(nil)
	return fmt.Sprintf("%x", val)
}
