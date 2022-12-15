package clist

import (
	"fmt"
	"sync"
)

type List struct {
	mu         sync.RWMutex
	wg         *sync.WaitGroup
	waitCh     chan struct{}
	head, tail *Element
	size       int
	maxSize    int
}

func NewList() *List {
	l := new(List)
	l.wg = waitGroup1()
	l.waitCh = make(chan struct{})
	l.maxSize = maxLength
	return l
}

func (l *List) Size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.size
}

func (l *List) Head() *Element {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.head
}

func (l *List) Tail() *Element {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.tail
}

// WaitChan ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// WaitChan 返回一个被阻塞的通道，如果链表的head或者tail变得不再是空的，那么此通道将不再阻塞，
// 换句话说，只要链表里有数据，那么返回的通道都不是阻塞的。
func (l *List) WaitChan() <-chan struct{} {
	l.mu.RLock()
	l.mu.RUnlock()
	return l.waitCh
}

func (l *List) Push(val interface{}) *Element {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := &Element{
		prevWaitCh: make(chan struct{}),
		nextWaitCh: make(chan struct{}),
		Value:      val,
	}
	if l.size == 0 {
		// 添加的头一个元素
		l.wg.Done()
		close(l.waitCh)
	}
	if l.size > l.maxSize {
		panic(fmt.Sprintf("the number of elements in the list exceeds the maximum limit: %d >= %d", l.size, l.maxSize))
	}
	l.size++
	if l.tail == nil {
		l.head = e
		l.tail = e
	} else {
		e.SetPrev(l.tail)
		l.tail.SetNext(e)
		l.tail = e
	}
	return e
}

func (l *List) Remove(e *Element) interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	prev := e.Prev()
	next := e.Next()
	if l.head == nil || l.tail == nil {
		panic("Remove(e) on empty CList")
	}
	if prev == nil && l.head != e {
		panic("Remove(e) with false head")
	}
	if next == nil && l.tail != e {
		panic("Remove(e) with false tail")
	}
	if l.size == 1 {
		// 目前链表里只剩一个元素了，删除这个元素后，链表里就没元素了
		l.wg = waitGroup1()
		l.waitCh = make(chan struct{})
	}
	l.size--
	if prev == nil {
		l.head = next
	} else {
		prev.SetNext(next)
	}
	if next == nil {
		l.tail = prev
	} else {
		next.SetPrev(prev)
	}
	e.SetRemoved()
	return e.Value
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

func waitGroup1() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return wg
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级常量，定义链表的最大长度

const maxLength = int(^uint(0) >> 1)
