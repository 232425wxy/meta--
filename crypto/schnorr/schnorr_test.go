package schnorr

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestSchnorr(t *testing.T) {
	public, private := GenerateKey()

	signature, err := private.Sign([]byte("hi"))
	assert.Nil(t, err)

	ret := public.Verify(signature)
	assert.True(t, ret)

	s := 8 << 3
	t.Log(s)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

type polynomial struct {
	items map[int]*big.Int
}

func (poly *polynomial) calc(x *big.Int) *big.Int {
	res := new(big.Int).SetInt64(0)
	for order, item := range poly.items {
		exp := new(big.Int).Exp(x, new(big.Int).SetInt64(int64(order)), q)
		exp.Mul(exp, item)
		res.Add(res, exp)
	}
	return res.Mod(res, q)
}

type Participant struct {
	poly      *polynomial
	k         *big.Int
	x         *big.Int
	ski       *big.Int
	pki       *big.Int
	f         *big.Int
	auxiliary *big.Int
}

func CreateParticipants(n, t int) []*Participant {
	participants := make([]*Participant, n)
	for i := 0; i < n; i++ {
		participant := &Participant{
			poly:      &polynomial{items: make(map[int]*big.Int)},
			k:         randGen(q),
			f:         new(big.Int).SetInt64(0),
			auxiliary: new(big.Int).SetInt64(1),
		}
		participant.x = new(big.Int).Exp(g, participant.k, q)
		for j := 0; j < t; j++ {
			participant.poly.items[j] = randGen(q)
		}
		participants[i] = participant
	}

	secretKey := new(big.Int).SetInt64(0)
	for _, participant := range participants {
		secretKey.Add(secretKey, participant.poly.items[0])
	}
	secretKey.Mod(secretKey, q)
	fmt.Printf("完整私钥：%s\n", secretKey)

	for _, participantx := range participants {
		for _, participanty := range participants {
			participantx.f.Add(participantx.f, participanty.poly.calc(participantx.x))
		}
		participantx.f.Mod(participantx.f, q)
	}

	F := func() *polynomial {
		poly := &polynomial{items: make(map[int]*big.Int)}
		for _, participant := range participants {
			for order, item := range participant.poly.items {
				if poly.items[order] == nil {
					poly.items[order] = new(big.Int).SetInt64(0)
				}
				poly.items[order].Add(poly.items[order], item)
			}
		}
		//for order, _ := range poly.items {
		//	poly.items[order].Mod(poly.items[order], q)
		//}
		return poly
	}

	for _, participant := range participants {
		f := F().calc(participant.x)
		if f.Cmp(participant.f) != 0 {
			panic("not equal")
		}
	}

	authorized := []*Participant{participants[0], participants[1], participants[2], participants[3]}
	for _, participantx := range authorized {
		for _, participanty := range authorized {
			if participantx.x.Cmp(participanty.x) == 0 {
				continue
			}
			neg := new(big.Int).Neg(participanty.x)
			diff := new(big.Int).Sub(participantx.x, participanty.x)
			inverse := calcInverseElem(diff, q)
			d := new(big.Int).Mul(neg, inverse)
			participantx.auxiliary.Mul(participantx.auxiliary, d)
		}
	}

	_secretKey := new(big.Int).SetInt64(0)
	for _, participant := range authorized {
		m := new(big.Int).Mul(participant.f, participant.auxiliary)
		_secretKey.Add(_secretKey, m)
	}
	_secretKey.Mod(_secretKey, q)

	fmt.Printf("计算私钥：%s\n", _secretKey)

	return participants
}

func TestExample(t *testing.T) {
	CreateParticipants(7, 4)
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
