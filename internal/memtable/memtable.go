package memtable

import (
	"sync"
)

type Memtable struct {
	list        *SkipList
	mu          sync.RWMutex
	currentSize int64 
	maxSize     int64 
}

func NewMemtable(maxSize int64) *Memtable {
	return &Memtable{
		list:    NewSkipList(),
		maxSize: maxSize,
	}
}

func (m *Memtable) Put(key, value []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.list.Put(key, value)
	m.currentSize += int64(len(key) + len(value) + 32)
}

func (m *Memtable) Get(key []byte) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.list.Get(key)
}

func (m *Memtable) IsFull() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentSize >= m.maxSize
}

func (m *Memtable) Entries() []struct{ Key, Value []byte } {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.list.Entries()
}