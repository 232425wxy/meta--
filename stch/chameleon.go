package stch

import "math/big"

type polynomial struct {
	items map[int]*big.Int
}

type Chameleon struct {
	k              *big.Int
	x              *big.Int
	fn             *polynomial
	n              int // 分布式成员数量
	signalToSendFX chan struct{}
}

func NewChameleon(n int) *Chameleon {
	ch := &Chameleon{}
	ch.k, ch.x = GenerateKAndX()
	ch.fn = &polynomial{items: make(map[int]*big.Int)}
	ch.n = n
	ch.signalToSendFX = make(chan struct{}, 1)
	ch.GenerateFn(n)
	return ch
}

func (ch *Chameleon) GenerateFn(num int) {
	for i := 0; i < num; i++ {
		ch.fn.items[i] = GeneratePolynomialItem()
	}
}

func (ch *Chameleon) GetX() *big.Int {
	return ch.x
}

func (ch *Chameleon) CanSendFX() <-chan struct{} {
	return ch.signalToSendFX
}
