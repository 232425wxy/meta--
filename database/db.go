package database

import (
	"fmt"
	"strings"
)

type DB interface {
}

type BackendType string

const (
	// GoLevelDBBackend ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
	//
	//	---------------------------------------------------------
	//
	// GoLevelDBBackend 可以实现将数据持久化到磁盘上。
	GoLevelDBBackend BackendType = "goleveldb"

	// MemDBBackend ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
	//
	//	---------------------------------------------------------
	//
	// MemDBBackend 只能将数据暂存在内存里。
	MemDBBackend BackendType = "memdb"
)

type creator func(name string, dir string) (DB, error)

var backends = map[BackendType]creator{}

// registerDBCreator ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// registerDBCreator 注册一个创建指定类型数据库的方法，如果该方法已存在，但是force为true，也就是说，
// 在强制情况下，即使对应的方法已存在，也会重新注册一个。
func registerDBCreator(backend BackendType, c creator, force bool) {
	_, ok := backends[backend]
	if ok && !force {
		return
	}
	backends[backend] = c
}

func NewDB(name string, backend BackendType, dir string) (DB, error) {
	create, ok := backends[backend]
	if !ok {
		keys := make([]string, 0)
		for k := range backends {
			keys = append(keys, string(k))
		}
		return nil, fmt.Errorf("unknown backend type %s, expected one of %s", backend, strings.Join(keys, ","))
	}
	return create(name, dir)
}
