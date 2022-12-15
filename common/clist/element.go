package clist

import "sync"

type Element struct {
	mu                     sync.RWMutex
	prev, next             *Element
	prevWaitCh, nextWaitCh chan struct{}
	removed                bool
	Value                  interface{}
}

// NextWaitChan ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NextWaitChan 返回一个阻塞的通道，直到下一个元素不为空，该通道才变得不阻塞。
func (e *Element) NextWaitChan() <-chan struct{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.nextWaitCh
}

// Next ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Next 非阻塞式地返回下一个元素，因此如果当前元素是最后一个元素，那么返回结果将是空的。
func (e *Element) Next() *Element {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.next
}

// Prev ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Prev 非阻塞式地返回前一个元素，因此如果当前元素是第一个元素，那么返回的结果将是空的。
func (e *Element) Prev() *Element {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.prev
}

// DetachPrev ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// DetachPrev 方法将当前元素与前面的元素断开连接，该方法只有在当前元素被remove的情况下才能被调用。
func (e *Element) DetachPrev() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.removed {
		panic("current element is not removed, so, cannot call DetachPrev() method")
	}
	e.prev = nil
}

func (e *Element) SetNext(elem *Element) {
	e.mu.Lock()
	defer e.mu.Unlock()
	old := e.next
	e.next = elem
	if old != nil && elem == nil {
		// 新添加的下一个元素是个空元素，那么将来想要获取下一个元素，就得等待啦
		e.nextWaitCh = make(chan struct{})
	}
	if old == nil && elem != nil {
		close(e.nextWaitCh)
	}
}

func (e *Element) SetPrev(elem *Element) {
	e.mu.Lock()
	defer e.mu.Unlock()
	old := e.prev
	e.prev = elem
	if old != nil && elem == nil {
		e.prevWaitCh = make(chan struct{})
	}
	if old == nil && elem != nil {
		close(e.prevWaitCh)
	}
}

func (e *Element) SetRemoved() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.removed = true
	if e.prev == nil {
		// 此元素都被移除了，将来不会再有机会在它的前面添加新元素了
		close(e.prevWaitCh)
	}
	if e.next == nil {
		// 此元素都被移除了，将来不会再有机会在它的后面添加新元素了
		close(e.nextWaitCh)
	}
}
