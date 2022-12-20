package btree

import "fmt"

type node struct {
	items     items
	children  children
	ownership *ownership
}

// mutableFor ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// mutableFor 方法接受一个ownership类型的参数ship作为输入，这个所有权参数一般来说就是B-Tree的所有权，
// 判断当前节点的所有权与B-Tree的所有权是否一样，如果一样，则直接返回当前的节点，让B-Tree对其进行修改，如
// 果不一样，则表明B-Tree对该节点无权进行修改，那么就克隆一个和当前节点一样的节点，只不过克隆出来的节点的所
// 有权与B-Tree一样，然后返回克隆出来的节点，这样，B-Tree就可以对被克隆出来的节点进行修改。
func (n *node) mutableFor(ship *ownership) *node {
	if n.ownership == ship {
		// 如果当前节点的所有权和给定的所有权一致，则直接返回当前节点
		return n
	}
	// 由于当前节点的所有权与给定的所有权不一致，则从给定的所有权的环境中获取一个新的节点
	nn := ship.getNode()
	if cap(nn.items) >= len(n.items) {
		// 如果已经开辟了足够大的空间，那么就不必再次开辟空间了，可以节省时间
		nn.items = nn.items[:len(n.items)]
	} else {
		nn.items = make(items, len(n.items), cap(n.items))
	}
	copy(nn.items, n.items)
	// 上面的步骤可以为新node存放元素的items开辟出与n.items一样大的空间

	if cap(nn.children) > len(n.children) {
		nn.children = nn.children[:len(n.children)]
	} else {
		nn.children = make(children, len(n.children), cap(n.children))
	}
	copy(nn.children, n.children)
	return nn
}

// mutableChild ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// mutableChild 方法接收一个整数index作为参数，首先寻找当前节点的第index个子节点，然后判断该子节点
// 的所有权与其父节点的所有权是否相同，如果相同则将该子节点返回出来，否则就克隆该子节点，并将克隆节点的所
// 有权改成父节点的所有权，最后用克隆节点替换掉父节点的第index个子节点，并将克隆节点返回出来。
func (n *node) mutableChild(index int) *node {
	c := n.children[index].mutableFor(n.ownership)
	n.children[index] = c
	return c
}

// split ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// split 方法接收参数s=j，假设在分割前，当前节点的元素集合和子节点集合分别是是：
// {i0, i1, ...,ij-1, ij, ij+1, ..., in}和{c0, c1, ...,cj-1, cj, cj+1, ..., cm}，
// 分割后，，当前节点的所含有的元素集合和子节点集合分别为：
// {i0, i1, ..., ij-1}和{c0, c1, ..., cj-1, cj}，
// 得到的新节点（另一半节点）所拥有的元素集合和子节点集合分别为：
// {ij+1, ..., in}和{cj+1, ..., cm}
func (n *node) split(s int) (Item, *node) {
	it := n.items[s]
	side := n.ownership.getNode() // side节点的所有权与节点n的一样
	// 将节点n的前s个元素保留下来，然后将剩下的元素给新的节点side
	side.items = append(side.items, n.items[s+1:]...)
	n.items.truncate(s) // 下标为s的元素没有被保存下来，也没有给新节点side，所以我们会在返回值处将其返回出来
	if len(n.children) > 0 {
		// 如果节点n还有若干个子节点，则将前s+1个子节点保留下来，然后将[s+1:]部分给新的节点side
		side.children = append(side.children, n.children[s+1:]...)
		n.children.truncate(s + 1) // 由于下标为s的子节点没有给新节点side，所以我们需要将下标为s的子节点保留下来
	}
	return it, side
}

// maybeSplitChild ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// maybeSplitChild 方法接收两个参数：i和threshold，它们都是int类型，参数i用来定位第当前节点的第
// i个子节点，threshold用来判断第i个子节点所含有的元素是否小于threshold，如果小于，则没必要对该子节
// 点进行分割，并返回false；如果不小于，则有必要对该子节点进行分割，分割的边界就等于threshold/2，并返
// 回true。threshold可以被认为是一个子节点所能拥有的元素个数上限，一旦拥有的元素个数达到threshold，
// 就需要对该子节点进行分割。
func (n *node) maybeSplitChild(i, threshold int) bool {
	if len(n.children[i].items) < threshold {
		// 这个子节点里面所含的元素个数不足以去划分
		return false
	}
	// 获取当前节点的第i个子节点，因为要对该子节点进行修改，所以得先确保该子节点的所有权与当前节点一样
	child := n.mutableChild(i)
	item, side := child.split(threshold / 2)
	n.items.insertAt(i, item)      // TODO 为什么把第i个子节点的第threshold/2个元素插到当前节点的第i个元素的位置呢
	n.children.insertAt(i+1, side) // 这个地方很好理解，分割的是第i个子节点，分割出来的新节点自然就插在第i个子节点的右边
	return true
}

// insert ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// insert 方法在指定的节点中插入给定的元素it，这个方法可以拆解成以下三步进行：
//  1. 首先查找当前节点的元素集合中是否存在给定的元素，如果存在就用给定的元素去替换集合中原先的元素，并将原先的元素返回出来；
//  2. 如果当前节点的元素集合中不存在给定的元素，且当前节点不存在子节点，那么我们就将给定的元素插入到当前节点所拥有的元素集合
//     中适当的位置，用index表示适当的位置，插入后，当前节点第index-1个元素的大小小于it，而第index+1个元素的大小大于it；
//  3. 如果当前节点的元素集合中不存在给定的元素，且当前节点存在子节点，那么在寻找当前节点集合中是否存在给定元素it的过程中，返
//     回的索引值index有两种取值可能，分别是：
//     - 等于当前节点所拥有的元素数量，在这种情况下，要查找的key大于当前节点所拥有的所有元素。
//     - 小于当前节点所拥有的元素数量，在这种情况下，要查找的key的大小介于index-1和index之间。
//     其实无论index取值情况如何，我们都会直接将it插入到当前节点的第index个子节点中，这里简单解释一下为什么，因为返回的index，
//     具有一个特点，那就是当前节点的第index个元素是第index个子节点所拥有的元素集合的上界，当前节点的第index-1个元素是第index
//     个子节点所拥有的元素集合的下界，所以我们直接进入第index个子节点里，试图把it插到该子节点的元素集合中。
//     在插入之前，我们需要先判断第index个子节点所含有的元素个数是否达到上限，如果达到了，就需要将第index个子节点分割成两个节点，
//     分别占据第index和第index+1个子节点的位置，然后我们要判断，index的取值究竟属于哪种情况，如果等于当前节点所拥有的元素数量，
//     我们就需要将it插入到第index+1个子节点中，否则插入到第index个子结点中，为的是保持B-Tree树中元素的排列顺序性质不变。
func (n *node) insert(it Item, threshold int) Item {
	index, found := n.items.find(it)
	if found {
		foundItem := n.items[index]
		n.items[index] = it // 即便该节点的元素集合中已经有了给定的元素，也依然用新元素去替代旧元素
		return foundItem
	}

	if len(n.children) == 0 {
		// 在当前节点没有子节点的情况下，直接在当前节点的适当位置插入给定的元素，所谓适当的位置，满足：
		// n.items[index-1] < it且n.items[index] > it
		n.items.insertAt(index, it)
		return nil
	}
	// TODO 为什么当前节点一定会有至少index个子节点呢？
	// 答：这是由B-Tree树的性质决定的，B-Tree要求树中每个节点(除叶子节点)含有的元素个数等于其指向子节点的指针个数减一，
	// 也就是说，如果一个节点含有m个元素，那么该节点就会指向m+1个子节点。解决了上面的疑问后，现在我们来具体分析：
	// 目前该节点的元素集合中不存在给定的元素，那么返回的index的值等于第一个大于给定元素的元素集合中元素的下标位置，
	// 在B-Tree里，此时index所指向的子节点内部存储的元素大小介于n.items[index-1]和n.items[index]之间，而it
	// 小于n.items[index]，并且大于n.items[index-1]，那么正好就将it插入到index所指向的子节点的元素集合中，在此
	// 之前，我们需要先判断一下第index个子节点内的元素个数是否达到上限，达到了还要对其进行分割。
	if n.maybeSplitChild(index, threshold) {
		// leastBiggerOrLast代表的元素具有两种情况：
		// 1. 给定的元素it大于当前节点所拥有的任意元素，那么leastBiggerOrLast表示的就是当前节点所拥有的元素集合中的最后一个元素
		// 2. 给定的元素大小介于当前节点所拥有的元素集合中某两个元素item'和item''之间，那么leastBiggerOrLast就等于item''
		leastBiggerOrLast := n.items[index]
		switch {
		case leastBiggerOrLast.Less(it):
			// 如果给定的元素大于当前节点所拥有的任意一个元素的值，则将该元素插入到被分割的节点的第二部分里
			index++
		case it.Less(leastBiggerOrLast):
			// 如果给定的元素大小介于当前节点所拥有的元素集合中某两个元素item'和item''之间，则对index什么也不做，这样就可以
			// 将给定的元素插入到被分割的节点的前半部分里。
		default:
			// 如果待插入的元素等于当前节点所拥有的元素集合中的某个元素，则用待插入的元素去替换掉这个元素，
			// 按理说这个情况是不会发生的呀
			n.items[index] = it
			return leastBiggerOrLast
		}
	}
	return n.mutableChild(index).insert(it, threshold)
}

// get ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// get 利用二分查找的思想，从B-Tree树中寻找给定的key，它的思想可以归纳如下：
//  1. 首先利用二分查找的思想在当前节点所拥有的元素集合中寻找是否存在给定的key，如果存在则直接返回找到的元素；
//  2. 在第一步里没有找到给定的key，想象一下B-Tree树的结构，在没找到的情况下，返回的索引值index取值有两种可能：
//     - 等于当前节点所拥有的元素数量，在这种情况下，要查找的key大于当前节点所拥有的所有元素。
//     - 小于当前节点所拥有的元素数量，在这种情况下，要查找的key的大小介于index-1和index之间。
//  3. 无论index的取值如何，由B-Tree树的结构特点所决定的，我们只需要跑到当前节点的第index个子节点中去继续寻找key，
//     这里简单解释一下为什么，因为在B-Tree树中，第index个元素的大小是第index个子节点所拥有元素集合的上界，而第index-1
//     个元素的大小是第index个子节点所拥有元素集合的下界，我们现在可以确定的是，key大于第index-1个元素的大小，小于第index
//     个元素的大小，所以，很自然的，我们就跑到第index个子节点里去寻找key。
func (n *node) get(key Item) Item {
	index, found := n.items.find(key)
	if found {
		return n.items[index]
	} else if len(n.children) > 0 {
		// 父节点里面没找到给定的key，返回的index可能的取值有如下两种可能：
		// 1. 等于len(n.items)，那么在这种情况下，要寻找的key可能在当前节点最后一个子节点所拥有的元素集合中；
		// 2. 小于len(n.items)，在这种情况下，要寻找的key可能在index所指向的子节点所拥有的元素集合中。
		// 这是一种迭代过程
		return n.children[index].get(key)
	}
	return nil
}

// min ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// min 方法返回以给定节点为树根所代表的子树中所有元素里最小的那个元素。
func min(n *node) Item {
	if n == nil {
		return nil
	}
	for len(n.children) > 0 {
		n = n.children[0]
	}
	if len(n.items) == 0 {
		return nil
	}
	return n.items[0]
}

// max ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// max 方法返回以给定节点为树根所代表的子树中所有元素里最大的那个元素。
func max(n *node) Item {
	if n == nil {
		return nil
	}
	for len(n.children) > 0 {
		n = n.children[len(n.children)-1]
	}
	if len(n.items) == 0 {
		return nil
	}
	return n.items[len(n.items)-1]
}

// remove ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// remove 删除以指定的节点为根的子树中给定的元素，这其中牵扯到子节点的元素变换，具体可以分为以下几种情况：
//  1. 指定的节点不含有子节点，则直接删除指定节点中的特定元素，这里不论是阐述最小的那个元素、还是删除最大的那个元素，亦或是
//     中间的某个元素，都是一样的；
//  2. 指定的节点含有子节点，要删除的特定元素的下标位置是i，那么第i个子节点的元素集合上限需要做出调整，也就是用第i个子节点
//     中最大的元素去替换指定节点中的特定元素(下标位置为i的元素)，但是这可能面临子节点元素数量达不到阈值的问题，针对这个问题，
//     分以下三种情况去处理：
//     ① 第i-1个子节点含有的元素数量超过了阈值，那么就将指定节点的第i-1个元素插入到第i个子节点的元素集合的第一个位置，然
//     用第i-1个子节点的最大的那个元素去替换指定节点的第i-1个元素，最后再把第i-1个子节点的最右边的子节点设置为第i个子
//     节点的最左边的子节点。做好以上准备工作后，即可保证第i个子节点的元素个数满足阈值要求，然后再去删除特定的元素。
//     ② 第i+1个子节点含有的元素数量超过了阈值，那么就将指定节点的第i个元素插入到第i个子节点的元素集合的最后一个位置，然
//     用第i+1个子节点的最小的那个元素去替换指定节点的第i个元素，最后再把第i+1个子节点的最左边的子节点设置为第i个子
//     节点的最右边的子节点。做好以上准备工作后，即可保证第i个子节点的元素个数满足阈值要求，然后再去删除特定的元素。
//     ③ 第i-1和第i+1个子节点的元素个数都不大于阈值，那么此时如果i小于指定节点所拥有的元素个数，我们可以将第i和第i+1个子
//     节点进行合并，如果i大于或等于指定节点所拥有的元素个数，那么我们可已经第i-1和第i个元素进行合并。做好以上准备工作后，
//     即可保证第i个子节点的元素个数满足阈值要求，然后再去删除特定的元素。
func (n *node) remove(it Item, minItems int, kind toRemove) Item {
	var index int
	var found bool
	switch kind {
	case removeItem:
		index, found = n.items.find(it)
		if len(n.children) == 0 {
			// TODO 为什么要在没有子节点的情况下删除指定的元素？
			// 回答：因为在没有子节点的情况下删除指定的元素，一定不会破坏B-Tree树的大小排序性质，
			// 假如当前节点有子节点，并且我们要删除的是第i个元素，那么第i+1个子节点的下界就变了，
			// 而且这不符合元素个数与子节点个数之间的关系。
			if found {
				return n.items.removeAt(index)
			}
			return nil
		}
	case removeMin:
		// 删除当前节点所拥有的元素中最小的那个元素
		if len(n.children) == 0 {
			return n.items.removeAt(0)
		}
		index = 0
	case removeMax:
		// 删除当前节点所拥有的元素中最大的那个元素
		if len(n.children) == 0 {
			return n.items.pop()
		}
		index = len(n.items)
	default:
		panic(fmt.Sprintf("invalid remove type: %d", kind))
	}

	// 如果能在前面switch分支里就将指定的元素删除掉，那么说明当前节点是个叶子节点，不含有子节点。
	// 到这里未能把指定的元素删除掉，那么就说明当前节点不是叶子节点，它还含有子节点。
	// 如果在含有子节点的情况下删除本节点元素集合中的某个元素，那么势必会对对应的子节点做出干扰，因为
	// 本节点所拥有的元素集合中的任意一个元素都是其子节点的元素上下界。假如我要删除第i个元素，那么第i
	// 个子节点的元素上界就被改变了，改变的方法就是将第i个子节点中最大的那个元素提取出来，作为其上界，
	// 那么这就牵扯到一个问题，就是提取出一个元素，会导致子节点的元素个数减一，但是B-Tree树要求每个子
	// 节点所用的的元素个数不能少于ceil(m/2)，m是B-Tree树的阶，所谓的阶就是B-Tree树中所有节点中拥
	// 有最多子节点的个数。

	// 经过上面一段注释的讲解，这里似乎就能明白为什么要判断子节点拥有的元素个数是否小于阈值了。
	if len(n.children[index].items) < minItems {
		// TODO 这里是干什么的呢？
		return n.growChildAndRemove(index, it, minItems, kind)
	}

	child := n.mutableChild(index)
	if found {
		removed := n.items[index]
		// removed这个元素是child这个子节点为根的子树所拥有的元素集合的上界，现在要把这个removed元素删掉了，那么，
		// 我们就需要从子树里寻找最大的那个元素作为这个子树新的上界，同时，为了满足B-Tree树的结构特性，还要把这个元素
		// 从子树里删除掉。
		n.items[index] = child.remove(nil, minItems, removeMax)
		return removed
	}
	return child.remove(it, minItems, kind)
}

// growChildAndRemove ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// growChildAndRemove 正如在remove方法中解释的那样，如果要删除指定节点中的某个元素，并且该节点还含有子节点，那么
// 就必须要对对应的子节点进行调整，因为我们删除的元素是对应子节点所拥有的元素集合的上界或者下界，为了保证B-Tree树结构
// 特点不变，那么就必须要将子节点里的最大元素或者最小元素拿出来，补充到父节点的元素集合中，这里想象一下B-Tree树的结构
// 和特点，应该很容易能够想明白。那么这样的话，子节点里的元素个数就会少一个，这很可能会导致子节点中元素个数小于阈值，所以，
// 我们可以从该子节点的左邻右舍借一个元素过来，保证元素个数不小于阈值。
func (n *node) growChildAndRemove(index int, it Item, minItems int, kind toRemove) Item {
	if index > 0 && len(n.children[index-1].items) > minItems {
		// 如果第index-1个子节点拥有的元素个数大于阈值，则把index-1这个子节点中最大的那个元素给index这个子节点
		child := n.mutableChild(index)
		borrowedFrom := n.mutableChild(index - 1)
		borrowedItem := borrowedFrom.items.pop()
		// 因为borrowedItem小于n.items[index-1]，所以我们得按下面的方式来调换元素
		child.items.insertAt(0, n.items[index-1])
		n.items[index-1] = borrowedItem
		if len(borrowedFrom.children) > 0 {
			// 因为index-1这个子节点的最大的元素被搞走了，所以相应的，它最大的子节点也要被拿走，这样才能符合B-Tree树
			// 的结构特点。
			child.children.insertAt(0, borrowedFrom.children.pop())
		}
	} else if index < len(n.items) && len(n.children[index+1].items) > minItems {
		// 如果第index+1个子节点拥有的元素个数大于阈值，则把index+1这个子节点中最小的那个元素给index这个子节点
		child := n.mutableChild(index)
		borrowedFrom := n.mutableChild(index + 1)
		borrowedItem := borrowedFrom.items.removeAt(0)
		// 因为borrowedItem大于n.items[index]，所以我们得按照下面的方式来调换元素
		child.items = append(child.items, n.items[index])
		n.items[index] = borrowedItem
		if len(borrowedFrom.children) > 0 {
			child.children = append(child.children, borrowedFrom.children.removeAt(0))
		}
	} else {
		// 将倒数第二个子节点和倒数第一个子节点进行合并
		if index >= len(n.items) {
			index--
		}
		child := n.mutableChild(index) // 左边的那个子节点
		mergeItem := n.items.removeAt(index)
		mergeChild := n.children.removeAt(index + 1)
		child.items = append(child.items, mergeItem)
		child.items = append(child.items, mergeChild.items...)
		child.children = append(child.children, mergeChild.children...)
		n.ownership.recycleNode(mergeChild) // 回收右边的那个节点
	}
	return n.remove(it, minItems, kind)
}

// iterate ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// iterate 方法用来迭代树中的元素
func (n *node) iterate(dir direction, start, end Item, includeStart bool, hit bool, iter ItemIterator) (bool, bool) {
	var ok, found bool
	var index int
	switch dir {
	case ascend: // 升序 start < end
		if start != nil {
			index, _ = n.items.find(start)
		}
		for i := index; i < len(n.items); i++ {
			// n.items[index]大于或等于start
			if len(n.children) > 0 {
				if hit, ok = n.children[i].iterate(dir, start, end, includeStart, hit, iter); !ok {
					return hit, false
				}
			}
			if !includeStart && !hit && start != nil && !start.Less(n.items[i]) {
				hit = true
				continue
			}
			hit = true
			if end != nil && !n.items[i].Less(end) {
				return hit, false
			}
			if !iter(n.items[i]) {
				return hit, false
			}
		}
		if len(n.children) > 0 {
			if hit, ok = n.children[len(n.children)-1].iterate(dir, start, end, includeStart, hit, iter); !ok {
				return hit, false
			}
		}
	case descend: // 降序 start > end
		if start != nil {
			index, found = n.items.find(start)
			if !found {
				index = index - 1
			}
		} else {
			index = len(n.items) - 1
		}
		for i := index; i >= 0; i-- {
			if start != nil && !n.items[i].Less(start) {
				if !includeStart || hit || start.Less(n.items[i]) {
					continue
				}
			}
			if len(n.children) > 0 {
				if hit, ok = n.children[i+1].iterate(dir, start, end, includeStart, hit, iter); !ok {
					return hit, false
				}
			}
			if end != nil && !end.Less(n.items[i]) {
				return hit, false
			}
			hit = true
			if !iter(n.items[i]) {
				return hit, false
			}
		}
		if len(n.children) > 0 {
			if hit, ok = n.children[0].iterate(dir, start, end, includeStart, hit, iter); !ok {
				return hit, false
			}
		}
	}
	return hit, true
}

func (n *node) reset(ship *ownership) bool {
	for _, child := range n.children {
		if !child.reset(ship) {
			return false
		}
	}
	return ship.recycleNode(n) != ftFreelistFull
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 子节点

// children ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// children 存储了B-Tree树中一个节点的所有子节点。
type children []*node

// insertAt ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// insertAt 在子节点集合的指定位置index处插入给定的节点n，注意：给定的index必须不大于len(children)。
func (c *children) insertAt(index int, n *node) {
	*c = append(*c, nil)
	if index < len(*c) {
		copy((*c)[index+1:], (*c)[index:])
	}
	(*c)[index] = n
}

// removeAt ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// removeAt 删除指定位置index处的节点，同时将index后面所有node往前移动一位。
func (c *children) removeAt(index int) *node {
	n := (*c)[index]
	copy((*c)[index:], (*c)[index+1:])
	(*c)[len(*c)-1] = nil
	*c = (*c)[:len(*c)-1]
	return n
}

// pop ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// pop 方法弹出子节点集合中的最后一个节点，并释放最后一个节点在集合中占据的内存空间。
func (c *children) pop() *node {
	index := len(*c) - 1
	n := (*c)[index]
	(*c)[index] = nil
	*c = (*c)[:index]
	return n
}

// truncate ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// truncate 截取前index个元素：[0, index)。
func (c *children) truncate(index int) {
	var toClear children // 很优雅的设计，将截取的剩余部分空间存储在这里，然后再将它们释放掉。
	*c, toClear = (*c)[:index], (*c)[index:]
	for len(toClear) > 0 {
		toClear = toClear[copy(toClear, nilChildren):]
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义常量，删除的元素类型，迭代器的方向

type toRemove uint8

const (
	removeItem toRemove = iota // 删除指定的元素
	removeMin                  // 删除子树里最小的那个元素
	removeMax                  // 删除子树里最大的那个元素
)

type direction int8

const (
	descend = direction(-1)
	ascend  = direction(1)
)
