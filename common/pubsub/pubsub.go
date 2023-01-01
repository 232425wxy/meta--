package pubsub

import (
	"fmt"
	"github.com/232425wxy/meta--/common/pubsub/query"
	"github.com/232425wxy/meta--/common/service"
	"sync"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 项目级全局函数

// NewServer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewServer 实例化一个Server对象。
func NewServer(options ...Option) *Server {
	s := &Server{subscriptions: make(map[string]map[string]struct{})}
	s.BaseService = *service.NewBaseService(nil, "PubSub")
	for _, opt := range options {
		opt(s)
	}
	// options里可能有设置Server.cmdsCap的操作
	s.cmds = make(chan cmd, s.cmdsCap)
	return s
}

// OptionBufferCapacity ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// OptionBufferCapacity 设置 Server.cmdsCap 的一个选项。
func OptionBufferCapacity(cap int) Option {
	return func(server *Server) {
		if cap > 0 {
			server.cmdsCap = cap
		}
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义订阅和取消订阅的服务

type Server struct {
	service.BaseService
	cmds          chan cmd
	cmdsCap       int
	mu            sync.RWMutex
	subscriptions map[string]map[string]struct{} // one subscriber -> many queries (string) -> empty struct
}

// Option ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Option 在实例化Server的时候，可以方便我们对Server做出更多的调整。
type Option func(server *Server)

// BufferCapacity ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// BufferCapacity 返回命令缓冲区的大小。
func (s *Server) BufferCapacity() int {
	return s.cmdsCap
}

// Subscribe ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Subscribe 给定订阅者的ID以及要订阅的查询，将其注册到服务中，传入的第三个参数是可选的，如果大于0，
// 则会创建一个消息处理缓冲区大于0的Subscription，否则就创建一个缓冲区大小为0的Subscription。
func (s *Server) Subscribe(clientID string, query *query.Query, outCapacity ...int) (*Subscription, error) {
	capacity := 0
	if len(outCapacity) > 0 {
		if outCapacity[0] < 0 {
			// 这里就不panic了，直接修正错误
			outCapacity[0] = 0
		}
		capacity = outCapacity[0]
	}
	return s.subscribe(clientID, query, capacity)
}

// SubscribeUnbuffered ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// SubscribeUnbuffered 根据给定的客户端ID、查询请求创建一个消息处理缓冲区大小为0的Subscription。
func (s *Server) SubscribeUnbuffered(clientID string, query *query.Query) (*Subscription, error) {
	return s.subscribe(clientID, query, 0)
}

// Unsubscribe ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Unsubscribe 取消客户端指定的订阅。
func (s *Server) Unsubscribe(clientID string, query *query.Query) error {
	s.mu.RLock()
	clientSubscriptions, ok := s.subscriptions[clientID]
	if ok {
		_, ok = clientSubscriptions[query.String()]
	}
	s.mu.RUnlock()
	if !ok {
		return NewErrSubscriptionNotFound(clientID, query.String())
	}
	select {
	case s.cmds <- cmd{op: unsub, clientID: clientID, query: query}:
		s.mu.Lock()
		delete(clientSubscriptions, query.String())
		if len(clientSubscriptions) == 0 {
			delete(s.subscriptions, clientID)
		}
		s.mu.Unlock()
		return nil
	case <-s.WaitStop():
		return nil
	}
}

// UnsubscribeAll ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// UnsubscribeAll 取消指定客户端的所有订阅。
func (s *Server) UnsubscribeAll(clientID string) error {
	s.mu.RLock()
	_, ok := s.subscriptions[clientID]
	s.mu.RUnlock()
	if !ok {
		return NewErrSubscriptionNotFound(clientID, "any subscription")
	}
	select {
	case s.cmds <- cmd{op: unsub, clientID: clientID}:
		s.mu.Lock()
		delete(s.subscriptions, clientID)
		s.mu.Unlock()
		return nil
	case <-s.WaitStop():
		return nil
	}
}

// NumClients ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NumClients 返回订阅的客户端数量。
func (s *Server) NumClients() int {
	s.mu.RLock()
	num := len(s.subscriptions)
	s.mu.RUnlock()
	return num
}

// NumClientSubscriptions ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NumClientSubscriptions 返回指定客户端订阅的数量。
func (s *Server) NumClientSubscriptions(clientID string) int {
	s.mu.RLock()
	num := len(s.subscriptions[clientID])
	s.mu.RUnlock()
	return num
}

// Publish ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Publish 发布给定的消息，事实上，这个方法在程序运行期间不会起作用的，因为它的events是空的，匹配的时候
// 会一直返回false。可以看 query.Query 的 Matches 方法。
func (s *Server) Publish(msg interface{}) error {
	return s.PublishWithEvents(msg, make(map[string][]string))
}

// PublishWithEvents ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// PublishWithEvents 带着一系列事件发布给定的消息。
func (s *Server) PublishWithEvents(msg interface{}, events map[string][]string) error {
	select {
	case s.cmds <- cmd{op: pub, msg: msg, events: events}:
		return nil
	case <-s.WaitStop():
		return nil
	}
}

// Start ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Start 开启服务
func (s *Server) Start() error {
	state := &state{
		subscriptions: make(map[string]map[string]*Subscription),
		queries:       make(map[string]*queryRefCount),
	}
	go s.loop(state)
	return s.BaseService.Start()
}

// Stop ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Stop 关闭订阅服务。
func (s *Server) Stop() error {
	s.cmds <- cmd{op: shutdown}
	return s.BaseService.Stop()
}

// subscribe ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// subscribe 给定订阅者的ID以及要订阅的查询，将其注册到服务中。
func (s *Server) subscribe(clientID string, query *query.Query, outCapacity int) (*Subscription, error) {
	s.mu.RLock()
	clientSubscriptions, ok := s.subscriptions[clientID]
	// 判断给定的client是否已经订阅过
	if ok {
		_, ok = clientSubscriptions[query.String()]
	}
	s.mu.RUnlock()
	if ok {
		// 已经订阅过了，汇报一下错误
		return nil, NewErrAlreadySubscribed(clientID, query.String())
	}

	// 还未订阅，那就创建订阅吧
	subscription := NewSubscription(outCapacity)
	select {
	case s.cmds <- cmd{op: sub, clientID: clientID, query: query, subscription: subscription}:
		s.mu.Lock()
		if _, ok = s.subscriptions[clientID]; !ok {
			// 如果这个客户端还没在这里注册过，那就给他注册一下
			s.subscriptions[clientID] = make(map[string]struct{})
		}
		s.subscriptions[clientID][query.String()] = struct{}{}
		s.mu.Unlock()
		return subscription, nil
	case <-s.WaitStop():
		return nil, nil
	}
}

// loop ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// loop 开启一个协程，时刻监视命令通道里新到达的命令，并处理之。
func (s *Server) loop(state *state) {
loop:
	for cmd := range s.cmds {
		switch cmd.op {
		case unsub:
			if cmd.query != nil {
				state.remove(cmd.clientID, cmd.query.String(), NewErrUnsubscribed(cmd.clientID, cmd.query.String()))
			} else {
				state.removeClient(cmd.clientID, NewErrUnsubscribed(cmd.clientID, "all"))
			}
		case shutdown:
			state.removeAll(fmt.Errorf("pubsub: server is shutdown"))
			break loop
		case sub:
			state.add(cmd.clientID, cmd.query, cmd.subscription)
		case pub:
			if err := state.send(cmd.msg, cmd.events); err != nil {
				s.Logger.Error("error querying for events", "err", err)
			}
		}
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义辅助变量

type state struct {
	// query string -> many clients -> subscription
	subscriptions map[string]map[string]*Subscription
	queries       map[string]*queryRefCount
}

func (s *state) add(clientID string, q *query.Query, subscription *Subscription) {
	if _, ok := s.subscriptions[q.String()]; !ok {
		s.subscriptions[q.String()] = make(map[string]*Subscription)
	}
	// 新增一个订阅
	s.subscriptions[q.String()][clientID] = subscription
	if _, ok := s.queries[q.String()]; !ok {
		s.queries[q.String()] = &queryRefCount{q: q, refCount: 0}
	}
	// 又有一个客户端订阅了该查询
	s.queries[q.String()].refCount++
}

func (s *state) remove(clientID string, q string, reason error) {
	clientSubscriptions, ok := s.subscriptions[q]
	if !ok {
		// 压根没有客户端订阅该查询
		return
	}
	subscription, ok := clientSubscriptions[clientID]
	if !ok {
		// 订阅该查询的客户端没有订阅
		return
	}
	subscription.cancel(reason)
	delete(s.subscriptions[q], clientID)
	if len(s.subscriptions[q]) == 0 {
		delete(s.subscriptions, q)
	}
	s.queries[q].refCount--
	if s.queries[q].refCount == 0 {
		delete(s.queries, q)
	}
}

func (s *state) removeClient(clientID string, reason error) {
	for q, clientSubscriptions := range s.subscriptions {
		if _, ok := clientSubscriptions[clientID]; ok {
			s.remove(clientID, q, reason)
		}
	}
}

func (s *state) removeAll(reason error) {
	for q, clientSubscriptions := range s.subscriptions {
		for clientID := range clientSubscriptions {
			s.remove(clientID, q, reason)
		}
	}
}

func (s *state) send(msg interface{}, events map[string][]string) error {
	for q, clientSubscriptions := range s.subscriptions {
		query := s.queries[q].q
		match, err := query.Matches(events)
		if err != nil {
			return fmt.Errorf("pubsub: failed to match events %q against query %q: %q", events, q, err)
		}
		if match {
			for clientID, subscription := range clientSubscriptions {
				if cap(subscription.msgChan) == 0 {
					subscription.msgChan <- NewMessage(msg, events)
				} else {
					select {
					case subscription.msgChan <- NewMessage(msg, events):
					default:
						s.remove(clientID, q, NewErrOutOfCapacity(clientID))
					}
				}
			}
		}
	}
	return nil
}

type queryRefCount struct {
	q        *query.Query
	refCount int
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义操作订阅的命令cmd

// operation ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// operation 操作订阅的操作符：订阅、发布、取消订阅、关闭订阅。
type operation uint8

const (
	sub operation = iota
	pub
	unsub
	shutdown
)

type cmd struct {
	op           operation
	query        *query.Query
	subscription *Subscription
	clientID     string
	msg          interface{}
	events       map[string][]string
}
