package chameleon

import (
	"crypto/sha256"
	"fmt"
	"math/big"
)

var (
	hk *big.Int = new(big.Int)
	tk *big.Int = new(big.Int)
)

func keyGen() {
	var err error
	tk = randGen(Q)
	if err != nil {
		panic(err)
	}
	hk.Exp(G, tk, P)
}

func Hash(message []byte, r, s *big.Int) []byte {
	// 生成 message || r.Bytes() 的哈希值
	h := sha256.New()
	h.Write(message)
	h.Write(r.Bytes())
	eBig := new(big.Int).SetBytes(h.Sum(nil))

	// 计算 hk**eBig mod P
	hk_eBig := new(big.Int).Exp(hk, eBig, P)
	// 计算 G**s mod P
	g_s := new(big.Int).Exp(G, s, P)
	// 计算 hk**eBig * G**s mod P
	tmpBig := new(big.Int).Mul(hk_eBig, g_s)
	tmpBig.Mod(tmpBig, P)
	// 计算r - hk**eBig * G**s mod P
	hBig := new(big.Int).Sub(r, tmpBig)
	hBig.Mod(hBig, Q)
	return hBig.Bytes()
}

func Forge(originHash []byte, message2 []byte) (*big.Int, *big.Int) {
	kBig := randGen(Q)
	hBig := new(big.Int).SetBytes(originHash)

	// 计算g**k + h
	tmpBig := new(big.Int).Exp(G, kBig, P)
	r2Big := new(big.Int).Add(hBig, tmpBig)
	r2Big.Mod(r2Big, Q)

	// 生成 message2 || r2.Bytes() 的哈希
	newHash := sha256.New()
	newHash.Write(message2)
	newHash.Write(r2Big.Bytes())

	// 计算新的 e
	eBig := new(big.Int).SetBytes(newHash.Sum(nil))

	tmpBig.Mul(eBig, tk)
	tmpBig.Mod(tmpBig, Q)
	s2Big := new(big.Int).Sub(kBig, tmpBig)
	s2Big.Mod(s2Big, Q)

	return r2Big, s2Big
}

func run() {
	keyGen()
	r1 := randGen(Q)
	s1 := randGen(Q)
	message1 := []byte("太难了！")
	h := Hash(message1, r1, s1)
	fmt.Println("origin Hash value:", h)

	message2 := []byte("给我成！")
	r2, s2 := Forge(h, message2)
	newHash := Hash(message2, r2, s2)
	fmt.Println("current Hash value:", newHash)
}
