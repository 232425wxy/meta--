package schnorr

import (
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
	p, _ = rand.Prime(rand.Reader, 128)
	one := new(big.Int).SetInt64(1)
	two := new(big.Int).SetInt64(2)
	q.Sub(p, one)
	q.Div(q, two)
	g, err = rand.Int(rand.Reader, p)
	if err != nil {
		fmt.Printf("Generation of random g in bounds [0...%v] failed.", p)
	}
	g.Exp(g, two, p)
}

func keyGen() {
	var err error
	sk, err = rand.Int(rand.Reader, q)
	if err != nil {
		fmt.Printf("Generation of random sk in bounds [0...%v] failed.", q)
	}
	pk.Exp(g, sk, p)
	k, err = rand.Int(rand.Reader, q)
	if err != nil {
		fmt.Printf("Generation of random k in bounds [0...%v] failed.", q)
	}
}

func sigGen(msg []byte) ([]byte, []byte) {
	K := new(big.Int).Exp(g, k, p)
	hash := sha256.New()
	KBytes := K.Bytes()
	hash.Write(KBytes)
	hash.Write(msg)
	h := hash.Sum(nil)
	c := new(big.Int).SetBytes(h)
	sig := k.Add(k, sk.Mul(sk, c))
	return c.Bytes(), sig.Bytes()
}

//func verify(c, sig []byte) {
//	cBig := new(big.Int).SetBytes(c)
//	cBig.Cmp()
//}

func exgcd(a, b, x, y *big.Int) *big.Int {
	var d *big.Int
	if b.Cmp(new(big.Int).SetInt64(0)) == 0 {
		x.SetInt64(1)
		y.SetInt64(0)
		return a
	}
	m := mod(a, b)        // 对的
	d = exgcd(b, m, y, x) // 对的
	di := div(a, b)       // 对的
	mu := mul(di, x)
	fmt.Printf("div: %v, x: %v, y: %v, mu: %v\n", di.String(), x.String(), y.String(), mu.String())
	y = sub(y, mu)
	fmt.Printf("y: %v\n", y.String())
	return d
	// div: 3, x: 0, y: 1, mu: 0
	//y: 1
	//div: 1, x: 1, y: 0, mu: 1
	//y: -1
	//div: 1, x: -1, y: 1, mu: -1
	//y: 2
	//div: 1, x: 2, y: -1, mu: 2
	//y: -3
	//div: 0, x: -3, y: 2, mu: 0
	//y: 2
	//div: 3, x: 0, y: 1, mu: 0
	//y: 1
	//div: 1, x: 1, y: 0, mu: 1
	//y: -1
	//div: 1, x: -1, y: 1, mu: -1
	//y: 2
	//div: 1, x: 2, y: -1, mu: 2
	//y: -3
	//div: 0, x: -3, y: 2, mu: 0
	//y: 2
}

func mod_reverse(a, m *big.Int) *big.Int {
	var d, x, y *big.Int
	x = new(big.Int)
	y = new(big.Int)
	d = exgcd(a, m, x, y)
	if d.Cmp(new(big.Int).SetInt64(1)) == 0 {
		xmod := mod(x, m)
		if xmod.Cmp(new(big.Int).SetInt64(0)) == -1 || xmod.Cmp(new(big.Int).SetInt64(0)) == 0 {
			return xmod.Add(xmod, m)
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

func mul(a, b *big.Int) *big.Int {
	res := new(big.Int)
	return res.Mul(a, b)
}

func div(a, b *big.Int) *big.Int {
	res := new(big.Int)
	return res.Div(a, b)
}

func add(a, b *big.Int) *big.Int {
	res := new(big.Int)
	return res.Add(a, b)
}

func sub(a, b *big.Int) *big.Int {
	res := new(big.Int)
	return res.Sub(a, b)
}
