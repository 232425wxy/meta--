package flowrate

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestMonitor_Limit(t *testing.T) {
	rate := 5120000 // 5MB/s
	monitor := NewMonitor(time.Second, int64(rate))
	total := math.MaxInt64
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-ticker.C:
				fmt.Println(monitor.Status())
			}
		}
	}()
	for total > 0 {
		transfer := rand.Intn(rate)
		monitor.Update(transfer)
		monitor.Limit()
		total -= transfer
	}
}

func TestUseful(t *testing.T) {
	rate := 5120000 // 5MB/s
	monitor := NewMonitor(time.Second, int64(rate))

	monitor.Update(rate * 2)
	monitor.Limit()
}
