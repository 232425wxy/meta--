package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 项目级全局函数

// Decode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Decode 反序列化数据。
func Decode(bz []byte, x interface{}) error {
	if len(bz) == 0 {
		return errors.New("cannot decode empty bytes")
	}

	rVal := reflect.ValueOf(x)
	if rVal.Kind() != reflect.Ptr {
		return errors.New("cannot decode into a non-pointer value")
	}
	rVal = rVal.Elem()
	if typeRegister.name(rVal.Type()) != "" {
		return decodeInterface(bz, rVal)
	}
	return decodeAll(bz, rVal)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义不可导出的工具函数

// decodeAll ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// decodeAll 这里用于递归调用。
func decodeAll(bz []byte, rVal reflect.Value) error {
	// 因为在最开始的时候，已经调用了rVal = rVal.Elem()，所以此时的rVal必定是可取地址的。
	if !rVal.CanAddr() {
		return errors.New("should decode into addressable value")
	}
	if bytes.Equal(bz, []byte("null")) {
		rVal.Set(reflect.Zero(rVal.Type()))
		return nil
	}
	for rVal.Kind() == reflect.Ptr {
		if rVal.IsNil() {
			// 说明此时rVal还是一个指针，reflect.New(x)方法会生成一个指向x的指针，所以如果此时
			// 不调用rVal.Type().Elem()，那么会生成一个指向rVal的指针，也就是指针的指针，显然
			// 将其赋值给rVal是不正确的。
			rVal.Set(reflect.New(rVal.Type().Elem()))
		}
		rVal = rVal.Elem()
	}
	if rVal.Addr().Type().Implements(jsonUnmarshalerType) {
		// 实现 json.UnmarshalJSON 方法的必须得是指针，所以这里调用了 Addr() 方法来取地址
		return rVal.Addr().Interface().(json.Unmarshaler).UnmarshalJSON(bz)
	}

	switch rVal.Type().Kind() {
	case reflect.Interface:
		return decodeInterface(bz, rVal)
	case reflect.Array, reflect.Slice:
		return decodeArrayOrSlice(bz, rVal)
	case reflect.Map:
		return decodeMap(bz, rVal)
	case reflect.Struct:
		return decodeStruct(bz, rVal)
	default:
		return decodeStdlib(bz, rVal)
	}
}

// decodeInterface ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// decodeInterface 方法接受两个参数：(bz []byte, rVal reflect.Value)，该方法在 rVal 被判定为接口，或者
// rVal背后的数据类型已经被注册时被调用，用来对接口类型的数据进行反序列化，其本质是对实现某个接口的结构体数据进行
// 反序列化，我们知道在序列化结构体数据时，得到的序列化结果形如{"type":xxx,"value":xxx}，所以我们需要定义一个
// 结构体来解析序列化数据里的"type"和"value"。
func decodeInterface(bz []byte, rVal reflect.Value) error {
	if !rVal.CanAddr() {
		return errors.New("interface value is not addressable")
	}
	wrapper := &interfaceWrapper{}
	err := json.Unmarshal(bz, wrapper)
	if err != nil {
		return err
	}
	if wrapper.Type == "" {
		return errors.New("interface type cannot be empty")
	}
	if len(wrapper.Value) == 0 {
		return errors.New("interface value cannot be empty")
	}

	for rVal.Kind() == reflect.Ptr {
		if rVal.IsNil() {
			// 不设置一个值进去的化，将来调用.Type()方法会panic
			rVal.Set(reflect.New(rVal.Type().Elem()))
		}
		rVal = rVal.Elem()
	}
	rTyp, returnPtr := typeRegister.lookup(wrapper.Type)
	if rTyp == nil {
		return fmt.Errorf("unknown type %q", wrapper.Type)
	}
	newRVal := reflect.New(rTyp)
	newRValElem := newRVal.Elem()
	if err = decodeAll(wrapper.Value, newRValElem); err != nil {
		return err
	}
	if rVal.Type().Kind() == reflect.Interface && returnPtr {
		if !newRVal.Type().AssignableTo(rVal.Type()) {
			return fmt.Errorf("invalid type %q for this value", wrapper.Type)
		}
		rVal.Set(newRVal)
	} else {
		if !newRValElem.Type().AssignableTo(rVal.Type()) {
			return fmt.Errorf("invalid type %q for this value", wrapper.Type)
		}
		rVal.Set(newRValElem)
	}
	return nil
}

// decodeArrayOrSlice ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// decodeArrayOrSlice 反序列化数组或者切片数据，在反序列化时，主要根据 reflect.Value 所指示的数据类型
// 来选择反序列化的方法。
func decodeArrayOrSlice(bz []byte, rVal reflect.Value) error {
	if !rVal.CanAddr() {
		return fmt.Errorf("list value is not addressable")
	}
	var rawSlice []json.RawMessage
	if err := json.Unmarshal(bz, &rawSlice); err != nil {
		return err
	}
	if rVal.Type().Kind() == reflect.Slice {
		rVal.Set(reflect.MakeSlice(reflect.SliceOf(rVal.Type().Elem()), len(rawSlice), len(rawSlice)))
	}
	if rVal.Len() != len(rawSlice) {
		return fmt.Errorf("got list of %v elements, expected %v", len(rawSlice), rVal.Len())
	}
	for i, raw := range rawSlice {
		if err := decodeAll(raw, rVal.Index(i)); err != nil {
			return err
		}
	}
	//if rVal.Type().Kind() == reflect.Slice && rVal.Len() == 0 {
	//	rVal.Set(reflect.Zero(rVal.Type()))
	//}
	return nil
}

// decodeMap ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// decodeMap 反序列化map类型数据。
func decodeMap(bz []byte, rVal reflect.Value) error {
	if !rVal.CanAddr() {
		return errors.New("map value is not addressable")
	}
	rawMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(bz, rawMap); err != nil {
		return err
	}
	if rVal.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("map key must be string, but got %q", rVal.Type().Key().Kind())
	}
	rVal.Set(reflect.MakeMapWithSize(rVal.Type(), len(rawMap)))
	for key, value := range rawMap {
		elem := reflect.New(rVal.Type().Elem()).Elem()
		if err := decodeAll(value, elem); err != nil {
			return err
		}
		rVal.SetMapIndex(reflect.ValueOf(key), elem)
	}
	return nil
}

// decodeStruct ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// decodeStruct 反序列化结构体。
func decodeStruct(bz []byte, rVal reflect.Value) error {
	if !rVal.CanAddr() {
		return fmt.Errorf("struct value is not addressable")
	}
	sInfo := makeStructInfo(rVal.Type())
	rawStruct := make(map[string]json.RawMessage)
	if err := json.Unmarshal(bz, &rawStruct); err != nil {
		return err
	}
	for i, fInfo := range sInfo.fields {
		if !fInfo.ignored {
			field := rVal.Field(i)
			value := rawStruct[fInfo.jsonName]
			if len(value) > 0 {
				if err := decodeAll(value, field); err != nil {
					return err
				}
			} else if !fInfo.omitEmpty {
				// 没有设置omitempty，那么就给它赋值默认的空值吧
				field.Set(reflect.Zero(field.Type()))
			}
		}
	}
	return nil
}

// decodeStdlib ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// decodeStdlib 方法接受两个参数作为输入参数：([]byte, reflect.Value)，该方法调用标准的 json.Decode
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

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

// interfaceWrapper ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// interfaceWrapper 用来解析结构体数据序列化的结果，因为对结构体数据序列化得到的结果形如{"type":xxx,"value":xxx}。
type interfaceWrapper struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的包级变量

// jsonUnmarshalerType ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// jsonUnmarshalerType 这里我们利用new方法实例化一个 json.Unmarshaler 接口对象，但是得到的是一个指针，
// 所以我们还要再调用Elem()方法，来获得指针指向的接口。
var jsonUnmarshalerType = reflect.TypeOf(new(json.Unmarshaler)).Elem()
