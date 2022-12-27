package database

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/common/btree"
	"sync"
)

type item struct {
	key   []byte
	value []byte
}

type operation struct {
	op
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
	m.btree.Insert(newPair(key, value))
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

// Iterator ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// Iterator start < end, 返回[start, end)区间内的元素。
func (m *MemDB) Iterator(start, end []byte) (Iterator, error) {
	iter := newMemDBIterator(m, start, end, false)
	return iter, nil
}

// ReverseIterator ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// ReverseIterator start < end，返回(start, end]区间内的元素
func (m *MemDB) ReverseIterator(start, end []byte) (Iterator, error) {
	iter := newMemDBIterator(m, start, end, true)
	return iter, nil
}

func (m *MemDB) Close() error {
	m.btree.Clear(false)
	return nil
}

func (m *MemDB) NewBatch() Batch {
	//TODO implement me
	panic("implement me")
}

func (m *MemDB) Stats() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	stats := make(map[string]string)
	stats["database.type"] = "memDB"
	stats["database.size"] = fmt.Sprintf("%d", m.btree.Length())
	return stats
}

var _ DB = (*MemDB)(nil)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 批量处理

type memBatch struct {
	db  *MemDB
	ops []operation
}

// Set ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// Set 往batch中插入一条存储键值对的指令。
func (m *memBatch) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if len(value) == 0 {
		return errValueEmpty
	}
	if m.ops == nil {
		return errBatchClosed
	}
	m.ops = append(m.ops, operation{key: key, value: value, op: opSet})
	return nil
}

// Delete ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// Delete 往batch里插入一条删除键值对的指令。
func (m *memBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if m.ops == nil {
		return errBatchClosed
	}
	m.ops = append(m.ops, operation{op: opDelete, key: key})
	return nil
}

// Write ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// Write 将batch里的所有指令都执行掉。
func (m *memBatch) Write() error {
	if m.ops == nil {
		return errBatchClosed
	}
	for _, op := range m.ops {
		switch op.op {
		case opSet:
			_ = m.db.Set(op.key, op.value)
		case opDelete:
			_ = m.db.Delete(op.key)
		default:
			return fmt.Errorf("unknown op type: %v", op.op)
		}
	}
	return nil
}

func (m *memBatch) WriteSync() error {
	return m.Write()
}

func (m *memBatch) Close() error {
	m.ops = nil
	return nil
}

var _ Batch = (*memBatch)(nil)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 内存数据库的迭代器

type memDBIterator struct {
	ch     chan *item
	cancel context.CancelFunc
	item   *item
	start  []byte
	end    []byte
}

func newMemDBIterator(db *MemDB, start, end []byte, reverse bool) *memDBIterator {
	ctx, cancel := context.WithCancel(context.Background())
	iter := &memDBIterator{
		ch:     make(chan *item, chBufferSize),
		cancel: cancel,
		start:  start,
		end:    end,
	}

	visitor := func(it btree.Item) bool {
		i := it.(*item)
		select {
		case iter.ch <- i:
			return true
		case <-ctx.Done():
			return false
		}
	}

	db.mu.RLock()
	go func() {
		// 放到协程里面，防止迭代器的通道不够大，一次性放不下所有数据，那么在将来就可以继续往里面放入数据
		defer db.mu.RUnlock()

		switch {
		case start == nil && end == nil && !reverse:
			// 从头迭代到尾
			db.btree.Ascend(visitor)
		case start == nil && end == nil && reverse:
			// 从尾迭代到头
			db.btree.Descend(visitor)
		case start != nil && end == nil && !reverse:
			// 从start迭代到尾
			db.btree.AscendFromPivotToLast(newKey(start), visitor)
		case start != nil && end == nil && reverse:
			// 从尾迭代到start
			db.btree.DescendFromLastToPivot(newKey(start), visitor)
		case start == nil && end != nil && !reverse:
			// 从头迭代到end
			db.btree.AscendFromFirstToPivot(newKey(end), visitor)
		case start == nil && end != nil && reverse:
			// 从end迭代到头
			db.btree.DescendFromPivotToFirst(newKey(end), visitor)
		case start != nil && end != nil && !reverse:
			// 从start迭代到end
			db.btree.AscendRange(newKey(start), newKey(end), visitor)
		case start != nil && end != nil && reverse:
			db.btree.DescendRange(newKey(end), newKey(start), visitor)
		}
		close(iter.ch) // 一旦关闭通道就不能往里面发送数据了，但是这里由于发送数据的操作在前面的switch分支里，所以在这里关闭通道不影响往通道里发送数据
	}()
	if it, ok := <-iter.ch; ok {
		iter.item = it
	}
	return iter
}

func (m *memDBIterator) Domain() (start []byte, end []byte) {
	return m.start, m.end
}

func (m *memDBIterator) Valid() bool {
	return m.item != nil
}

func (m *memDBIterator) Next() {
	m.assertIsValid()
	if it, ok := <-m.ch; ok {
		m.item = it
	} else {
		m.item = nil
	}
}

func (m *memDBIterator) Key() (key []byte) {
	m.assertIsValid()
	return m.item.key
}

func (m *memDBIterator) Value() (value []byte) {
	m.assertIsValid()
	return m.item.value
}

func (m *memDBIterator) Error() error {
	return nil
}

func (m *memDBIterator) Close() error {
	m.cancel()
	for range m.ch {

	}
	m.item = nil
	return nil
}

func (m *memDBIterator) assertIsValid() {
	if !m.Valid() {
		panic("MemDB iterator is invalid")
	}
}

var _ Iterator = (*memDBIterator)(nil)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级常量

const (
	bTreeDegree  = 32
	chBufferSize = 64
)

type op uint8

const (
	opSet op = iota
	opDelete
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义错误

var errKeyEmpty = errors.New("key cannot be empty")

var errValueEmpty = errors.New("value cannot be empty")

var errBatchClosed = errors.New("DB Batch is closed")
