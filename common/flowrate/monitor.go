package flowrate

import (
	"sync"
	"time"
)

type Monitor struct {
	mu      sync.Mutex
	active  bool
	start   time.Duration
	bytes   int64         // 总共已经传输的字节数
	samples int64         // 总共已经传输的次数
	rSample float64       // 最近一次传输时的传输速率：byte/second
	rPeak   float64       // 传输速率的峰值
	sLast   time.Duration // 最近一次传输结束的时间
	sBytes  int64         // 距离sLast时间到现在已经传输的字节数
	sRate   time.Duration // 每个传输sample发生的频率
	tBytes  int64         //当前传输总共需要传输的字节数
	tLast   time.Duration // 距离上一次传输数据到现在所经历的时间
}

// NewMonitor ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewMonitor 创建一个新的流量监控器，流量监控器在网络通信中可以避免因自己发送数据过快，导致对方处理不
// 过来的情况发生，也可以避免对方发送数据过快，自己处理不过来。
func NewMonitor(sampleRate time.Duration) *Monitor {
	if sampleRate = clockRound(sampleRate); sampleRate <= 0 {
		sampleRate = 5 * clockPrecision
	}
	now := clock()
	return &Monitor{
		active: true,
		start:  now,
		sLast:  now,
		sRate:  sampleRate,
		tLast:  now,
	}
}
