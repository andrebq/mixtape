package generics

import "sync"

type (
	Cell[T any] struct {
		l sync.Locker
		v T
	}
)

func (c *Cell[T]) Put(v T) T {
	var old T
	c.l.Lock()
	old = c.v
	c.v = v
	c.l.Unlock()
	return old
}

func (c *Cell[T]) Get() T {
	c.l.Lock()
	v := c.v
	c.l.Unlock()
	return v
}
