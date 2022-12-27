package database

import (
	"fmt"
	"strings"
)

type DB interface {
	// Get 根据给定的key获取对应的value
	Get([]byte) ([]byte, error)

	// Has 给定key，判断对应的key-value是否存在
	Has([]byte) (bool, error)

	// Set 往数据库里添加键值对，如果已存在指定的key，则替换其对应的value
	Set([]byte, []byte) error

	// SetSync 往数据库里添加键值对，如果已存在指定的key，则替换其对应的value，然后将内容持久化到磁盘
	SetSync([]byte, []byte) error

	// Delete 删除指定的key和其对应的value，如果指定的key不存在，则什么也不做
	Delete([]byte) error

	// DeleteSync 删除指定的key和其对应的value，然后将删除操作同步到磁盘，如果指定的key不存在，则什么也不做
	DeleteSync([]byte) error

	// Iterator 返回一个[start, end)区间的迭代器，如果start等于nil，则迭代器的起始位置从数据库的第一个key开始，
	// 如果end等于nil，则迭代器的结束位置是数据库的最后一个key
	Iterator(start, end []byte) (Iterator, error)

	// ReverseIterator 返回一个反转迭代器，区间位置是[start, end)，如果start等于nil，则迭代器的起始位置从数据库的第一个key开始，
	// 如果end等于nil，则迭代器的结束位置是数据库的最后一个key。
	ReverseIterator(start, end []byte) (Iterator, error)

	// Close 关闭数据库连接
	Close() error

	// NewBatch 创建一个batch
	NewBatch() Batch

	// Stats 返回数据库的状态信息
	Stats() map[string]string
}

type Batch interface {
	// Set 往数据库里添加键值对，还不会同步到磁盘上
	Set([]byte, []byte) error

	// Delete 删除指定的key和对应的value，还不会同步到磁盘上
	Delete([]byte) error

	// Write 调用此方法后不能再调用其他方法，只有Close方法能被调用
	Write() error

	// WriteSync 将batch里积攒的操作同步到磁盘上，此方法被调用后不能再调用其他方法，只有Close方法能被调用
	WriteSync() error

	// Close 关闭batch
	Close() error
}

type Iterator interface {
	// Domain 返回迭代器所能到达的区间[start, end)
	Domain() (start []byte, end []byte)

	// Valid 验证迭代器当前所处位置是否正确，如果迭代器一旦处于错误的状态则它会永远保持在错误的状态不变
	Valid() bool

	// Next 将迭代器移动到数据库里的下一个key所处的位置
	Next()

	// Key 返回迭代器当前所处位置的key
	Key() (key []byte)

	// Value 返回当前所处位置的value
	Value() (value []byte)

	// Error 如果迭代器在之前遇到错误，那么该方法返回迭代器最后一次遇到的错误
	Error() error

	// Close 关闭迭代器
	Close() error
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

func NewDB(name string, dir string, backend BackendType) (DB, error) {
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

func init() {
	registerDBCreator(MemDBBackend, func(name string, dir string) (DB, error) {
		return NewMemDB(), nil
	}, false)

	registerDBCreator(GoLevelDBBackend, func(name string, dir string) (DB, error) {
		return NewGoLevelDB(name, dir)
	}, false)
}
