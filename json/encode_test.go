package json

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	RegisterType(&Cat{}, "Animal/Cat")
	RegisterType(Dog{}, "Animal/Dog")
}

type Animal interface {
	Eat(food string)
}

type Cat struct {
	Name string
	Age  uint8
}

func (c *Cat) Eat(f string) {
	fmt.Printf("%s eats %s.\n", c.Name, f)
}

type Dog struct {
	Name string
	Age  int64
}

func (d Dog) Eat(f string) {
	fmt.Printf("%s eats %s.\n", d.Name, f)
}

func TestEncode(t *testing.T) {
	testCases := map[string]struct {
		value  interface{}
		output string
	}{
		"nil": {nil, "null"},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			bz, err := Marshal(test.value)
			assert.Nil(t, err)
			assert.JSONEq(t, test.output, string(bz))
		})
	}
}
