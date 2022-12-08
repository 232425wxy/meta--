package flusher

import (
	"sync"
	"time"
)

// Flusher ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Flusher 是一个定时刷新器。
type Flusher struct {
	Ch    chan struct{}
	quit  chan struct{}
	dur   time.Duration
	mu    sync.Mutex
	timer *time.Timer
	isSet bool
}

// NewFlusher ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewFlusher
func NewFlusher(dur time.Duration) *Flusher {
	var ch = make(chan struct{})
	var quit = make(chan struct{})
	var f = &Flusher{Ch: ch, dur: dur, quit: quit}
	f.timer = time.AfterFunc(dur, f.fireRoutine) // 过了dur时间后，会调用fireRoutine方法
	f.timer.Stop()
	return f
}

// fireRoutine ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// fireRoutine 被放在一个定时器里运行，每隔一段时间就往Ch通道里传入一个信号，这样接收方
// 拿到这个信号后就知道可以刷新信道了，fireRoutine一旦被执行一次，那么计时器就会停止工作，
// 除非重新设置计时时间。
func (f *Flusher) fireRoutine() {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case f.Ch <- struct{}{}:
		f.isSet = false // 此时计时器已经停止工作了，所以isSet被置为false。
	default:
		f.timer.Reset(f.dur)
	}
}

// Set ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Set 重新设置计时器，这样将来还能触发fireRoutine，往Ch通道里传递信号。
func (f *Flusher) Set() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.isSet {
		f.isSet = true
		f.timer.Reset(f.dur)
	}
}

func (f *Flusher) Unset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.isSet = false
	f.timer.Stop()
}

func (f *Flusher) Stop() bool {
	if f == nil {
		return false
	}
	close(f.quit)
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.timer.Stop()
}
