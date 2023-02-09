package consensus

import (
	"github.com/232425wxy/meta--/common/service"
	"time"
)

const tickTockBufferSize = 10

type timeoutInfo struct {
	Duration time.Duration `json:"duration"`
	Height   int64         `json:"height"`
	Round    int16         `json:"round"`
	Step     Step          `json:"step"`
}

type TimeoutTicker struct {
	service.BaseService
	timer    *time.Timer
	tickChan chan timeoutInfo
	tockChan chan timeoutInfo
}

func NewTimeoutTicker() *TimeoutTicker {
	tt := &TimeoutTicker{
		BaseService: *service.NewBaseService(nil, "TimeoutTicker"),
		timer:       time.NewTimer(0),
		tickChan:    make(chan timeoutInfo, tickTockBufferSize),
		tockChan:    make(chan timeoutInfo, tickTockBufferSize),
	}
	tt.stopTimer()
	return tt
}

func (tt *TimeoutTicker) Start() error {
	go tt.workRoutine()
	return nil
}

func (tt *TimeoutTicker) Stop() error {
	tt.stopTimer()
	return tt.BaseService.Stop()
}

func (tt *TimeoutTicker) ScheduleTimeout(ti timeoutInfo) {
	tt.tickChan <- ti
}

func (tt *TimeoutTicker) TockChan() <-chan timeoutInfo {
	return tt.tockChan
}

func (tt *TimeoutTicker) stopTimer() {
	if !tt.timer.Stop() { // 已经关闭了或者已经超时触发了
		select {
		case <-tt.timer.C:
		default:
		}
	}
}

func (tt *TimeoutTicker) workRoutine() {
	var info timeoutInfo
	for {
		select {
		case newInfo := <-tt.tickChan:
			tt.Logger.Debug("received new tick", "old_tick", info, "new_tick", newInfo)
			if newInfo.Height < info.Height {
				// 忽略以前的消息
				continue
			} else if newInfo.Height == info.Height {
				// 当前区块的消息
				if newInfo.Round < info.Round {
					continue
				} else if newInfo.Round == info.Round {
					if info.Step > 0 && newInfo.Step <= info.Step {
						continue
					}
				}
			}
			tt.stopTimer()
			info = newInfo
			tt.timer.Reset(info.Duration)
			tt.Logger.Debug("scheduled timeout", "duration", info.Duration, "height", info.Height, "round", info.Round, "step", info.Step)
		case <-tt.timer.C:
			tt.Logger.Debug("timed out", "duration", info.Duration, "height", info.Height, "round", info.Round, "step", info.Step)
			go func(tock timeoutInfo) { tt.tockChan <- tock }(info)
		case <-tt.WaitStop():
			return
		}
	}
}
