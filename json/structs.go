package json

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

// cache ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// cache 是一个包级的全局变量，用来在程序运行期间存储被json包解析过的结构体的structInfo信息。
var cache = &structInfoCache{mapping: make(map[reflect.Type]*structInfo)}

// structInfoCache ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// structInfoCache 结构存储了所有已经解析过的struct的 structInfo。
type structInfoCache struct {
	sync.RWMutex
	mapping map[reflect.Type]*structInfo
}

// get ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// get 根据给定的struct的reflect.Type，返回对应的structInfo。
func (c *structInfoCache) get(typ reflect.Type) *structInfo {
	c.RLock()
	defer c.RUnlock()
	return c.mapping[typ]
}

// set ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// set 接受两个输入参数，给定的struct的reflect.Type和对应的structInfo。
func (c *structInfoCache) set(typ reflect.Type, info *structInfo) {
	c.Lock()
	defer c.Unlock()
	c.mapping[typ] = info
}

// structInfo ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// structInfo 结构体用来存储结构体中所有字段的`json:?`信息。
type structInfo struct {
	fields []*fieldInfo
}

// fieldInfo ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// fieldInfo 结构体用来存储结构体里单个字段的`json:?`信息。
type fieldInfo struct {
	jsonName  string
	omitEmpty bool
	ignored   bool
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// makeStructInfo ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// makeStructInfo 方法接受一个struct的reflect.Type，然后对其进行解析，获取该struct对应的structInfo。
func makeStructInfo(rTyp reflect.Type) *structInfo {
	if rTyp.Kind() != reflect.Struct {
		panic(fmt.Sprintf("can't make struct info for non-struct value %v", rTyp))
	}
	if info := cache.get(rTyp); info != nil {
		return info
	}
	// 到目前为止还未遇到rTyp所指向的struct
	fields := make([]*fieldInfo, 0, rTyp.NumField())
	for i := 0; i < cap(fields); i++ {
		fieldI := rTyp.Field(i)
		info := &fieldInfo{
			jsonName:  fieldI.Name,
			omitEmpty: false,
			// 不可导出的字段会被忽略掉
			ignored: !fieldI.IsExported(),
		}
		tag := fieldI.Tag.Get("json")
		if tag == "-" {
			info.ignored = true
		} else if tag != "" {
			opts := strings.Split(tag, ",")
			if opts[0] != "" {
				info.jsonName = opts[0]
			}
			for _, opt := range opts[1:] {
				if opt == "omitempty" {
					info.omitEmpty = true
				}
			}
		}
		fields = append(fields, info)
	}
	sInfo := &structInfo{fields: fields}
	cache.set(rTyp, sInfo)
	return sInfo
}
