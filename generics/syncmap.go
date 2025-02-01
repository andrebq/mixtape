package generics

import (
	"iter"
	"sync"
)

type (
	SyncMap[K comparable, V any] struct {
		l     sync.RWMutex
		items map[K]V
	}
)

func (s *SyncMap[K, V]) Put(k K, v V) (V, bool) {
	return s.update(k, v, false)
}

func (s *SyncMap[K, V]) Update(k K, fn func(v V, present bool) (newval V, keep bool)) {
	s.l.Lock()
	defer s.l.Unlock()
	if s.items == nil {
		s.items = map[K]V{}
	}
	old, found := s.items[k]
	newval, keep := fn(old, found)
	if !keep {
		delete(s.items, k)
	} else {
		s.items[k] = newval
	}
}

func (s *SyncMap[K, V]) Get(k K) (V, bool) {
	s.l.RLock()
	if s.items == nil {
		s.l.RUnlock()
		var zero V
		return zero, false
	}
	v, found := s.items[k]
	s.l.RUnlock()
	return v, found
}

func (s *SyncMap[K, V]) Delete(k K) (V, bool) {
	var zero V
	return s.update(k, zero, true)
}

func (s *SyncMap[K, V]) LockedIter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		s.l.RLock()
		if s.items == nil {
			s.l.RUnlock()
			return
		}
		defer s.l.RUnlock()
		for k, v := range s.items {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (s *SyncMap[K, V]) update(k K, v V, del bool) (V, bool) {
	s.l.Lock()
	if s.items == nil && del {
		s.l.Unlock()
		return v, false
	} else if !del {
		s.items = map[K]V{}
	}
	oldv, found := s.items[k]
	s.items[k] = v
	s.l.Unlock()
	return oldv, found
}
