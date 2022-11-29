package json

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestDecode(t *testing.T) {
	intPtr0 := new(int)
	*intPtr0 = 0
	intPtr1 := new(int)
	*intPtr1 = 1
	testCases := map[string]struct {
		data  string
		value interface{}
		err   string
	}{
		"bool true":                    {data: `true`, value: true, err: ""},
		"bool false":                   {data: `false`, value: false, err: ""},
		"int 0":                        {data: `0`, value: int(0), err: ""},
		"int 3":                        {data: `3`, value: int(3), err: ""},
		"int -3":                       {data: `-3`, value: int(-3), err: ""},
		"int64 0":                      {data: `0`, value: int64(0), err: ""},
		"int64 -3":                     {data: `-3`, value: int64(-3), err: ""},
		"uint 0":                       {data: `0`, value: int(0), err: ""},
		"uint 3":                       {data: `3`, value: int(3), err: ""},
		"uint64 0":                     {data: `0`, value: int64(0), err: ""},
		"uint64 3":                     {data: `3`, value: int64(3), err: ""},
		"string \"\"":                  {data: `""`, value: string(""), err: ""},
		"string \"\" error":            {data: ``, value: string(""), err: "cannot decode empty bytes"},
		"string \"hello\"":             {data: `"hello"`, value: string("hello"), err: ""},
		"array int null":               {data: "[]", value: [0]int{}, err: ""},
		"array [0]int":                 {data: `[]`, value: [0]int{}, err: ""},
		"array [2]int":                 {data: `[77,88]`, value: [2]int{77, 88}, err: ""},
		"array [2]byte":                {data: `[61,62]`, value: [2]byte{61, 62}, err: ""},
		"array [0]byte":                {data: `[]`, value: [0]byte{}, err: ""},
		"array byte []":                {data: "[]", value: [0]byte{}, err: ""},
		"array string []":              {data: "[]", value: [0]string{}, err: ""},
		"array [0]string":              {data: `[]`, value: [0]string{}, err: ""},
		"array [2]string":              {data: `["hello","golang"]`, value: [2]string{"hello", "golang"}, err: ""},
		"slice []int []":               {data: `[]`, value: []int{}, err: ""},
		"slice []int":                  {data: `[]`, value: []int{}, err: ""},
		"slice []int{1, 2, 3}":         {data: `[1,2,3]`, value: []int{1, 2, 3}, err: ""},
		"slice []string []":            {data: `[]`, value: []string{}, err: ""},
		"slice [2]string":              {data: `["hi","go"]`, value: []string{"hi", "go"}, err: ""},
		"int ptr null":                 {data: `null`, value: (*int)(nil), err: ""},
		"int ptr 0":                    {data: `0`, value: intPtr0, err: ""},
		"int ptr 1":                    {data: `1`, value: intPtr1, err: ""},
		"struct Cat":                   {data: `{"type":"Animal/Cat","value":{"Name":"tom","Age":10}}`, value: Cat{Name: "tom", Age: 10}, err: ""},
		"struct Cat ptr":               {data: `{"type":"Animal/Cat","value":{"Name":"tom","Age":10}}`, value: &Cat{Name: "tom", Age: 10}, err: ""},
		"struct Cat empty":             {data: `{"type":"Animal/Cat","value":{}}`, value: Cat{}, err: ""},
		"struct Cat ptr null":          {data: `null`, value: (*Cat)(nil), err: ""},
		"interface Animal Cat ptr":     {data: `{"type":"Animal/Cat","value":{"Name":"tom","Age":10}}`, value: Animal(&Cat{Name: "tom", Age: 10}), err: ""},
		"interface Animal array empty": {data: `[]`, value: [0]Animal{}, err: ""},
		"interface Animal *Cat Dog":    {data: `[{"type":"Animal/Cat","value":{"Name":"tom","Age":10}},{"type":"Animal/Dog","value":{"Name":"tick","Age":3}}]`, value: []Animal{&Cat{Name: "tom", Age: 10}, Dog{Name: "tick", Age: 3}}, err: ""},
		"interface Animal slice empty": {data: `[]`, value: []Animal{}, err: ""},
		"Tags":                         {data: `{"name":"name","Ignored":"ignore","tags":{"name":"child"}}`, value: Tags{Name: "name", Tags: &Tags{Name: "child"}}},
		"CustomValue":                  {data: `"CustomValue"`, value: CustomValue{}, err: ""},
		"CustomValuePtr":               {data: "CustomValuePtr", value: CustomValuePtr{X: "CustomValuePtr"}, err: ""},
		"CustomValuePtr null":          {data: "null", value: (*CustomValuePtr)(nil), err: ""},
		"float32 3.14":                 {data: `3.14`, value: float32(3.14), err: ""},
		"float32 3.14 neg":             {data: `-3.14`, value: float32(-3.14), err: ""},
		"float64 3.14":                 {data: `3.14`, value: float64(3.14), err: ""},
		"float64 3.14 neg":             {data: `-3.14`, value: float64(-3.14), err: ""},
		"float64 -3.14E+00 neg":        {data: `-3.14E+00`, value: float64(-3.14), err: ""},
		"time empty":                   {data: `"0001-01-01T00:00:00Z"`, value: time.Time{}, err: ""},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			dst := reflect.New(reflect.TypeOf(test.value)).Interface()
			err := Decode([]byte(test.data), dst)
			if test.err == "" {
				elem := reflect.ValueOf(dst).Elem().Interface()
				assert.Equal(t, test.value, elem)
			} else {
				assert.Equal(t, err.Error(), test.err)
			}
		})
	}
}

func TestName(t *testing.T) {
	d := Dog{Name: "tick", Age: 12}
	a := Animal(d)
	bz, err := Encode(a)
	assert.Equal(t, nil, err)
	t.Log(string(bz))

	dst := new(Animal)
	err = Decode(bz, dst)
	assert.Nil(t, err)
	t.Log(*dst)
}
