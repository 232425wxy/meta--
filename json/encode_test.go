package json

import (
	"fmt"
	"reflect"
	"testing"
)

func f(x interface{}) {
	rVal := reflect.ValueOf(x)
	rTyp := reflect.TypeOf(x)

	fmt.Println("val is ptr:", rVal.Kind() == reflect.Ptr)
	fmt.Println("typ is ptr:", rTyp.Kind() == reflect.Ptr)
}

func TestReflectPtr(t *testing.T) {
	ptrA := &structJSON{}
	f(ptrA)

	elemA := structJSON{}
	f(elemA)
}

func TestReflectArray(t *testing.T) {
	s := [3]string{"1", "2", "3"}
	rVal := reflect.ValueOf(s)
	if rVal.IsNil() { // panic: reflect: call of reflect.Value.IsNil on array Value [recovered]
		t.Log("true")
	}
}
