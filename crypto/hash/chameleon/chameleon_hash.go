package chameleon

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

var p, _ = new(big.Int).SetString("c0a0f171e583d4efb262c1783a6ed6d995fc1d4eea476149cf8ea40078d27ad7", 16)

var g, _ = new(big.Int).SetString("3e446bf8e43afebc6b49bc7220d19a415c9bda8cbbcc25189c67d27b33df73cf", 16)

var q = new(big.Int).Div(new(big.Int).Sub(p, new(big.Int).SetInt64(1)), new(big.Int).SetInt64(2))

var (
	hk *big.Int = new(big.Int)
	tk *big.Int = new(big.Int)
)

func randGen(upper *big.Int) *big.Int {
	randomBig, err := rand.Int(rand.Reader, upper)
	if err != nil {
		panic(err)
	}
	return randomBig
}

func keyGen() {
	var err error
	tk = randGen(q)
	if err != nil {
		panic(err)
	}
	hk.Exp(g, tk, p)
}

func hash(message []byte, r, s *big.Int) []byte {
	// 生成 message || r.Bytes() 的哈希值
	h := sha256.New()
	h.Write(message)
	h.Write(r.Bytes())
	eBig := new(big.Int).SetBytes(h.Sum(nil))

	// 计算 hk**eBig mod p
	hk_eBig := new(big.Int).Exp(hk, eBig, p)
	// 计算 g**s mod p
	g_s := new(big.Int).Exp(g, s, p)
	// 计算 hk**eBig * g**s mod p
	tmpBig := new(big.Int).Mul(hk_eBig, g_s)
	tmpBig.Mod(tmpBig, p)
	// 计算r - hk**eBig * g**s mod p
	hBig := new(big.Int).Sub(r, tmpBig)
	hBig.Mod(hBig, q)
	return hBig.Bytes()
}

func forge(originHash []byte, message2 []byte) (*big.Int, *big.Int) {
	kBig := randGen(q)
	hBig := new(big.Int).SetBytes(originHash)

	// 计算g**k + h
	tmpBig := new(big.Int).Exp(g, kBig, p)
	r2Big := new(big.Int).Add(hBig, tmpBig)
	r2Big.Mod(r2Big, q)

	// 生成 message2 || r2.Bytes() 的哈希
	newHash := sha256.New()
	newHash.Write(message2)
	newHash.Write(r2Big.Bytes())

	// 计算新的 e
	eBig := new(big.Int).SetBytes(newHash.Sum(nil))

	tmpBig.Mul(eBig, tk)
	tmpBig.Mod(tmpBig, q)
	s2Big := new(big.Int).Sub(kBig, tmpBig)
	s2Big.Mod(s2Big, q)

	return r2Big, s2Big
}

func run() {
	keyGen()
	r1 := randGen(q)
	s1 := randGen(q)
	message1 := []byte("太难了！")
	h := hash(message1, r1, s1)
	fmt.Println("origin hash value:", h)

	message2 := []byte("给我成！")
	r2, s2 := forge(h, message2)
	newHash := hash(message2, r2, s2)
	fmt.Println("current hash value:", newHash)
}
