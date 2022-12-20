package btree

import (
	"fmt"
	"sort"
)

// Item ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// Item 任何实现了Less方法的实例或结构体都可作为B-Tree上的元素。
type Item interface {
	// Less 方法用来比较两个元素的大小，如果!a.Less(b)且!b.Less(a)，则说明a==b，在
	// 这种情况下，我们只能从a和b中选一个放在B-Tree上。
	Less(other Item) bool
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 一个节点可以存储若干个元素，用items来表示

type items []Item

// insertAt ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// insertAt 在指定位置index插入给定的元素it，注意：给定的index必须小于或等于len(items)。
func (its *items) insertAt(index int, it Item) {
	*its = append(*its, nil)
	if index < len(*its) {
		copy((*its)[index+1:], (*its)[index:])
	}
	(*its)[index] = it
}

// removeAt ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// removeAt 删除指定位置index上的元素，并将index后的元素前移一个位置，注意：index必须小于len(items)。
func (its *items) removeAt(index int) Item {
	it := (*its)[index]
	copy((*its)[index:], (*its)[index+1:]) // its[len(its):] 不会导致数组越界
	(*its)[len(*its)-1] = nil
	*its = (*its)[:len(*its)-1]
	return it
}

// pop ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// pop 方法弹出items的最后一个元素，同时释放最后一个元素在items里所占据的空间。
func (its *items) pop() Item {
	index := len(*its) - 1
	it := (*its)[index]
	(*its)[index] = nil   // 释放内存空间
	*its = (*its)[:index] // 缩短切片长度
	return it
}

// truncate ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// truncate 截取items的前index个元素：[0, index)，注意：index必须不大于len(items)。
func (its *items) truncate(index int) {
	var toClear items // 这个地方太优雅了，要把截取的剩下部分释放掉，避免内存浪费
	*its, toClear = (*its)[:index], (*its)[index:]
	for len(toClear) > 0 {
		toClear = toClear[copy(toClear, nilItems):]
	}
}

// find ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// find 判断items里是否存在给定的it，如果存在，则返回其在items里的索引值。
func (its *items) find(it Item) (int, bool) {
	// 利用二分查找方法从items里找到第一个大于it的元素，并返回该元素的索引值
	index := sort.Search(len(*its), func(i int) bool {
		return it.Less((*its)[i])
	})
	// 假设items里第一个大于it的元素是item'，那么假设排在item'前面的元素是item''，由于item'>it，
	// 且item''不可能大于it，只可能小于或等于it，但是此时item''也不小于it，那么只有item''等于it
	// 这一情况了，所以，通过上面的逻辑可以判断出items里是否存在给定的it。
	if index > 0 && !(*its)[index-1].Less(it) {
		return index - 1, true
	}
	return index, false
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 迭代器

type ItemIterator func(it Item) bool

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// B-Tree

type BTree struct {
	degree int
	length int
	root   *node
	ship   *ownership
}

func New(degree int) *BTree {
	if degree <= 1 {
		panic(fmt.Sprintf("invalid degree: %d", degree))
	}
	return &BTree{degree: degree, ship: &ownership{freelist: NewFreeList(defaultFreeListSize)}}
}

func (bt *BTree) Clone() (bt2 *BTree) {
	ship1, ship2 := *bt.ship, *bt.ship // 值拷贝
	*bt2 = *bt
	*bt2.ship = ship2
	*bt.ship = ship1
	return bt2
}

func (bt *BTree) Insert(it Item) Item {
	if it == nil {
		panic("cannot insert nil Item into B-Tree")
	}
	if bt.root == nil { // 此时B-Tree还是一棵空树
		bt.root = bt.ship.getNode()
		bt.root.items = append(bt.root.items, it)
		bt.length++
		return nil
	} else {
		bt.root = bt.root.mutableFor(bt.ship)
		if len(bt.root.items) >= bt.maxItems() {
			_it, second := bt.root.split(bt.maxItems() / 2)
			oldRoot := bt.root
			bt.root = bt.ship.getNode()
			bt.root.items = append(bt.root.items, _it)
			bt.root.children = append(bt.root.children, oldRoot, second)
		}
	}
	out := bt.root.insert(it, bt.maxItems())
	if out == nil {
		bt.length++
	}
	return out
}

func (bt *BTree) Delete(it Item) Item {
	return bt.deleteItem(it, removeItem)
}

func (bt *BTree) DeleteMin() Item {
	return bt.deleteItem(nil, removeMin)
}

func (bt *BTree) DeleteMax() Item {
	return bt.deleteItem(nil, removeMax)
}

// Ascend ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// Ascend 方法利用给定的遍历方法按照升序的顺序从头到尾遍历树中的所有元素，遍历范围可以表示为：[first, last]。
func (bt *BTree) Ascend(iter ItemIterator) {
	if bt.root == nil {
		return
	}
	bt.root.iterate(ascend, nil, nil, false, false, iter)
}

// AscendRange ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// AscendRange 方法利用给定的迭代器遍历指定范围[start, end)内的所有元素。
func (bt *BTree) AscendRange(start, end Item, iter ItemIterator) {
	if bt.root == nil {
		return
	}
	bt.root.iterate(ascend, start, end, true, false, iter)
}

// AscendFromFirstToPivot ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// AscendFromFirstToPivot 方法利用给定的迭代器从树中第一个元素遍历到指定元素的前一个元素为止：[first, pivot)
func (bt *BTree) AscendFromFirstToPivot(pivot Item, iter ItemIterator) {
	if bt.root == nil {
		return
	}
	bt.root.iterate(ascend, nil, pivot, false, false, iter)
}

// AscendFromPivotToLast ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// AscendFromPivotToLast 方法利用给定的迭代器从指定的位置遍历到树中的最后一个元素：[pivot, last]
func (bt *BTree) AscendFromPivotToLast(pivot Item, iter ItemIterator) {
	if bt.root == nil {
		return
	}
	bt.root.iterate(ascend, pivot, nil, true, false, iter)
}

// Descend ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// Descend 方法利用给定的迭代器从树中最后一个元素开始往前遍历：last->first。
func (bt *BTree) Descend(iter ItemIterator) {
	if bt.root == nil {
		return
	}
	bt.root.iterate(descend, nil, nil, false, false, iter)
}

// DescendRange ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// DescendRange 方法利用给定的迭代器从给定的终点元素开始往前遍历到给定的七点元素的前一个元素，注意：end >= start，
// 遍历的范围是：(start, end]
func (bt *BTree) DescendRange(end, start Item, iter ItemIterator) {
	if bt.root == nil {
		return
	}
	bt.root.iterate(descend, end, start, true, false, iter)
}

// DescendFromPivotToFirst ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// DescendFromPivotToFirst 方法利用给定的迭代器从指定的终点元素开始往前遍历到树中的第一个元素，遍历范围是：[first, pivot]
func (bt *BTree) DescendFromPivotToFirst(pivot Item, iter ItemIterator) {
	if bt.root == nil {
		return
	}
	bt.root.iterate(descend, pivot, nil, true, false, iter)
}

// DescendFromLastToPivot ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// DescendFromLastToPivot 方法利用给定的迭代器从树中的最后一个元素往前遍历到指定的元素，遍历的范围是：(pivot, last]
func (bt *BTree) DescendFromLastToPivot(pivot Item, iter ItemIterator) {
	if bt.root == nil {
		return
	}
	bt.root.iterate(descend, nil, pivot, false, false, iter)
}

func (bt *BTree) Get(key Item) Item {
	if bt.root == nil {
		return nil
	}
	return bt.root.get(key)
}

func (bt *BTree) Min() Item {
	return min(bt.root)
}

func (bt *BTree) Max() Item {
	return max(bt.root)
}

func (bt *BTree) Has(key Item) bool {
	return bt.Get(key) != nil
}

func (bt *BTree) Length() int {
	return bt.length
}

func (bt *BTree) Clear(recycle bool) {
	if bt.root != nil && recycle {
		bt.root.reset(bt.ship)
	}
	bt.root, bt.length = nil, 0
}

// deleteItem ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// deleteItem
func (bt *BTree) deleteItem(it Item, kind toRemove) Item {
	if bt.root == nil || len(bt.root.items) == 0 {
		return nil
	}
	bt.root = bt.root.mutableFor(bt.ship)
	out := bt.root.remove(it, bt.minItems(), kind)
	if len(bt.root.items) == 0 && len(bt.root.children) > 0 {
		oldRoot := bt.root
		bt.root = bt.root.children[0]
		bt.ship.recycleNode(oldRoot)
	}
	if out != nil {
		bt.length--
	}
	return out
}

// maxItems ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// maxItems 规定了B-Tree树中一个节点所能拥有的元素个数最多是多少。
func (bt *BTree) maxItems() int {
	return bt.degree*2 - 1
}

// minItems ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// minItems 规定了B-Tree树中一个节点所能拥有的元素个数最少是多少。
func (bt *BTree) minItems() int {
	return bt.degree - 1
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义常量和包级变量

type freeType uint8

const defaultFreeListSize = 32

const (
	ftNotOwned freeType = iota
	ftFreelistFull
	ftStored
)

var (
	// nilItems ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
	//
	// nilItems 是一个长度为16的[]Item切片，里面的元素都是nil。
	nilItems = make(items, 16)

	nilChildren = make(children, 16)
)
