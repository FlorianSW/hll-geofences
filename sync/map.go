package sync

import "sync"

type Map[K, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Load(k K) (v V, ok bool) {
	value, ok := m.m.Load(k)
	if !ok {
		return v, ok
	}
	return value.(V), ok
}

func (m *Map[K, V]) LoadOrStore(k K, v V) (V, bool) {
	actual, loaded := m.m.LoadOrStore(k, v)
	return actual.(V), loaded
}

func (m *Map[K, V]) Store(k K, v V) {
	m.m.Store(k, v)
}

func (m *Map[K, V]) Delete(k K) {
	m.m.Delete(k)
}

func (m *Map[K, V]) Range(f func(k K, v V) bool) {
	m.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}
