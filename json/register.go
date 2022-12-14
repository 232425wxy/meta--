package json

import (
	"fmt"
	"reflect"
	"sync"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 项目级全局函数

// RegisterType ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// RegisterType 方法接受两个参数，一个是某类型的实例，另一个是为该类型取的名字，然后将其注册到 pbtypes 结构体里。
func RegisterType(x interface{}, name string) {
	if x == nil {
		panic("cannot register nil type")
	}
	if name == "" {
		panic(fmt.Sprintf("cannot register the type %v with empty name", x))
	}
	typ := reflect.ValueOf(x).Type()
	err := typeRegister.register(name, typ)
	if err != nil {
		panic(err)
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

// typeRegister ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// typeRegister 实例用来存储程序运行期间所有需要json序列化的类型信息。
var typeRegister = &types{
	byType: make(map[reflect.Type]*typeInfo),
	byName: make(map[string]*typeInfo),
}

// types ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// types 结构体存储程序运行期间遇到的所有类型的信息。
type types struct {
	sync.RWMutex
	byType map[reflect.Type]*typeInfo
	byName map[string]*typeInfo
}

// register ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// register 接受两个变量做为输入参数，一个是type的类型名，另一个是type的reflect.Type，该方法就是将
// 给定的两个参数注册到 types 结构体的两个map里。
func (t *types) register(name string, typ reflect.Type) error {
	returnPtr := false
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		returnPtr = true
	}
	info := &typeInfo{
		name:      name,
		rTyp:      typ,
		returnPtr: returnPtr,
	}
	t.Lock()
	defer t.Unlock()
	if _, ok := t.byType[info.rTyp]; ok {
		return fmt.Errorf("the type %v is already registered", info.rTyp)
	}
	if _, ok := t.byName[info.name]; ok {
		return fmt.Errorf("the type with name %s is already registered", info.name)
	}
	t.byName[info.name] = info
	t.byType[info.rTyp] = info
	return nil
}

// lookup ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// lookup 给定一个type的name，然后根据name在 types 处寻找注册过的对应类型的 reflect.Type，并将
// reflect.Type 作为第一个返回值返回，第二个返回参数则代表当初注册的类型是否是一个指针。
func (t *types) lookup(name string) (reflect.Type, bool) {
	t.RLock()
	defer t.RUnlock()
	if info, ok := t.byName[name]; ok {
		return info.rTyp, info.returnPtr
	}
	return nil, false
}

// name ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// name 方法接受一个type的reflect.Type，然后拿着它去 types 的byType里寻找曾经注册过的类型信息，
// 如果找到的话，则返回当初注册时为该类型取的名字，否则返回""。
func (t *types) name(typ reflect.Type) string {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	t.RLock()
	defer t.RUnlock()
	if info, ok := t.byType[typ]; ok {
		return info.name
	}
	return ""
}

// typeInfo ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// typeInfo 结构体存储单个类型的具体信息。
type typeInfo struct {
	name      string
	rTyp      reflect.Type
	returnPtr bool
}
