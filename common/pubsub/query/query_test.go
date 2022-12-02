package query

import (
	"fmt"
	"testing"
)

func TestRegexpNum(t *testing.T) {
	res := numRegex.Find([]byte("7625165 87687hello 9.99yuan"))
	t.Log(fmt.Sprintf("res:%s", res))
}
