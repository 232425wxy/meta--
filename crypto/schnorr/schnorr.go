package schnorr

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/232425wxy/meta--/crypto/hash/sha256"
	"math/big"
)

var (
	p *big.Int
	q *big.Int
	g *big.Int

	sk *big.Int
	pk *big.Int
	k  *big.Int
)

func setup() {
	var err error
	p, _ = rand.Prime(rand.Reader, 512)
	one := new(big.Int).SetInt64(1)
	two := new(big.Int).SetInt64(2)
	q = new(big.Int)
	q.Sub(p, one)
	q.Div(q, two) // qq = (pp-1)/2
	g, err = rand.Int(rand.Reader, p)
	if err != nil {
		fmt.Printf("Generation of random g in bounds [0...%v] failed.", p)
	}
	g.Exp(g, two, p) // g = g**2 mod pp
}

func keyGen() {
	var err error
	sk, err = rand.Int(rand.Reader, q)
	if err != nil {
		fmt.Printf("Generation of random sk in bounds [0...%v] failed.", q)
	}
	pk = new(big.Int)
	pk.Exp(g, sk, p) // pk = g**sk mod pp
	k, err = rand.Int(rand.Reader, q)
	if err != nil {
		fmt.Printf("Generation of random k in bounds [0...%v] failed.", q)
	}
}

func sigGen(msg []byte) ([]byte, []byte) {
	K := new(big.Int).Exp(g, k, p) // K = g**k mod pp
	hash := sha256.New()
	KBytes := K.Bytes()
	hash.Write(KBytes)
	hash.Write(msg)
	h := hash.Sum(nil)
	c := new(big.Int).SetBytes(h)
	sig := add(k, mul(sk, c, p), p)
	fmt.Printf("签名：(%v, %v)\n", c.String(), sig.String())
	return c.Bytes(), sig.Bytes()
}

func verify(c, sig, msg []byte) bool {
	cBig := new(big.Int).SetBytes(c)
	sigBig := new(big.Int).SetBytes(sig)
	fmt.Printf("签名：(%v, %v)\n", cBig.String(), sigBig.String())
	inverseC := sub(new(big.Int).SetInt64(0), cBig)
	res1 := new(big.Int).Exp(pk, inverseC, p)
	res2 := new(big.Int).Exp(g, sigBig, p)
	m := mul(res1, res2, p)
	hash := sha256.New()
	hash.Write(m.Bytes())
	hash.Write(msg)
	return bytes.Equal(c, hash.Sum(nil))
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 利用扩展欧几里得算法求逆元

// ax + by = 1，求a mod b 的逆元
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

func mul(a, b, m *big.Int) *big.Int {
	res := new(big.Int).Mul(a, b)
	res = mod(res, m)
	return res
}

func add(a, b, m *big.Int) *big.Int {
	res := new(big.Int).Add(a, b)
	res = mod(res, m)
	return res
}

func exp(a, b, m *big.Int) *big.Int {
	//ex := mod(b, m)
	res := new(big.Int).Exp(a, b, m)
	return res
}

func sub(a, b *big.Int) *big.Int {
	return new(big.Int).Sub(a, b)
}
