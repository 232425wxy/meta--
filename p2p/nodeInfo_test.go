package p2p

import (
	"fmt"
	"testing"
)

type Dog struct {
}

func (d *Dog) Eat() {
	fmt.Println("狗在吃东西")
}

func (d Dog) Run() {
	fmt.Println("狗在跑")
}

func TestPtr(t *testing.T) {
	d1 := &Dog{}
	d2 := Dog{}

	d1.Eat()
	d1.Run()

	d2.Eat()
	d2.Run()
}
