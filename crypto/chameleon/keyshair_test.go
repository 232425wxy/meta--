package chameleon

import (
	"fmt"
	"github.com/232425wxy/meta--/crypto/sha256"
	"math/big"
	"testing"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

type Polynomial struct {
	items map[int]*big.Int
}

func (poly *Polynomial) calc(x *big.Int) *big.Int {
	res := new(big.Int).SetInt64(0)
	for order, item := range poly.items {
		exp := new(big.Int).Exp(x, new(big.Int).SetInt64(int64(order)), Q)
		exp.Mul(exp, item)
		res.Add(res, exp)
	}
	return res.Mod(res, Q)
	//return res
}

type Participant2 struct {
	poly      *Polynomial
	k         *big.Int
	x         *big.Int
	ski       *big.Int
	pki       *big.Int
	ai        *big.Int
	f         *big.Int
	auxiliary *big.Int
}

func CreateParticipants(n, t int) []*Participant2 {
	participants := make([]*Participant2, n)
	for i := 0; i < n; i++ {
		participant := &Participant2{
			poly:      &Polynomial{items: make(map[int]*big.Int)},
			k:         randGen(Q),
			f:         new(big.Int).SetInt64(0),
			auxiliary: new(big.Int).SetInt64(1),
		}
		participant.x = new(big.Int).Exp(G, participant.k, Q)
		for j := 0; j < t; j++ {
			participant.poly.items[j] = randGen(Q)
		}
		participants[i] = participant
	}

	secretKey := new(big.Int).SetInt64(0)
	for _, participant := range participants {
		secretKey.Add(secretKey, participant.poly.items[0])
	}
	secretKey.Mod(secretKey, Q)
	fmt.Printf("完整私钥：%s\n", secretKey)

	for _, participantx := range participants {
		for _, participanty := range participants {
			participantx.f.Add(participantx.f, participanty.poly.calc(participantx.x))
			participantx.f.Mod(participantx.f, Q)
		}
	}

	authorized := []*Participant2{participants[0], participants[1], participants[2], participants[3]}
	for _, participantx := range authorized {
		for _, participanty := range authorized {
			if participantx.x.Cmp(participanty.x) == 0 {
				continue
			}
			neg := new(big.Int).Neg(participanty.x)
			diff := new(big.Int).Sub(participantx.x, participanty.x)
			inverse := calcInverseElem(diff, Q)
			d := new(big.Int).Mul(neg, inverse)
			participantx.auxiliary.Mul(participantx.auxiliary, d)
		}
	}

	CID := new(big.Int).SetInt64(0)
	_secretKey := new(big.Int).SetInt64(0)
	for _, participant := range authorized {
		m := new(big.Int).Mul(participant.f, participant.auxiliary)
		m = m.Mod(m, Q)
		participant.ski = m
		participant.pki = new(big.Int).Exp(G, m, Q)
		_secretKey.Add(_secretKey, m)
		hk.Mul(hk, participant.pki)
		hk.Mod(hk, Q)
		CID.Add(CID, participant.x)
		CID.Mod(CID, Q)
	}
	_secretKey.Mod(_secretKey, Q)

	fmt.Printf("计算私钥：%s\n", _secretKey)

	fmt.Println("变色龙哈希函数的公钥：", hk.String())

	// 计算变色龙哈希值

	msg := []byte("name=wxy")

	sigma := randGen(Q)

	type R struct {
		val1 *big.Int
		val2 *big.Int
	}

	r := &R{
		val1: new(big.Int).Exp(G, sigma, Q),
		val2: new(big.Int).Exp(hk, sigma, Q),
	}

	_ = r

	alpha := HashBigInt(CID, hk)

	for _, participant := range authorized {
		participant.ai = new(big.Int).Exp(alpha, participant.k, Q)
	}

	h := new(big.Int).Exp(G, sigma, Q)
	e := new(big.Int).Exp(alpha, HashBytes(msg), Q)
	h.Mul(h, e)
	h.Mod(h, Q)

	fmt.Printf("哈希值：%s\n", h.String())

	// 计算哈希碰撞

	_msg := []byte("name=fsj")

	ee := new(big.Int).Sub(HashBytes(msg), HashBytes(_msg))
	fmt.Println("msg-_msg", ee.String())

	s1 := new(big.Int).Mul(authorized[0].ski, ee)
	s1.Add(s1, authorized[0].k)
	d1 := new(big.Int).Exp(alpha, s1, Q)
	_ = d1

	s2 := new(big.Int).Mul(authorized[1].ski, ee)
	s2.Add(s2, authorized[1].k)
	d2 := new(big.Int).Exp(alpha, s2, Q)
	_ = d2

	s3 := new(big.Int).Mul(authorized[2].ski, ee)
	s3.Add(s3, authorized[2].k)
	d3 := new(big.Int).Exp(alpha, s3, Q)
	_ = d3

	s4 := new(big.Int).Mul(authorized[3].ski, ee)
	s4.Add(s4, authorized[3].k)
	d4 := new(big.Int).Exp(alpha, s4, Q)
	_ = d4

	one1 := new(big.Int).Exp(G, s1, Q)
	_pk1 := calcInverseElem(authorized[0].pki, Q)
	three1 := new(big.Int).Exp(_pk1, ee, Q)
	four1 := new(big.Int).Mul(one1, three1)
	four1.Mod(four1, Q)
	fmt.Println("差值：", authorized[0].x.Sub(authorized[0].x, four1))

	one2 := new(big.Int).Exp(G, s2, Q)
	_pk2 := calcInverseElem(authorized[1].pki, Q)
	three2 := new(big.Int).Exp(_pk2, ee, Q)
	four2 := new(big.Int).Mul(one2, three2)
	four2.Mod(four2, Q)
	fmt.Println("差值：", authorized[1].x.Sub(authorized[1].x, four2))

	one3 := new(big.Int).Exp(G, s3, Q)
	_pk3 := calcInverseElem(authorized[2].pki, Q)
	three3 := new(big.Int).Exp(_pk3, ee, Q)
	four3 := new(big.Int).Mul(one3, three3)
	four3.Mod(four3, Q)
	fmt.Println("差值：", authorized[2].x.Sub(authorized[2].x, four3))

	one4 := new(big.Int).Exp(G, s4, Q)
	_pk4 := calcInverseElem(authorized[3].pki, Q)
	three4 := new(big.Int).Exp(_pk4, ee, Q)
	four4 := new(big.Int).Mul(one4, three4)
	four4.Mod(four4, Q)
	fmt.Println("差值：", authorized[3].x.Sub(authorized[3].x, four4))

	A := new(big.Int).SetInt64(1)
	for _, participant := range authorized {
		A.Mul(A, participant.ai)
		A.Mod(A, Q)
	}
	_A := calcInverseElem(A, Q)

	c := new(big.Int).Mul(d1, d2)
	c.Mul(c, d3)
	c.Mul(c, d4)
	c.Mul(c, _A)

	_r := &R{}
	tmp := new(big.Int).Exp(alpha, ee, Q)
	_r.val1 = new(big.Int).Mul(r.val1, tmp)
	_r.val1.Mod(_r.val1, Q)

	_r.val2 = new(big.Int).Mul(r.val2, c)
	_r.val2.Mod(_r.val2, Q)

	one := new(big.Int).Exp(alpha, HashBytes(_msg), Q)
	res := new(big.Int).Mul(_r.val1, one)
	res.Mod(res, Q)

	fmt.Println("碰撞哈希值：", res.String())

	return participants
}

func TestExample(t *testing.T) {
	CreateParticipants(7, 4)

	fmt.Println("---------------------------------")

	seven := new(big.Int).SetInt64(9)
	two := new(big.Int).SetInt64(5)
	_two := calcInverseElem(two, seven)
	//fmt.Println(_two)
	//_ex := new(big.Int).SetInt64(-3)
	ex := new(big.Int).SetInt64(6)

	res1 := new(big.Int).Exp(_two, ex, seven)
	res2 := new(big.Int).Exp(two, ex, seven)
	res3 := new(big.Int).Mul(res1, res2)
	res3.Mod(res3, seven)
	fmt.Println(res3)
}

func HashBigInt(vals ...*big.Int) *big.Int {
	fn := sha256.New()
	for _, val := range vals {
		fn.Write(val.Bytes())
	}
	h := fn.Sum(nil)
	res := new(big.Int).SetBytes(h)
	//res.Mod(res, Q)
	return res
}

func HashBytes(m []byte) *big.Int {
	fn := sha256.New()
	fn.Write(m)
	h := fn.Sum(nil)
	res := new(big.Int).SetBytes(h)
	return res
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

func TestName2(t *testing.T) {
	num := new(big.Int).SetInt64(-2)
	pp := new(big.Int).SetInt64(19)
	t.Log(calcInverseElem(num, pp))
	num2 := new(big.Int).Set(calcInverseElem(num, pp))
	res := new(big.Int).Mul(num2, num)
	t.Log(res.Mod(res, pp))
}
