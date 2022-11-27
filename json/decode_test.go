package json

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestCanAddr(t *testing.T) {
	d := &Dog{}
	rVal := reflect.ValueOf(d)
	t.Log(rVal.Elem().CanAddr())
}

func TestDecodeTime(t *testing.T) {
	tt := time.Now()
	bz, err := Marshal(tt)
	assert.Nil(t, err)
	t.Log(string(bz))

	dst := &time.Time{}
	err = json.Unmarshal(bz, dst)
	assert.Nil(t, err)
	t.Log(dst.Format(time.RFC3339))
}

func TestDecodeNilMap(t *testing.T) {
	m := map[string]int{}
	rVal := reflect.ValueOf(&m)

	rVal.Elem().Set(reflect.Zero(rVal.Elem().Type()))
	t.Log(m)
}
