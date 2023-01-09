package events

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const listener = "client"

const event = "NewBlock"

func TestNormal(t *testing.T) {
	evsw := NewEventSwitch()
	err := evsw.AddListenerWithEvent(listener, event, func(data EventData) {
		fmt.Println("get", data)
	})
	assert.Nil(t, err)
	evsw.FireEvent(event, "aaaaaa")
}

func TestDefer(t *testing.T) {
	defer func() {
		fmt.Println(1)
		fmt.Println(2)
		fmt.Println(3)
	}()
}
