package stch

import (
	"encoding/json"
	"math/big"
	"os"
)

type KeyPoly struct {
	K    *big.Int    `json:"k"`
	Poly *polynomial `json:"poly"`
}

func NewKP(n int) *KeyPoly {
	kp := &KeyPoly{}
	kp.K, _ = GenerateKAndX()
	kp.Poly = &polynomial{Items: make(map[int]*big.Int)}
	for i := 0; i < n; i++ {
		kp.Poly.Items[i] = GeneratePolynomialItem()
	}
	return kp
}

func (kp *KeyPoly) Save(path string) {
	bz, err := json.Marshal(kp)
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	file.Write(bz)
}

func LoadInitConfig(path string) *KeyPoly {
	kp := &KeyPoly{}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	content := make([]byte, 4096)
	n, err := file.Read(content)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content[:n], kp)
	if err != nil {
		panic(err)
	}
	return kp
}
