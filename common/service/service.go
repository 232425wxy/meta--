package service

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/log"
	"sync/atomic"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 项目级全局变量

var (
	// ErrAlreadyStarted ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
	//  ---------------------------------------------------------
	// ErrAlreadyStarted 如果尝试启动已经启动的服务就会报告该错误。
	ErrAlreadyStarted = errors.New("already started")

	// ErrAlreadyStopped ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
	//  ---------------------------------------------------------
	// ErrAlreadyStopped 如果尝试关闭已经被关闭的服务就会报告该错误。
	ErrAlreadyStopped = errors.New("already stopped")

	// ErrNotRunning ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
	//  ---------------------------------------------------------
	// ErrNotRunning 如果尝试关闭还为开启的服务就会报告该错误。
	ErrNotRunning = errors.New("not running")
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义服务

// Service ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Service 简化了github.com/tendermint/tendermint里面定义的Service接口功能
type Service interface {
	Start() error
	Stop() error
	WaitStop() <-chan struct{}
	IsRunning() bool
}

// BaseService ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// BaseService 是一个基础服务对象。
type BaseService struct {
	Logger  log.Logger
	name    string
	started uint32
	stopped uint32
	quit    chan struct{}
}

// NewBaseService ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewBaseService 新建一个基础服务，如果给定的logger等于nil，那么会新建一个处理器是 log.DiscardHandler
// 的日志记录器，即啥也不会记录的日志记录器。
func NewBaseService(logger log.Logger, name string) *BaseService {
	if logger == nil {
		logger = log.New()
	}

	return &BaseService{
		Logger:  logger,
		name:    name,
		started: 0,
		stopped: 0,
		quit:    make(chan struct{}),
	}
}

// SetLogger ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// SetLogger 重新设置一个日志记录器。
func (bs *BaseService) SetLogger(logger log.Logger) {
	bs.Logger = logger
}

// Start ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Start 真正的服务在启动完以后调用该方法，因为我们将服务的启动逻辑放在了真正的服务那里，以简化对代码的理解。
func (bs *BaseService) Start() error {
	if atomic.CompareAndSwapUint32(&bs.started, 0, 1) {
		if atomic.LoadUint32(&bs.stopped) == 1 {
			// 服务已经停止了，不能再被启动了
			bs.Logger.Error(fmt.Sprintf("BaseService{%s}: cannot start, because already stopped", bs.name))
			atomic.StoreUint32(&bs.started, 0)
			return ErrAlreadyStopped
		}
		bs.Logger.Info(fmt.Sprintf("Service{%s}: start", bs.name))
		return nil
	}
	bs.Logger.Warn(fmt.Sprintf("BaseService{%s}: repeat start", bs.name))
	return ErrAlreadyStarted
}

// Stop ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Stop 关闭服务，该方法在真正的服务关闭以后再被调用。
func (bs *BaseService) Stop() error {
	if atomic.CompareAndSwapUint32(&bs.stopped, 0, 1) {
		if atomic.LoadUint32(&bs.started) == 0 {
			bs.Logger.Warn(fmt.Sprintf("BaseService{%s}: cannot stop, because service is not running", bs.name))
			return ErrNotRunning
		}
		bs.Logger.Info(fmt.Sprintf("BaseService{%s}: stop", bs.name))
		close(bs.quit)
		return nil
	}
	bs.Logger.Warn(fmt.Sprintf("BaseService{%s}: repeat stop", bs.name))
	return ErrAlreadyStopped
}

// WaitStop ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// WaitStop 等待服务的关闭，调用该方法会被阻塞，直到服务被关闭为止。
func (bs *BaseService) WaitStop() <-chan struct{} {
	return bs.quit
}

func (bs *BaseService) IsRunning() bool {
	return atomic.LoadUint32(&bs.started) == 1 && atomic.LoadUint32(&bs.stopped) == 0
}
