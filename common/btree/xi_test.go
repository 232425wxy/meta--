package btree

import (
	"fmt"
	"testing"
)

func test() *node {
	n := new(node)
	var nn *node
	nn = n
	fmt.Println(&n)
	fmt.Println(&nn)
	return nn
}

func TestName(t *testing.T) {
	nn := test()
	fmt.Println(&nn)
}
