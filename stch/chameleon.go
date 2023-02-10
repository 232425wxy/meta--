package stch

import "math/big"

type polynomial struct {
	items map[int]*big.Int
}

type Chameleon struct {
	k  *big.Int
	x  *big.Int
	fn *polynomial
}

func NewChameleon() *Chameleon {
	ch := &Chameleon{}
	ch.k, ch.x = GenerateKAndX()
	ch.fn = &polynomial{items: make(map[int]*big.Int)}
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
