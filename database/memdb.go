package database

import (
	"bytes"
	"errors"
	"github.com/232425wxy/meta--/common/btree"
	"sync"
)

type item struct {
	key   []byte
	value []byte
}

// Less ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Less 实现B-Tree里元素的Less方法，通过比较元素的键key来比较不同元素之间的大小。
func (i *item) Less(other btree.Item) bool {
	return bytes.Compare(i.key, other.(*item).key) == -1
}

// newKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// newKey 返回的item仅包含给定的参数key，不包含value。
func newKey(key []byte) *item {
	return &item{key: key}
}

// newPair ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// newPair 根据给定的key和value生成一对键值对。
func newPair(key, value []byte) *item {
	return &item{key: key, value: value}
}

type MemDB struct {
	mu    sync.RWMutex
	btree *btree.BTree
}

func NewMemDB() *MemDB {
	return &MemDB{btree: btree.New(bTreeDegree)}
}

func (m *MemDB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errKeyEmpty
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := m.btree.Get(newKey(key))
	if res != nil {
		return res.(*item).value, nil
	}
	return nil, nil
}

func (m *MemDB) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, errKeyEmpty
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.btree.Has(newKey(key)), nil
}

func (m *MemDB) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if len(value) == 0 {
		return errValueEmpty
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.btree.ReplaceOrInsert(newPair(key, value))
	return nil
}

func (m *MemDB) SetSync(key []byte, value []byte) error {
	return m.Set(key, value)
}

func (m *MemDB) Delete(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.btree.Delete(newKey(key))
	return nil
}

func (m *MemDB) DeleteSync(key []byte) error {
	return m.Delete(key)
}

func (m *MemDB) Iterator(start, end []byte) (Iterator, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemDB) ReverseIterator(start, end []byte) (Iterator, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemDB) Close() error {
	//TODO implement me
	panic("implement me")
}

func (m *MemDB) NewBatch() Batch {
	//TODO implement me
	panic("implement me")
}

func (m *MemDB) Stats() map[string]string {
	//TODO implement me
	panic("implement me")
}

var _ DB = (*MemDB)(nil)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 内存数据库的迭代器

type memDBIterator struct {
	ch <-chan *item
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级常量

const (
	bTreeDegree = 32
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义错误

var errKeyEmpty = errors.New("key cannot be empty")

var errValueEmpty = errors.New("value cannot be empty")
