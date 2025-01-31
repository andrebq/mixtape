package generics

import "sync"

type (
	Queue[T any] struct {
		l     sync.Mutex
		items []T
	}
)

func (q *Queue[T]) Offer(v T) {
	q.l.Lock()
	q.items = append(q.items, v)
	q.l.Unlock()
}

func (q *Queue[T]) Take() (T, bool) {
	var zero T
	q.l.Lock()
	if len(q.items) == 0 {
		q.l.Unlock()
		return zero, false
	}
	v := q.items[0]
	q.items[len(q.items)-1] = zero
	copy(q.items, q.items[1:])
	q.items = q.items[:len(q.items)-1]
	q.l.Unlock()
	return v, true
}
