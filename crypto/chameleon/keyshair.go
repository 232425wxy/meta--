package chameleon

import (
	"crypto/rand"
	"math/big"
)

// Participant ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Participant 定义了分布式变色龙哈希函数的秘密陷门分片持有者的信息。
type Participant struct {
	k          *big.Int              // 秘密值
	x          *big.Int              // 身份标识符：G**k
	fn         *polynomial           // 随机构造的多项式
	ski        *big.Int              // 变色龙哈希函数的陷门分片
	pki        *big.Int              // 节点的公钥：G**ski
	lambda     *big.Int              // f1(xi)+...+fn(xi) = F(xi)
	n, t       int                   // 节点总数和阈值
	neighbours map[string]*neighbour // ID string -> neighbour
}

// NewParticipant ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewParticipant 实例化一个分布式变色龙哈希函数的参与者，该方法在区块链系统启动之初调用，为了方便，
// 我们应该在系统刚启动时就确定系统中会有多少个节点，然后阈值应该是多少，因此该方法会初始化节点的密钥参
// 数，并且会根据阈值随机构造多项式。
func NewParticipant(n, t int) *Participant {
	participant := &Participant{n: n, t: t}
	participant.k = randGen(Q) // 随机选择秘密参数
	participant.lambda = new(big.Int).Set(Zero)
	participant.x = new(big.Int).Exp(G, participant.k, Q) // 生成身份标识符
	// 随机构造t-1次多项式
	participant.fn = &polynomial{items: make(map[int]*big.Int)}
	for order := 0; order < t; order++ {
		participant.fn.items[order] = randGen(Q)
	}
	participant.neighbours = make(map[string]*neighbour)
	return participant
}

type polynomial struct {
	items map[int]*big.Int
}

// calc ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// calc 接受一个参数x，然后计算多项式f(x)的值。
func (poly *polynomial) calc(x *big.Int) *big.Int {
	res := new(big.Int).SetInt64(0)
	for order, item := range poly.items {
		exp := new(big.Int).Exp(x, new(big.Int).SetInt64(int64(order)), Q)
		exp.Mul(exp, item)
		res.Add(res, exp)
	}
	return res.Mod(res, Q)
}

// neighbour ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// neighbour 在系统启动之初，所有节点对外广播自己的身份标识符，然后分别用自己的多项式计算其他节点身份标识符对应的
// 函数值，并将其秘密发送给对应节点。
type neighbour struct {
	id        *big.Int
	publicKey *big.Int
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// randGen ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// randGen 方法接受一个大整数upper作为输入参数，随机生成一个值不大于upper的大整数。
func randGen(upper *big.Int) *big.Int {
	randomBig, err := rand.Int(rand.Reader, upper)
	if err != nil {
		panic(err)
	}
	return randomBig
}
