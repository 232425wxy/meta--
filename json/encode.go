package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API项目级全局函数

// Encode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Encode 方法接受一个变量对象，对该变量进行序列化，如果是自定义的结构体，需要先调用 RegisterType
// 方法注册该结构体。
func Encode(x interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	if x == nil {
		err := writeStr(buffer, "null")
		return buffer.Bytes(), err
	}
	rVal := reflect.ValueOf(x)
	if typeRegister.name(rVal.Type()) != "" {
		err := encodeInterface(buffer, rVal)
		return buffer.Bytes(), err
	}
	err := encodeAll(buffer, rVal)
	return buffer.Bytes(), err
}

func EncodeIndent(x interface{}, prefix, indent string) ([]byte, error) {
	bz, err := Encode(x)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if err = json.Indent(buf, bz, prefix, indent); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// encodeAll ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// encodeAll 接受两个参数：(io.Writer, reflect.Value)，第二个参数是要编码的数据对象，递归地调用 encodeAll
// 方法，对数据对象进行序列化，然后写入到Writer里。
func encodeAll(w io.Writer, rVal reflect.Value) error {
	if !rVal.IsValid() {
		return errors.New("invalid reflect value")
	}
	// 递归地获取到指针所指向的值
	for rVal.Kind() == reflect.Ptr {
		if rVal.IsNil() {
			return writeStr(w, "null")
		}
		rVal = rVal.Elem()
	}
	// 将时间转换为"2006-01-02T15:04:05Z07:00"格式
	if rVal.Type() == timeType {
		t, err := time.Parse(time.RFC3339, rVal.Interface().(time.Time).Format(time.RFC3339))
		if err != nil {
			return err
		}
		rVal = reflect.ValueOf(t)
	}
	if rVal.Type().Implements(jsonMarshalerType) {
		return encodeStdlib(w, rVal.Interface())
	} else if rVal.CanAddr() && rVal.Addr().Type().Implements(jsonMarshalerType) {
		return encodeStdlib(w, rVal.Addr().Interface())
	}

	switch rVal.Type().Kind() {
	case reflect.Interface:
		return encodeInterface(w, rVal)
	case reflect.Array, reflect.Slice:
		return encodeArrayOrSlice(w, rVal)
	case reflect.Map:
		return encodeMap(w, rVal)
	case reflect.Struct:
		return encodeStruct(w, rVal)
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		return writeStr(w, strconv.FormatInt(rVal.Int(), 10))
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		return writeStr(w, strconv.FormatUint(rVal.Uint(), 10))
	case reflect.Float64, reflect.Float32:
		return writeStr(w, strconv.FormatFloat(rVal.Float(), 'E', -1, 32<<int(rVal.Type().Kind()-reflect.Float32)))
	default:
		return encodeStdlib(w, rVal.Interface())
	}
}

// encodeInterface ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// encodeInterface 如果我们需要序列化的数据对象是一个interface{}，那么我们需要获取到interface{}所
// 指向的真正的数据。
func encodeInterface(w io.Writer, rVal reflect.Value) error {
	for rVal.Kind() == reflect.Interface {
		if rVal.IsNil() {
			return writeStr(w, "null")
		}
		rVal = rVal.Elem()
	}
	name := typeRegister.name(rVal.Type())
	if name == "" {
		return fmt.Errorf("cannot encode unregistered type %v", rVal.Type())
	}
	if err := writeStr(w, fmt.Sprintf(`{"type":%q,"value":`, name)); err != nil {
		return err
	}
	if err := encodeAll(w, rVal); err != nil {
		return err
	}
	return writeStr(w, "}")
}

// encodeArrayOrSlice ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// encodeArrayOrSlice 序列化数组或者切片数据，数组和切片里的元素会通过递归方式来找到对应的序列化方法。
func encodeArrayOrSlice(w io.Writer, rVal reflect.Value) error {
	// array不能调用IsNil()方法
	if rVal.Kind() == reflect.Slice && rVal.IsNil() {
		return writeStr(w, "[]")
	}

	length := rVal.Len()
	if err := writeStr(w, "["); err != nil {
		return err
	}
	for i := 0; i < length; i++ {
		if err := encodeAll(w, rVal.Index(i)); err != nil {
			return err
		}
		if i < length-1 {
			if err := writeStr(w, ","); err != nil {
				return err
			}
		}
	}
	return writeStr(w, "]")
}

// encodeMap ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// encodeMap 方法用于序列化map类型的数据，这里要求map的键必须是string类型的，然后map的值会通过递归
// 方式找到对应的序列化方法。
func encodeMap(w io.Writer, rVal reflect.Value) error {
	if rVal.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("encode map, needs string key, but got %v", rVal.Type().Key().Kind())
	}
	if rVal.IsNil() {
		return writeStr(w, "null")
	}
	if err := writeStr(w, "{"); err != nil {
		return err
	}
	keys := rVal.MapKeys()
	length := len(keys)
	for i := 0; i < length; i++ {
		key := keys[i]
		val := rVal.MapIndex(key)
		if err := encodeStdlib(w, key.Interface()); err != nil {
			return err
		}
		if err := writeStr(w, ":"); err != nil {
			return err
		}
		if err := encodeAll(w, val); err != nil {
			return err
		}
		if i < length-1 {
			if err := writeStr(w, ","); err != nil {
				return err
			}
		}
	}
	return writeStr(w, "}")
}

// encodeStruct ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// encodeStruct 方法用于序列化结构体对象，结构体的字段名由于是string类型，可以直接使用标准方法进行序列化，
// 字段所存储的值由于类型无法确定，需要在运行期间递归地找到对应的序列化方法。
func encodeStruct(w io.Writer, rVal reflect.Value) error {
	sInfo := makeStructInfo(rVal.Type())
	if err := writeStr(w, "{"); err != nil {
		return err
	}
	length := len(sInfo.fields)
	writeComma := false
	for i := 0; i < length; i++ {
		field := rVal.Field(i)
		if sInfo.fields[i].ignored || (sInfo.fields[i].omitEmpty && field.IsZero()) {
			continue
		}
		if writeComma {
			if err := writeStr(w, ","); err != nil {
				return err
			}
		}
		if err := encodeStdlib(w, sInfo.fields[i].jsonName); err != nil {
			return err
		}
		if err := writeStr(w, ":"); err != nil {
			return err
		}
		if err := encodeAll(w, field); err != nil {
			return err
		}
		writeComma = true
	}
	return writeStr(w, "}")
}

// encodeStdlib ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// encodeStdlib 方法接受两个参数：(io.Writer, interface{})，该方法利用golang官方的 json.Encode
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

// jsonMarshaler ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// jsonMarshaler 这里我们利用new方法实例化一个 json.Marshaler 接口对象，但是得到的是一个指针，
// 所以我们还要再调用Elem()方法，来获得指针指向的接口。
var jsonMarshalerType = reflect.TypeOf(new(json.Marshaler)).Elem()
