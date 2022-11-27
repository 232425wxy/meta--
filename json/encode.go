package json

import (
	"encoding/json"
	"io"
	"reflect"
	"time"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// encodeAll ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// encodeAll 接受两个参数：(io.Writer, reflect.Value)，第二个参数是要编码的数据对象，递归地调用 encodeAll
// 方法，对数据对象进行序列化，然后写入到Writer里。
//func encodeAll(w io.Writer, rVal reflect.Value) error {
//	if !rVal.IsValid() {
//		return errors.New("invalid reflect value")
//	}
//	// 递归地获取到指针所指向的值
//	for rVal.Kind() == reflect.Ptr {
//		if rVal.IsNil() {
//			return writeStr(w, "nil")
//		}
//		rVal = rVal.Elem()
//	}
//	// 将时间转换为"2006-01-02T15:04:05Z07:00"格式
//	if rVal.Type() == timeType {
//		t, err := time.Parse(time.RFC3339, rVal.Interface().(time.Time).Format(time.RFC3339))
//		if err != nil {
//			return err
//		}
//		rVal = reflect.ValueOf(t)
//	}
//	if rVal.Type().Implements(jsonMarshalerType) {
//		return encodeStdlib(w, rVal.Interface())
//	} else if rVal.CanAddr() && rVal.Addr().Type().Implements(jsonMarshalerType) {
//		return encodeStdlib(w, rVal.Addr().Interface())
//	}
//
//	switch rVal.Type().Kind() {
//	case reflect.Interface:
//		return encodeInterface(w, rVal)
//	case reflect.Array, reflect.Slice:
//
//	case reflect.Map:
//
//	case reflect.Struct:
//
//	case reflect.Int64, reflect.Int:
//
//	case reflect.Uint64, reflect.Uint:
//
//	default:
//		return encodeStdlib(w, rVal.Interface())
//	}
//}

// encodeInterface ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//  ---------------------------------------------------------
// encodeInterface 如果我们需要序列化的数据对象是一个interface{}，那么我们需要获取到interface{}所
// 指向的真正的数据。
//func encodeInterface(w io.Writer, rVal reflect.Value) error {
//	for rVal.Kind() == reflect.Interface {
//		if rVal.IsNil() {
//			return writeStr(w, "nil")
//		}
//		rVal = rVal.Elem()
//	}
//	name := typeRegister.name(rVal.Type())
//	if name == "" {
//		return fmt.Errorf("cannot encode unregistered type %v", rVal.Type())
//	}
//	if err := writeStr(w, fmt.Sprintf(`{"type":%q,"value":`, name)); err != nil {
//		return err
//	}
//	if err := encodeAll(w, rVal); err != nil {
//		return err
//	}
//	return writeStr(w, "}")
//}

// encodeArrayOrSlice ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//  ---------------------------------------------------------
// encodeArrayOrSlice
//func encodeArrayOrSlice(w io.Writer, rVal reflect.Value) error {
//
//}

// encodeStdlib ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// encodeStdlib 方法接受两个参数：(io.Writer, interface{})，该方法利用golang官方的 json.Marshal
// 方法对给定的第二个入参进行序列化，然后将序列化的结果写入到给定的第一个入参中。
func encodeStdlib(w io.Writer, x interface{}) error {
	bz, err := json.Marshal(x)
	if err != nil {
		return err
	}
	_, err = w.Write(bz)
	return err
}

// writeStr ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// writeStr 方法接受两个参数：(io.Writer, string)，该方法就是将第二个字符串参数写入到第一个Writer参数里。
func writeStr(w io.Writer, str string) error {
	_, err := w.Write([]byte(str))
	return err
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的包级变量

var timeType = reflect.TypeOf(time.Time{})
var jsonMarshalerType = reflect.TypeOf(json.Marshaler(nil))
