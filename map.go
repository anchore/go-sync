package sync

import (
	"maps"
	"slices"
	"sync"
)

type Map[Key comparable, Value any] struct {
	lock   sync.RWMutex
	values map[Key]Value
}

func NewMap[Key comparable, Value any]() *Map[Key, Value] {
	return &Map[Key, Value]{values: map[Key]Value{}}
}

func (m *Map[Key, Value]) Set(k Key, v Value) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.values[k] = v
}

func (m *Map[Key, Value]) Get(k Key) (v Value, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.values[k], true
}

func (m *Map[Key, Value]) Delete(k Key) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.values, k)
}

func (m *Map[Key, Value]) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return len(m.values)
}

func (m *Map[Key, Value]) Keys() []Key {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return slices.Collect(maps.Keys(m.values))
}

func (m *Map[Key, Value]) Values() []Value {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return slices.Collect(maps.Values(m.values))
}
