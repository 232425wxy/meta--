package btree

import (
	"fmt"
	"sort"
	"sync"
)

type Item interface {
	Less(other Item) bool
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 二叉树

type BTree struct {
	degree int
	length int
	root   *node
	cow    *copyOnWriteContext
}

func New(degree int) *BTree {
	if degree <= 1 {
		panic(fmt.Sprintf("invalid degree: %d", degree))
	}
	fl := NewFreeList(DefaultFreeListSize)
	return &BTree{degree: degree, cow: &copyOnWriteContext{freelist: fl}}
}

func (t *BTree) Clone() *BTree {
	cow1, cow2 := *t.cow, *t.cow
	c := *t
	t.cow = &cow1
	c.cow = &cow2
	return &c
}

func (t *BTree) ReplaceOrInsert(item Item) Item {
	if item == nil {
		panic("prohibit inserting nil item in BTree")
	}
	if t.root == nil {
		t.root = t.cow.newNode()
		t.root = t.cow.newNode()
		t.root.items = append(t.root.items, item)
		t.length++
		return nil
	} else {
		t.root = t.root.mutableFor(t.cow)
		if len(t.root.items) >= t.maxItems() {
			item2, second := t.root.split(t.maxItems() / 2)
			oldroot := t.root
			t.root = t.cow.newNode()
			t.root.items = append(t.root.items, item2)
			t.root.children = append(t.root.children, oldroot, second)
		}
	}
	out := t.root.insert(item, t.maxItems())
	if out == nil {
		t.length++
	}
	return out
}

// maxItems ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// maxItems 返回树中一个节点最大所能持有的item数量。
func (t *BTree) maxItems() int {
	return t.degree*2 - 1
}

// minItems ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// minItems 返回树中一个节点最少需要持有的item数量。
func (t *BTree) minItems() int {
	return t.degree - 1
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// items -> []Item

type items []Item

// insertAt ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// insertAt 将新给定的元素插入到index这个位置处。
func (s *items) insertAt(index int, item Item) {
	*s = append(*s, nil)
	if index < len(*s) && index >= 0 {
		copy((*s)[index+1:], (*s)[index:])
		(*s)[index] = item
		return
	} else {
		// 如果给定的索引位置大于已经扩增后的s的长度，那么就直接在s的最后添加给定的新元素
		(*s)[len(*s)-1] = item
	}
}

// removeAt ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// removeAt 删除指定位置的元素，并将指定位置之后的元素通通向前移。
func (s *items) removeAt(index int) Item {
	if index >= len(*s) {
		return nil
	}
	item := (*s)[index]
	copy((*s)[index:], (*s)[index+1:])
	(*s)[len(*s)-1] = nil
	*s = (*s)[:len(*s)-1]
	return item
}

// pop ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// pop 弹出最后一个元素，并且将最后一个元素从列表中删除。
func (s *items) pop() Item {
	index := len(*s) - 1
	p := (*s)[index]
	(*s)[index] = nil
	*s = (*s)[:index]
	return p
}

// truncate ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// truncate 在指定的位置截断列表，只保留前index个元素，后面的元素通通丢弃。
func (s *items) truncate(index int) {
	var clear items
	*s, clear = (*s)[:index], (*s)[index:]
	for len(clear) > 0 {
		// 将16个空的Item拷贝给clear的前16个位置，这样垃圾回收机制会自动收回这片空间，
		// 同时每次复制都会返回复制的元素个数，这样就可以借助这个来逐渐压缩clear的空间
		clear = clear[copy(clear, make([]Item, 16)):]
	}
}

// find ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// find 从列表中找到指定元素的位置，如果找不到，就返回-1。
func (s *items) find(item Item) (index int, found bool) {
	i := sort.Search(len(*s), func(i int) bool {
		// 从0开始到len(*s)，返回s里最先遇到的比给定的item大的元素的索引值
		return item.Less((*s)[i])
	})
	if i > 0 && !(*s)[i-1].Less(item) {
		// 按道理来说，item不小于s[i-1]，那么此时如果s[i-1]也不小于item，那么只能说明s[i-1]和item相等
		return i - 1, true
	}
	return i, false
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// node 二叉树里的节点

type node struct {
	items    items
	children children
	cow      *copyOnWriteContext
}

// mutableFor ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// mutableFor 返回一个n的复制品，如果给定的context等于n的context，则返回n自身。
func (n *node) mutableFor(cow *copyOnWriteContext) *node {
	if n.cow == cow {
		return n
	}
	// 从context里取出一个node，同时该node会继承该context
	nn := cow.newNode()
	if cap(nn.items) >= len(n.items) {
		// 如果是new的一个node，则不可能发生此情况
		nn.items = nn.items[:len(n.items)]
	} else {
		// 如果是new的一个node，则复制n的items空间
		nn.items = make(items, len(n.items), cap(n.items))
	}
	// 以上操作都是为了完美复制n的items
	copy(nn.items, n.items)
	if cap(nn.children) >= len(n.children) {
		nn.children = nn.children[:len(n.children)]
	} else {
		nn.children = make(children, len(n.children), cap(n.children))
	}
	copy(nn.children, n.children)
	return nn
}

// mutableChild ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// mutableChild 复制n的第i个孩子，并将n的第i个孩子重新设置为复制品，然后返回第i个孩子的复制品。
func (n *node) mutableChild(i int) *node {
	// 这个地方返回的就是n.children[i]，但是c是新实例化的一个变量，所以c的地址和n.children[i]的
	// 地址不一样
	c := n.children[i].mutableFor(n.cow)
	// 重新给n.children[i]赋值，那么n.children[i]的地址就被改变了
	n.children[i] = c
	return c
}

// split ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// split 方法返回n的第i个item，然后将n的items分割成两部分：[0:i+1]和[i+1:]，把n的children也分成两部分：
// [0:i+2]和[i+2:]，然后将n的items的[i+1:]部分放到新的node里，将n的children的[i+1:]部分也放到新的node
// 里，最后返回item和新node。
func (n *node) split(i int) (Item, *node) {
	item := n.items[i]
	// 从context的FreeList里取出一个新的node
	next := n.cow.newNode()
	next.items = append(next.items, n.items[i+1:]...)
	n.items.truncate(i)
	if len(n.children) > 0 {
		next.children = append(next.children, n.children[i+1:]...)
		n.children.truncate(i + 1)
	}
	return item, next
}

// maybeSplitChild ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// maybeSplitChild 对n的第i个孩子节点进行分割。
func (n *node) maybeSplitChild(i, maxItems int) bool {
	if len(n.children[i].items) < maxItems {
		// 还不值得分割
		return false
	}
	first := n.mutableChild(i)
	item, second := first.split(maxItems / 2)
	n.items.insertAt(i, item)
	n.children.insertAt(i+1, second)
	return true
}

// insert ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// insert 在node的items里插入一个指定的Item，如果该Item已经存在，则用给定的Item替换已经存在的那个，并将
// 已经存在的那个返回出来；如果该node不存在子节点，就在自己的items里插入指定的item
func (n *node) insert(item Item, maxItems int) Item {
	i, b := n.items.find(item)
	if b {
		found := n.items[i]
		n.items[i] = item
		return found
	}
	// 想要插入的Item在node的items里不存在，并且该node没有子节点，那么就在自己的items里插入该item
	if len(n.children) == 0 {
		n.items.insertAt(i, item)
		return nil
	}
	// 该node有子节点
	if n.maybeSplitChild(i, maxItems) {
		inTree := n.items[i]
		switch {
		case item.Less(inTree):
		case inTree.Less(item):
			i++
		default:
			out := n.items[i]
			n.items[i] = item
			return out
		}
	}
	return n.mutableChild(i).insert(item, maxItems)
}

// get ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// get
func (n *node) get(key Item) Item {
	i, found := n.items.find(key)
	if found {
		return n.items[i]
	} else if len(n.children) > 0 {
		return n.children[i].get(key)
	}
	return nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// children -> []*node

type children []*node

// insertAt ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// insertAt 在给定的位置插入给定的node，被插入地方后面的node往后顺移一位。
func (c *children) insertAt(index int, n *node) {
	*c = append(*c, nil)
	if index < len(*c) && index >= 0 {
		copy((*c)[index+1:], (*c)[index:])
		(*c)[index] = n
		return
	} else {
		// 如果给定的索引位置大于已经扩增后的c的长度，那么就直接在c的最后添加给定的新元素
		(*c)[len(*c)-1] = n
	}
}

// removeAt ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// removeAt 删除指定位置的node，被删除位置后面的node全都往前顺移一位。
func (c *children) removeAt(index int) *node {
	if index >= len(*c) {
		return nil
	}
	item := (*c)[index]
	copy((*c)[index:], (*c)[index+1:])
	(*c)[len(*c)-1] = nil
	*c = (*c)[:len(*c)-1]
	return item
}

func (c *children) pop() *node {
	index := len(*c) - 1
	p := (*c)[index]
	(*c)[index] = nil
	*c = (*c)[:index]
	return p
}

func (c *children) truncate(index int) {
	var clear children
	*c, clear = (*c)[:index], (*c)[index:]
	for len(clear) > 0 {
		// 将16个空的Item拷贝给clear的前16个位置，这样垃圾回收机制会自动收回这片空间，
		// 同时每次复制都会返回复制的元素个数，这样就可以借助这个来逐渐压缩clear的空间
		clear = clear[copy(clear, make([]*node, 16)):]
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// FreeList 代表二叉树里的一系列节点

type FreeList struct {
	mu       sync.RWMutex
	freelist []*node
}

// NewFreeList ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewFreeList 新建一个容量为size的node列表。
func NewFreeList(size int) *FreeList {
	return &FreeList{freelist: make([]*node, 0, size)}
}

// newNode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// newNode 这个方法从FreeList中拿取一个node返回，如果FreeList里没有node可拿了，那就new一个再返回。
func (fl *FreeList) newNode() *node {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	index := len(fl.freelist) - 1
	if index < 0 {
		// 表明此时fl.freelist里面没有节点
		return new(node)
	}
	// 获取最后一个节点
	n := fl.freelist[index]
	fl.freelist[index] = nil
	fl.freelist = fl.freelist[:index]
	return n
}

// freeNode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// freeNode 往FreeList里添加一个指定的node，但是如果FreeList的容量不够了，就无法添加，那么会返回false。
func (fl *FreeList) freeNode(n *node) bool {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	if len(fl.freelist) < cap(fl.freelist) {
		fl.freelist = append(fl.freelist, n)
		return true
	}
	return false
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// copyOnWriteContext

type copyOnWriteContext struct {
	freelist *FreeList
}

// newNode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// newNode new一个node，或者从FreeList里取一个node出来，取出来的node会继承copyOnWriteContext，
// 继承的context会作为在将来能否被回收的依据。
func (c *copyOnWriteContext) newNode() *node {
	n := c.freelist.newNode()
	n.cow = c
	return n
}

// freeNode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// freeNode 回收给定的node，如果这个node的context等于c，那么就可以回收，因为这代表是c生成的node，否则就不能
// 回收，另外，也有可能回收失败，在c的FreeList容量满的时候就会回收失败。
func (c *copyOnWriteContext) freeNode(n *node) freeType {
	if n.cow == c {
		n.items.truncate(0)
		n.children.truncate(0)
		n.cow = nil
		if c.freelist.freeNode(n) {
			return ftStored
		} else {
			return ftFreelistFull
		}
	} else {
		return ftNotOwned
	}
}

type ItemIterator func(i Item) bool

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 常量

type freeType int

const (
	ftFreelistFull freeType = iota
	ftStored
	ftNotOwned
)

const DefaultFreeListSize = 32
