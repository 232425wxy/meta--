package json

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type structJSON struct {
	a int    `json:"A"`
	B string `json:"b,omitempty"`
	C byte   `json:"-"`
}

func TestMakeStructInfo(t *testing.T) {
	typ := reflect.TypeOf(structJSON{})
	sInfo := makeStructInfo(typ)
	for i := 0; i < len(sInfo.fields); i++ {
		info := sInfo.fields[i]
		fmt.Println("-----------------------------")
		fmt.Println("name:", info.jsonName)
		fmt.Println("ignored:", info.ignored)
		fmt.Println("omitEmpty:", info.omitEmpty)
	}

	s := structJSON{a: 199, B: "hello"}
	bz, err := json.Marshal(s)
	assert.Nil(t, err)
	dst := &structJSON{}
	err = json.Unmarshal(bz, dst)
	assert.Nil(t, err)
	fmt.Println(*dst)
}
