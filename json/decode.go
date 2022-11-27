package json

import (
	"encoding/json"
	"errors"
	"reflect"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义不可导出的工具函数

// decodeAll ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//  ---------------------------------------------------------
// decodeAll
//func decodeAll(bz []byte, rVal reflect.Value) error {
//	// 因为在最开始的时候，已经调用了rVal = rVal.Elem()，所以此时的rVal必定是可取地址的。
//	if !rVal.CanAddr() {
//		return errors.New("should decode into addressable value")
//	}
//	if bytes.Equal(bz, []byte("null")) {
//		rVal.Set(reflect.Zero(rVal.Type()))
//	}
//}

// decodeStdlib ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// decodeStdlib 方法接受两个参数作为输入参数：([]byte, reflect.Value)，该方法调用标准的 json.Unmarshal
// 方法将第一个参数里的内容反序列化到第二个参数里。
func decodeStdlib(bz []byte, rVal reflect.Value) error {
	if !rVal.CanAddr() && rVal.Kind() != reflect.Ptr {
		return errors.New("cannot decode into non-pointer object")
	}
	target := rVal
	if rVal.Kind() != reflect.Ptr {
		target = reflect.New(rVal.Type())
	}
	if err := json.Unmarshal(bz, target.Interface()); err != nil {
		return err
	}
	// rVal 不是指针，所以直接用值给它赋值，所以在这里要调用Elem()方法。
	rVal.Set(target.Elem())
	return nil
}
