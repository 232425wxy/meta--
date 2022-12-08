package flowrate

import (
	"encoding/json"
	"sync"
	"time"
)

type Monitor struct {
	mu          sync.Mutex
	active      bool
	start       time.Duration
	bytes       int64         // 总共已经传输的字节数
	samples     int64         // 总共已经传输的次数
	rSample     float64       // 最近一次传输时的传输速率：byte/second
	rPeak       float64       // 传输速率的峰值
	sLast       time.Duration // 最近一次传输结束的时间
	sBytes      int64         // 距离sLast时间到现在已经传输的字节数
	sRate       time.Duration // 每个传输sample可以持续的时间
	tBytes      int64         // 当前传输总共需要传输的字节数
	tLast       time.Duration // 上一次传输至少1字节数据的时间点
	limitRate   int64         // 流量速率上限，https://github.com/mxk/go-flowrate/flowrate里并没有严格遵守流量上限这个规则
	limitSample int64         // 一次传输sample里最多能够传输的字节数：limitRate * sRate
}

// NewMonitor ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewMonitor 创建一个新的流量监控器，流量监控器在网络通信中可以避免因自己发送数据过快，导致对方处理不
// 过来的情况发生，也可以避免对方发送数据过快，自己处理不过来。
func NewMonitor(sampleRate time.Duration, limitRate int64) *Monitor {
	if sampleRate = clockRound(sampleRate); sampleRate <= 0 {
		sampleRate = 5 * clockPrecision
	}
	now := clock()
	return &Monitor{
		active:      true,
		start:       now,
		sLast:       now,
		sRate:       sampleRate,
		tLast:       now,
		limitRate:   limitRate,
		limitSample: round(float64(limitRate) * sampleRate.Seconds()),
	}
}

// Limit ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Limit 方法接受一个传输速率上限，如果本次sample传输的数据已经达到流量上限了，那么就等待
// 进入下一个sample中继续传输数据。
func (m *Monitor) Limit() {
	m.mu.Lock()
	now := m.update(0)
	for m.sBytes >= m.limitSample && m.active {
		// 本次传输sample里传输的字节数大于流量上限了，需要等待
		now = m.waitNextSample(now)
	}
	m.mu.Unlock()
}

// Update ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Update 记录本次传输sample传输的字节数，注意：当前sample可能并没有结束。
func (m *Monitor) Update(n int) {
	m.mu.Lock()
	m.update(n)
	m.mu.Unlock()
}

// update ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// update 每传输一次数据，就将当前传输的字节数更新到监视器中。
func (m *Monitor) update(n int) (now time.Duration) {
	if !m.active {
		return now
	}
	if now = clock(); n > 0 {
		m.tLast = now
	}
	m.sBytes += int64(n) // 当前传输的字节数累加到本次sample中传输的字节数
	if sTime := now - m.sLast; sTime >= m.sRate {
		// 如果本次sample的时间达到sRate，即一次sample的时间上限，那么就进入下一个sample里。
		t := sTime.Seconds()
		if m.rSample = float64(m.sBytes) / t; m.rSample > m.rPeak {
			m.rPeak = m.rSample
		}
		m.reset(now)
	}
	return now
}

// reset ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// reset 将当前sample里传输的状态更新到监视器总状态里，然后为下一次sample做准备。
func (m *Monitor) reset(sampleTime time.Duration) {
	if m.sBytes > m.limitSample {
		m.bytes += m.limitSample
		m.sBytes -= m.limitSample
	} else {
		m.bytes += m.sBytes // 总共已经传输的字节数加上刚刚sample里传输的字节数
		m.sBytes = 0
	}
	m.samples++          // 传输的sample次数加一
	m.sLast = sampleTime // 最近一次传输的时间
}

// waitNextSample ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// waitNextSample 等待进入下一个传输sample里，一般是因为传输的字节数达到了流量上限才会调用此方法。
func (m *Monitor) waitNextSample(now time.Duration) time.Duration {
	const minWait = 5 * time.Millisecond
	last := m.sLast // 上一次传输结束的时间
	for m.sLast == last && m.active {
		d := last + m.sRate - now
		m.mu.Unlock()
		if d < minWait {
			d = minWait
		}
		time.Sleep(d)
		m.mu.Lock()
		now = m.update(0)
	}
	return now
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义可以展示监视器状态的结构体

type Status struct {
	Start    time.Time     `json:"Start"`
	Bytes    int64         `json:"Bytes"`    // 总共传输的字节数
	Samples  int64         `json:"Samples"`  // 总共已经传输的次数
	CurRate  int64         `json:"CurRate"`  // 瞬时传输速率
	AvgRate  int64         `json:"AvgRate"`  // 平均传输速率
	PeakRate int64         `json:"PeakRate"` // 传输速率峰值
	Duration time.Duration `json:"Duration"` // 监视器启动至今的时间
	Active   bool          `json:"Active"`   // 表明监视器是否活跃
}

func (m *Monitor) Status() Status {
	m.mu.Lock()
	s := Status{
		Start:    clockToTime(m.start),
		Bytes:    m.bytes,
		Samples:  m.samples,
		CurRate:  round(m.rSample),
		AvgRate:  round(float64(m.bytes) / (m.sLast.Seconds() - m.start.Seconds())),
		PeakRate: round(m.rPeak),
		Duration: m.sLast - m.start,
		Active:   m.active,
	}
	m.mu.Unlock()
	return s
}

func (s Status) String() string {
	bz, _ := json.Marshal(s)
	return string(bz)
}
