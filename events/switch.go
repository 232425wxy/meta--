package events

import (
	"fmt"
	"sync"
)

type EventSwitch struct {
	mu         sync.RWMutex
	eventCells map[string]*eventCell     // event -> eventCell
	listeners  map[string]*eventListener // listener name -> eventListener
}

type eventCell struct {
	mu sync.RWMutex
	// 一个event有多个listener关注，每个listener都有针对该event的处理办法：回调函数
	listeners map[string]EventCallback // listener name -> EventCallback
}

type eventListener struct {
	name    string
	mu      sync.RWMutex
	removed bool
	events  []string // 一个listener可以关注多个event
}

type EventCallback func(data EventData)

func NewEventSwitch() *EventSwitch {
	return &EventSwitch{
		eventCells: make(map[string]*eventCell),
		listeners:  make(map[string]*eventListener),
	}
}

func (sw *EventSwitch) AddListenerWithEvent(listenerName string, event string, cb EventCallback) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	cell := sw.eventCells[event]
	if cell == nil {
		cell = &eventCell{listeners: make(map[string]EventCallback)}
		sw.eventCells[event] = cell
	}
	cell.listeners[listenerName] = cb

	listener := sw.listeners[listenerName]
	if listener == nil {
		listener = &eventListener{name: listenerName, removed: false, events: nil}
		sw.listeners[listenerName] = listener
	}
	if err := listener.AddEvent(event); err != nil {
		return err
	}
	return nil
}

func (sw *EventSwitch) RemoveListener(listenerName string) {
	sw.mu.RLock()
	listener := sw.listeners[listenerName]
	sw.mu.RUnlock()
	if listener == nil {
		return
	}
	sw.mu.Lock()
	delete(sw.listeners, listenerName)
	sw.mu.Unlock()

	listener.removed = true
	for _, event := range listener.events {
		sw.mu.Lock()
		cell := sw.eventCells[event]
		sw.mu.Unlock()
		if cell == nil {
			return
		}
		if cell.RemoveListener(listenerName) == 0 {
			sw.mu.Lock()
			cell.mu.Lock()
			delete(sw.eventCells, event)
			cell.mu.Unlock()
			sw.mu.Unlock()
		}
	}
}

func (sw *EventSwitch) FireEvent(event string, data EventData) {
	sw.mu.RLock()
	cell := sw.eventCells[event]
	sw.mu.RUnlock()
	if cell == nil {
		return
	}
	cell.fire(data)
}

func (cell *eventCell) RemoveListener(listenerName string) int {
	cell.mu.Lock()
	defer cell.mu.Unlock()
	delete(cell.listeners, listenerName)
	return len(cell.listeners)
}

func (cell *eventCell) fire(data EventData) {
	cell.mu.RLock()
	callbacks := make([]EventCallback, 0, len(cell.listeners))
	for _, cb := range cell.listeners {
		callbacks = append(callbacks, cb)
	}
	cell.mu.RUnlock()
	for _, cb := range callbacks {
		cb(data)
	}
}

func (listener *eventListener) AddEvent(event string) error {
	listener.mu.Lock()
	if listener.removed {
		listener.mu.Unlock()
		return fmt.Errorf("listener %s is removed", listener.name)
	}
	listener.events = append(listener.events, event)
	listener.mu.Unlock()
	return nil
}
