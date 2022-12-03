package pubsub

import (
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
