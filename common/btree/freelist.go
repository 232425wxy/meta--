package btree

import "sync"

type FreeList struct {
	mu       sync.Mutex
	freelist []*node
}

// ownership ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// ownership 决定了一个节点的所有权，如果一棵B-Tree树的ownership和一个节点的ownership相同，
// 则这棵树就可以对该节点做出修改，否则就不可以修改这个节点
type ownership struct {
	freelist *FreeList
}

// getNode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// getNode 方法从ownership的环境中获取一个节点，并且该节点的所有权继承自ownership。
func (ship *ownership) getNode() *node {
	n := ship.freelist.getNode()
	n.ownership = ship
	return n
}

// recycleNode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// recycleNode 回收给定的节点，如果该节点的所有权与当前的ownership相等，则将该节点放回到ownership所维护的
// 自由列表里，这个自由列表可能已经满了，也可能没满，如果已经满了，则什么也不做，返回ftFreelistFull代码。另外，
// 给定的节点的所有权可能不等于当前的ownership，那么也将什么也不做，只管返回ftNotOwned代码。
func (ship *ownership) recycleNode(n *node) freeType {
	if n.ownership == ship {
		// 如果要回收的节点正是从当前ownership的环境中拿出来的，那么就放回该环境里
		n.items.truncate(0)
		n.children.truncate(0)
		n.ownership = nil
		if ship.freelist.recycleNode(n) {
			// 当前ownership的自由列表足够大，可以放下要回收的节点
			return ftStored
		} else {
			// 当前ownership的自由列表不够大，无法放下要回收的节点
			return ftFreelistFull
		}
	} else {
		// 要回收的节点不是由当前的ownership衍生出来的
		return ftNotOwned
	}
}

// NewFreeList ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// NewFreeList 创建一个自由列表，这个新创建的自由列表目前长度为0，但是容量为size。
func NewFreeList(size int) *FreeList {
	return &FreeList{freelist: make([]*node, 0, size)}
}

// getNode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// getNode 从自由列表里获取一个节点，如果此时自由列表里没有节点了，那么就new一个并返回，
// 如果有，那么就把自由列表里最后一个节点返回出来，同时还要把自由列表里返回出来的节点释放掉。
func (fl *FreeList) getNode() *node {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	index := len(fl.freelist) - 1
	if index < 0 { // 此时自由列表里还是空的
		return new(node)
	}
	n := fl.freelist[index]
	fl.freelist[index] = nil
	fl.freelist = fl.freelist[:index]
	return n
}

// recycleNode ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// recycleNode 回收给定的节点，如果此时自由列表的长度小于容量大小，则代表自由列表还能装节点，
// 那么就将给定的元素添加到自由列表里，并返回true，否则返回false，且什么也不做。
func (fl *FreeList) recycleNode(n *node) bool {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	if len(fl.freelist) < cap(fl.freelist) {
		fl.freelist = append(fl.freelist, n)
		return true
	}
	return false
}
