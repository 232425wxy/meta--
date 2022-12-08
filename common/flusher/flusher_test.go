package flusher

import (
	"fmt"
	"testing"
	"time"
)

func TestTimerAfterFunc(t *testing.T) {
	isSet := true
	timer := time.AfterFunc(time.Second, func() {
		fmt.Println("fire")
		isSet = false
	})

	for {
		if !isSet {
			timer.Reset(time.Second)
			isSet = true
		}
	}
}
