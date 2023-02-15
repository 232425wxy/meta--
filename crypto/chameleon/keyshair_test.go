package chameleon

import (
	"crypto/rand"
	"fmt"
	"github.com/232425wxy/meta--/crypto/sha256"
	"math/big"
	"testing"
)

var Q, _ = new(big.Int).SetString("7FFFFFFFFFFFFFFFD6FC2A2C515DA54D57EE2B10139E9E78EC5CE2C1E7169B4AD4F09B208A3219FDE649CEE7124D9F7CBE97F1B1B1863AEC7B40D901576230BD69EF8F6AEAFEB2B09219FA8FAF83376842B1B2AA9EF68D79DAAB89AF3FABE49ACC278638707345BBF15344ED79F7F4390EF8AC509B56F39A98566527A41D3CBD5E0558C159927DB0E88454A5D96471FDDCB56D5BB06BFA340EA7A151EF1CA6FA572B76F3B1B95D8C8583D3E4770536B84F017E70E6FBF176601A0266941A17B0C8B97F4E74C2C1FFC7278919777940C1E1FF1D8DA637D6B99DDAFE5E17611002E2C778C1BE8B41D96379A51360D977FD4435A11C30942E4BFFFFFFFFFFFFFFFF", 16)

var G, _ = new(big.Int).SetString("2", 10)

var (
	hk *big.Int = new(big.Int).SetInt64(1)
	tk *big.Int = new(big.Int).SetInt64(0)
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
	alphai    *big.Int
	fnx       *big.Int
	auxiliary *big.Int
}

type random struct {
	r1 *big.Int
	r2 *big.Int
}

func (r *random) String() string {
	return fmt.Sprintf("random{\n\tr1: %s\n\tr2: %s\n}", r.r1, r.r2)
}

type segSig struct {
	s *big.Int
	d *big.Int
}

func TestDistributedChameleonHash(t *testing.T) {
	n := 4
	ps := make([]*Participant2, n)
	for i := 0; i < n; i++ {
		p := &Participant2{
			poly: &Polynomial{items: make(map[int]*big.Int)},
			k:    randGen(Q),
		}
		p.x = new(big.Int).Exp(G, p.k, Q)
		for j := 0; j < n; j++ {
			p.poly.items[j] = randGen(Q)
		}
		ps[i] = p
	}

	for _, pi := range ps {
		fnx := new(big.Int).SetInt64(0)
		for _, pj := range ps {
			fnx.Add(fnx, pj.poly.calc(pi.x))
		}
		pi.fnx = fnx.Mod(fnx, Q)
	}

	cid := new(big.Int).SetInt64(0)
	for _, pi := range ps {
		x := new(big.Int).SetInt64(1)
		for _, pj := range ps {
			if pi.x.Cmp(pj.x) == 0 {
				continue
			}
			_xj := new(big.Int).Neg(pj.x)
			diff_xi_xj := new(big.Int).Sub(pi.x, pj.x)
			_diff_xi_xj := calcInverseElem(diff_xi_xj, Q)
			x.Mul(x, new(big.Int).Mul(_xj, _diff_xi_xj))
		}
		pi.ski = new(big.Int).Mul(pi.fnx, x)
		pi.ski.Mod(pi.ski, Q)
		tk.Add(tk, pi.ski)
		pi.pki = new(big.Int).Exp(G, pi.ski, Q)
		cid.Add(cid, pi.x)
	}
	hk = new(big.Int).Exp(G, tk, Q)
	hk_ := new(big.Int).SetInt64(1)
	for _, p := range ps {
		hk_.Mul(hk_, p.pki)
	}
	hk_.Mod(hk_, Q)
	fmt.Println("比较一下：", hk_.Cmp(hk))
	cid.Mod(cid, Q)

	alpha := HashBigInt(cid, hk) // 论文里的alpha
	_alpha := calcInverseElem(alpha, Q)
	Alpha := new(big.Int).SetInt64(1) // 所有节点的alpha连乘得到
	for _, p := range ps {
		p.alphai = new(big.Int).Exp(alpha, p.k, Q)
		Alpha.Mul(Alpha, p.alphai)
	}
	//Alpha.Mod(Alpha, Q)
	_Alpha := calcInverseElem(Alpha, Q)

	// 计算变色龙哈希值
	msg := []byte("画江湖之不良人")
	sigma := HashBytes(msg)
	originHash := new(big.Int).Exp(G, sigma, Q)
	originHash.Mul(originHash, new(big.Int).Exp(alpha, HashBytes(msg), Q))
	originHash.Mod(originHash, Q)
	fmt.Println("原始哈希值：", originHash)

	originRandom := &random{
		r1: new(big.Int).Exp(G, sigma, Q),
		r2: new(big.Int).Exp(hk, sigma, Q),
	}
	fmt.Println("原始随机系数：", originRandom)

	// 计算哈希碰撞

	segs := make([]*segSig, n)
	_msg := []byte("流浪地球")
	e := new(big.Int).Sub(HashBytes(msg), HashBytes(_msg))  // msg-_msg
	_e := new(big.Int).Sub(HashBytes(_msg), HashBytes(msg)) // _msg-msg
	inverse := false
	if _e.Cmp(new(big.Int).SetInt64(0)) < 0 {
		inverse = true
	}

	for i := 0; i < n; i++ {
		// 节点i
		si := new(big.Int).Add(new(big.Int).Mul(ps[i].ski, e), ps[i].k)
		di := new(big.Int)
		if si.Cmp(new(big.Int).SetInt64(0)) < 0 {
			_si := new(big.Int).Neg(si)
			inverseAlpha := calcInverseElem(alpha, Q)
			di = new(big.Int).Exp(inverseAlpha, _si, Q)
		} else {
			di = new(big.Int).Exp(alpha, si, Q)
		}

		_xi := new(big.Int).Exp(G, si, Q)
		if inverse {
			_pki := calcInverseElem(ps[i].pki, Q)
			_xi.Mul(_xi, new(big.Int).Exp(_pki, e, Q))
		} else {
			_xi.Mul(_xi, new(big.Int).Exp(ps[i].pki, _e, Q))
		}
		_xi.Mod(_xi, Q)
		if _xi.Cmp(ps[i].x) == 0 {
			fmt.Printf("节点%d为计算哈希碰撞贡献的密钥分片信息是正确的\n", i)
			segs[i] = &segSig{
				s: si,
				d: di,
			}
		} else {
			fmt.Printf("节点%d为计算哈希碰撞贡献的密钥分片信息是错误的\n", i)
		}
	}

	c := new(big.Int).SetInt64(1)
	for _, seg := range segs {
		c.Mul(c, seg.d)
	}
	c.Mul(c, _Alpha)

	redactRandom := &random{}
	if e.Cmp(new(big.Int).SetInt64(0)) < 0 {
		redactRandom.r1 = new(big.Int).Mul(originRandom.r1, new(big.Int).Exp(_alpha, _e, Q))
	} else {
		redactRandom.r1 = new(big.Int).Mul(originRandom.r1, new(big.Int).Exp(alpha, e, Q))
	}
	redactRandom.r1.Mod(redactRandom.r1, Q)
	redactRandom.r2 = new(big.Int).Mul(originRandom.r2, c)
	redactRandom.r2.Mod(redactRandom.r2, Q)

	fmt.Println("碰撞随机系数：", redactRandom)

	redactHash := new(big.Int).Mul(redactRandom.r1, new(big.Int).Exp(alpha, HashBytes(_msg), Q))
	redactHash.Mod(redactHash, Q)
	fmt.Println("碰撞哈希：", redactHash)
	// 集中式验证
	//ver := new(big.Int).Exp(redactRandom.r1, tk, Q)
	//if ver.Cmp(redactRandom.r2) == 0 && redactHash.Cmp(originHash) == 0 {
	//	fmt.Println("计算哈希碰撞成功")
	//} else {
	//	fmt.Println("计算哈希碰撞失败")
	//}

	// 分布式验证
	ver := new(big.Int).SetInt64(1)
	for _, p := range ps {
		ver.Mul(ver, new(big.Int).Exp(redactRandom.r1, p.ski, Q))
	}
	ver.Mod(ver, Q)
	if ver.Cmp(redactRandom.r2) == 0 && redactHash.Cmp(originHash) == 0 {
		fmt.Println("计算哈希碰撞成功")
	} else {
		fmt.Println("计算哈希碰撞失败")
	}

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

// randGen 方法接受一个大整数upper作为输入参数，随机生成一个值不大于upper的大整数。
func randGen(upper *big.Int) *big.Int {
	randomBig, err := rand.Int(rand.Reader, upper)
	if err != nil {
		panic(err)
	}
	return randomBig
}
