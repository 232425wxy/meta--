package flowrate

import (
	"math"
	"time"
)

// flowrate包存在的意义就是防止自己发送消息过快，对方来不及处理，也防止对方发送消息过快，自己来不及处理。

// clockPrecision ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// clockPrecision 定义了时钟的精度：20毫秒。
const clockPrecision = 20 * time.Millisecond

// baseTime ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// baseTime 可以看成是程序启动的基准时间。
var baseTime = time.Now().Round(clockPrecision)

// clock ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// clock 返回现在离基准时间 baseTime 过去了多久时间。
func clock() time.Duration {
	return time.Now().Sub(baseTime).Round(clockPrecision)
}

// clockToTime ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// clockToTime 给定一个时间段，将该时间段加到基准时间 baseTime 上，得到 time.Time 并返回。
func clockToTime(c time.Duration) time.Time {
	return baseTime.Add(c)
}

// clockRound ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// clockRound 给定一个时间段t，然后将该时间段的精度设置为 clockPrecision，公式如下：
//
//	new = (t+precision/2) / precision * precision
//
// 所以假如给了一个95ms这样的时间段，经过计算以后得到的值就是100ms，而像87ms这样的时间段，
// 经过计算后，得到的值就是80ms。
func clockRound(t time.Duration) time.Duration {
	return (t + clockPrecision>>1) / clockPrecision * clockPrecision
}

// round ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// round 方法接受一个小数作为入参，然后对该小数进行四舍五入得到一个整数。
func round(x float64) int64 {
	if _, frac := math.Modf(x); frac >= 0.5 {
		return int64(math.Ceil(x))
	}
	return int64(math.Floor(x))
}
