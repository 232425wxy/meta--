package pubsub

import (
	"sync"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 创建订阅

// NewSubscription ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewSubscription 根据给定的缓冲区大小数值创建一个新的订阅，新订阅中管理消息的缓冲区大小由给定的参数确定。
func NewSubscription(outCapacity int) *Subscription {
	return &Subscription{
		msgChan:   make(chan Message, outCapacity),
		cancelled: make(chan struct{}),
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义客户端订阅的结构体

// Subscription ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Subscription 一个订阅包含三个部分内容：
//  1. 首先就是订阅的消息，它由消息内容data和事件events组成，这个通过一个channel来管理消息；
//  2. 另一个就是取消订阅的开关，取消订阅也不会关闭管理消息的通道，以避免客户端收到一个nil消息；
//  3. 第三个就是err，用来记录为什么关闭该订阅。
type Subscription struct {
	msgChan   chan Message
	cancelled chan struct{}
	reason    error
	mu        sync.RWMutex
}

// MsgOut ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// MsgOut 返回订阅中管理消息的通道，区块链节点可以从该通道里获取消息，将来取消该订阅时，不应该关闭该
// 通道，以免接收方收到一条nil消息。
func (s *Subscription) MsgOut() <-chan Message {
	return s.msgChan
}

// CancelledWait ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// CancelledWait 返回控制取消该订阅的通道，该方法不是用来取消订阅的，而是在select结构中用于等待订阅
// 被取消的。
func (s *Subscription) CancelledWait() <-chan struct{} {
	return s.cancelled
}

// Reason ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Reason 该方法用来返回取消订阅的原因。
func (s *Subscription) Reason() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reason
}

// cancel ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// cancel 方法是用来真正取消订阅的，还要给定取消订阅的原因。
func (s *Subscription) cancel(err error) {
	s.mu.Lock()
	s.reason = err
	s.mu.Unlock()
	close(s.cancelled)
}

// 定义发布和订阅事件里的消息结构体

// Message ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Message 内部含有数据data和events，将这两个东西绑定在一起，构成一条订阅消息。
type Message struct {
	data   interface{}
	events map[string][]string
}

// NewMessage ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewMessage 根据给定的data和events新建一条消息。
func NewMessage(data interface{}, events map[string][]string) Message {
	return Message{data: data, events: events}
}

// Data ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Data 返回发布的消息内的具体数据data。
func (msg Message) Data() interface{} {
	return msg.data
}

// Events ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Events 返回发布的消息所包含的事件。
func (msg Message) Events() map[string][]string {
	return msg.events
}
