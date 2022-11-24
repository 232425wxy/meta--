// handler.go
// 该文件内定义了打印日志的处理器，处理器决定了日志记录被打印到什么地方，以及什么级别的日志才会被打印。
// 这里主要就涉及到两种处理器：
// 	1. StreamHandler
//	2. LvlFilterHandler / FilterHandler

package log

import (
	"fmt"
	"github.com/go-stack/stack"
	"io"
	"reflect"
	"sync"
	"sync/atomic"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 接口

// StreamHandler ♏ | (o゜▽゜)o☆吴翔宇
//
// StreamHandler 接受两个参数：io.Writer 和 Format，其中第一个参数用来接受日志信息，第二个参数决定将以
// 什么样的格式把日志记录写入到 io.Writer 里。
func StreamHandler(wr io.Writer, fmtr formatter) Handler {
	h := FuncHandler(func(r *Record) error {
		_, err := wr.Write(fmtr.format(r))
		return err
	})
	// 这里h是真正记录日志的句柄，SyncHandler将h包装成一个多线程安全的Handler
	return lazyHandler(syncHandler(h))
}

// LvlFilterHandler ♏ | (o゜▽゜)o☆吴翔宇
//
// LvlFilterHandler 方法接受两个参数，分别是日志等级和 Handler，第一个参数设置了日志等级阈值，只有日志级别
// 小于第一个参数的日志才能被输出，众所周知，critical日志级别最高，trace日志级别最低。
func LvlFilterHandler(maxLvl Lvl, h Handler) Handler {
	return FilterHandler(func(r *Record) (pass bool) {
		return r.Lvl <= maxLvl
	}, h)
}

// FilterHandler ♏ | (o゜▽゜)o☆吴翔宇
//
// FilterHandler 接受两个参数作为入参，分别是函数fn func(r *Record) bool和 Handler，如果fn的返回值等于
// true，则调用 Handler 的Log方法将日志内容输出出去，否则什么也不干，忽略这条日志信息。例如，我们只输出日志中
// 存在"err"键，并且其对应的值不等于"nil"的日志：
//
//	logger.SetHandler(FilterHandler(func(r *Record) bool {
//	    for i := 0; i < len(r.Ctx); i += 2 {
//	        if r.Ctx[i] == "err" {
//	            return r.Ctx[i+1] != nil
//	        }
//	    }
//	    return false
//	}, h))
func FilterHandler(fn func(r *Record) bool, h Handler) Handler {
	return FuncHandler(func(r *Record) error {
		if fn(r) {
			return h.Log(r)
		}
		return nil
	})
}

// DiscardHandler ♏ | (o゜▽゜)o☆吴翔宇
//
// DiscardHandler 是一个假的处理器，如果处理器被设置成它，则不会打印出日志。
func DiscardHandler() Handler {
	return FuncHandler(func(r *Record) error {
		return nil
	})
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

// Handler ♏ | (o゜▽゜)o☆吴翔宇
//
// Handler 接收日志记录器产生的日志条目，然后 Handler 定义了怎样将日志条目输出出去。
type Handler interface {
	Log(r *Record) error
}

func FuncHandler(fn func(r *Record) error) Handler {
	return funcHandler(fn)
}

type funcHandler func(*Record) error

// Log ♏ | (o゜▽゜)o☆吴翔宇
//
// Log 实际上就是调用 funcHandler 函数。
func (fh funcHandler) Log(r *Record) error {
	return fh(r)
}

// swapHandler ♏ | (o゜▽゜)o☆吴翔宇
//
// swapHandler 可以在多线程情况下安全的切换 Handler。
type swapHandler struct {
	handler atomic.Value
}

func (h *swapHandler) Log(r *Record) error {
	return (*h.handler.Load().(*Handler)).Log(r)
}

func (h *swapHandler) Swap(newHandler Handler) {
	h.handler.Store(&newHandler)
}

func (h *swapHandler) Get() Handler {
	return *h.handler.Load().(*Handler)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// syncHandler ♏ | (o゜▽゜)o☆吴翔宇
//
// syncHandler 接收一个 Handler 作为输入参数，将给定的 Handler 包装成一个多线程安全的 Handler。
func syncHandler(h Handler) Handler {
	var mu sync.Mutex
	return FuncHandler(func(r *Record) error {
		mu.Lock()
		defer mu.Unlock()

		return h.Log(r)
	})
}

// lazyHandler ♏ | (o゜▽゜)o☆吴翔宇
//
// lazyHandler 方法接受一个 Handler 作为输入参数，LazyHandler 其实就是再将给定的 Handler 进行包装，
// LazyHandler 内部调用 FuncHandler 函数，并将其返回值返回，FuncHandler 接受的参数是一个函数的定义，
// 这个函数的定义是由 LazyHandler 函数设计的，将来我们调用 LazyHandler 函数返回的 Handler 的Log方法
// 时，实际上就是调用 LazyHandler -> FuncHandler -> 入参函数定义，在这个函数的定义内，会将日志记录 Record
// 里的 Ctx 过滤一遍，目的就是找到 Ctx 的value里面是否存在 Lazy 的实例，如果有的话，就执行这个 Lazy 实
// 例里的Fn函数，并将Fn函数的返回值替代Ctx中对应位置处的value。
func lazyHandler(h Handler) Handler {
	return FuncHandler(func(r *Record) error {
		hadErr := false
		for i := 1; i < len(r.Ctx); i += 2 {
			lz, ok := r.Ctx[i].(Lazy)
			if ok {
				v, err := evaluateLazy(lz)
				if err != nil {
					hadErr = true
					r.Ctx[i] = err
				} else {
					if cs, ok := v.(stack.CallStack); ok {
						// r.Call 是调用栈中的一个条目，调用栈的栈顶表示最开始调用的地方，越往下代表调用的越深，
						// TrimBelow方法就是将cs这个调用栈中处在r.Call条目之下的所有调用条目去除掉，例如cs是
						// [logger_test.go:31 testing.go:1446 asm_amd64.s:1594]，r.Call是 testing.go:1446，
						// 那么调用TrimBelow之后，cs就会变成 [logger_test.go:31 testing.go:1446]，TrimRuntime
						// 方法则是将cs调用栈中调用GOROOT源码的调用条目去掉，例如这里的testing.go:1446就是GOROOT
						// 里的代码，那么在执行完TrimRuntime之后，cs就只剩下[logger_test.go:31]了。
						// 实际上，在以太坊中，这段代码似乎永远都不会调用到。
						v = cs.TrimBelow(r.Call).TrimRuntime()
					}
					r.Ctx[i] = v
				}
			}
		}

		if hadErr {
			r.Ctx = append(r.Ctx, errorKey, "bad lazy")
		}

		return h.Log(r)
	})
}

// evaluateLazy ♏ | (o゜▽゜)o☆吴翔宇
//
// evaluateLazy 方法接收一个 Lazy 实例作为参数，Lazy 是一个结构体，内部只有一个 Fn 作为其唯一的
// 字段，Fn的类型是interface{}，所以理论上可以是任意数据类型，但是 evaluateLazy 方法要求Fn必须
// 是一个函数，而且该函数不能含有入参，而且必须具有返回值。在满足以上条件之后，evaluateLazy 方法会
// 运行Fn函数，并将其返回值返回出来。
func evaluateLazy(lz Lazy) (interface{}, error) {
	t := reflect.TypeOf(lz.Fn)

	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("INVALID_LAZY, not func: %+v", lz.Fn)
	}

	if t.NumIn() > 0 {
		return nil, fmt.Errorf("INVALID_LAZY, func takes args: %+v", lz.Fn)
	}

	if t.NumOut() == 0 {
		return nil, fmt.Errorf("INVALID_LAZY, no func return val: %+v", lz.Fn)
	}

	value := reflect.ValueOf(lz.Fn)
	// 因为lz.Fn是一个不接受任何输入参数的函数，因此调用时，传入的参数就为[]reflect.Value{}。
	results := value.Call([]reflect.Value{})
	if len(results) == 1 {
		return results[0].Interface(), nil
	}
	values := make([]interface{}, len(results))
	for i, v := range results {
		values[i] = v.Interface()
	}
	return values, nil
}
