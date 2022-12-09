package cmap

import "sync"

type CMap struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewCap() *CMap {
	return &CMap{data: make(map[string]interface{})}
}

func (m *CMap) Set(key string, value interface{}) {
	m.mu.Lock()
	m.data[key] = value
	m.mu.Unlock()
}

func (m *CMap) Get(key string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key]
}

func (m *CMap) Has(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok
}

func (m *CMap) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *CMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}
