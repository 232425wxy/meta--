package consensus

import (
	"fmt"
	"github.com/232425wxy/meta--/log"
	"os"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	ticker := NewTimeoutTicker()
	logger := log.New()
	logger.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	ticker.SetLogger(logger)
	ticker.Start()
	go func() {
		for {
			select {
			case tock := <-ticker.TockChan():
				fmt.Println("tock:", tock)
			}
		}
	}()

	go func() {
		for {
			time.Sleep(time.Millisecond * 1100)
			ti := timeoutInfo{
				Duration: time.Second,
				Height:   3,
				Round:    0,
				Step:     NewRoundStep,
			}
			ticker.ScheduleTimeout(ti)
		}
	}()

	select {}
}
