package json

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func init() {
	RegisterType(&Cat{}, "Animal/Cat")
	RegisterType(Dog{}, "Animal/Dog")
	//RegisterType(Tags{}, "Tags")
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

type CustomValue struct {
	X string `json:"x"`
}

func (c CustomValue) MarshalJSON() ([]byte, error) {
	return []byte(`"CustomValue"`), nil
}

func (c CustomValue) UnmarshalJSON(bz []byte) error {
	c.X = "CustomValue"
	return nil
}

type CustomValuePtr struct {
	X string
}

func (c *CustomValuePtr) MarshalJSON() ([]byte, error) {
	return []byte(`"CustomValuePtr"`), nil
}

func (c *CustomValuePtr) UnmarshalJSON(bz []byte) error {
	c.X = "CustomValuePtr"
	return nil
}

type Tags struct {
	Name      string `json:"name"`
	OmitEmpty string `json:",omitempty"`
	Ignored   string `json:"-"`
	Tags      *Tags  `json:"tags,omitempty"`
}

func TestEncode(t *testing.T) {
	testCases := map[string]struct {
		value  interface{}
		output string
	}{
		"nil":                {nil, "null"},
		"string":             {"foo", `"foo"`},
		"float32":            {float32(3.14), `3.14E+00`},
		"float32 neg":        {float32(-3.14), `-3.14E+00`},
		"float64":            {3.14, `3.14E+00`},
		"float64 neg":        {-3.14, `-3.14E+00`},
		"int":                {int(100), `100`},
		"int8":               {int8(8), `8`},
		"int16":              {int16(16), `16`},
		"int32":              {int32(32), `32`},
		"int64":              {int64(64), `64`},
		"int neg":            {int(-100), `-100`},
		"int8 neg":           {int8(-8), `-8`},
		"int16 neg":          {int8(-16), `-16`},
		"int32 neg":          {int8(-32), `-32`},
		"int64 neg":          {int8(-64), `-64`},
		"uint":               {uint(100), `100`},
		"uint8":              {uint8(8), `8`},
		"uint16":             {uint16(16), `16`},
		"uint32":             {uint16(32), `32`},
		"uint64":             {uint16(64), `64`},
		"time":               {time.Time{}, `"0001-01-01T00:00:00Z"`},
		"CustomValue":        {CustomValue{}, `"CustomValue"`},
		"CustomValue Ptr":    {&CustomValue{}, `"CustomValue"`},
		"CustomValuePtr":     {CustomValuePtr{X: "abc"}, `{"X":"abc"}`},
		"CustomValuePtr Ptr": {&CustomValuePtr{X: "abc"}, `"CustomValuePtr"`},
		"slice nil":          {[]int(nil), `[]`},
		"slice empty":        {[]int{}, `[]`},
		"slice bytes":        {[]byte{1, 2, 3}, `[1,2,3]`},
		"slice int64":        {[]int64{1, 2, 3}, `[1,2,3]`},
		"slice uint64":       {[]uint64{1, 2, 3}, `[1,2,3]`},
		"array empty":        {[3]byte{}, `[0,0,0]`},
		"string array empty": {[3]string{}, `["","",""]`},
		"array bytes":        {[3]byte{1, 2, 3}, `[1,2,3]`},
		"array uint64":       {[3]uint64{1, 2, 3}, `[1,2,3]`},
		"map empty":          {map[string]int{}, `{}`},
		"map nil":            {map[string]int(nil), `null`},
		"map string int":     {map[string]int{"abc": 2, "def": 3}, `{"abc":2,"def":3}`},
		"Cat interface":      {Animal(&Cat{Name: "tom", Age: 12}), `{"type":"Animal/Cat","value":{"Name":"tom","Age":12}}`},
		"Dog interface":      {Animal(Dog{Name: "tick", Age: 3}), `{"type":"Animal/Dog","value":{"Name":"tick","Age":3}}`},
		"Tags empty":         {Tags{}, `{"name":""}`},
		"Tags":               {Tags{Name: "name", OmitEmpty: "foo", Ignored: "no", Tags: &Tags{Name: "child"}}, `{"name":"name","OmitEmpty":"foo","tags":{"name":"child"}}`},
		"Animal slice":       {[]Animal{&Cat{Name: "tom", Age: 12}, Dog{Name: "tick", Age: 3}}, `[{"type":"Animal/Cat","value":{"Name":"tom","Age":12}},{"type":"Animal/Dog","value":{"Name":"tick","Age":3}}]`},
		"Animal array":       {[2]Animal{&Cat{Name: "tom", Age: 12}, Dog{Name: "tick", Age: 3}}, `[{"type":"Animal/Cat","value":{"Name":"tom","Age":12}},{"type":"Animal/Dog","value":{"Name":"tick","Age":3}}]`},
		"bool false":         {false, `false`},
		"bool true":          {true, `true`},
		"struct Cat":         {value: Cat{Name: "tom", Age: 10}, output: `{"type":"Animal/Cat","value":{"Name":"tom","Age":10}}`},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			bz, err := Encode(test.value)
			assert.Nil(t, err)
			assert.Equal(t, test.output, string(bz))
		})
	}
}

func TestInterface(t *testing.T) {
	c := Animal(&Cat{})
	rVal := reflect.ValueOf(c)
	t.Log(rVal.Kind() == reflect.Interface)
	t.Log(rVal.Type().Kind() == reflect.Ptr)
	t.Log(rVal.Type().Kind() == reflect.Struct)
	name := typeRegister.name(rVal.Type())
	t.Log(name)
}
