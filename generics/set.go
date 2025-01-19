package generics

type (
	Set[T comparable] struct {
		items map[T]struct{}
	}
)

func SetOf[T comparable](items ...T) *Set[T] {
	s := &Set[T]{}
	s.PutAll(items...)
	return s
}

func (s *Set[T]) PutAll(items ...T) {
	s.init()
	for _, v := range items {
		s.items[v] = struct{}{}
	}
}

func (s *Set[T]) AppendTo(out []T) []T {
	if s.items == nil {
		return out
	}
	for k := range s.items {
		out = append(out, k)
	}
	return out
}

func (s *Set[T]) init() {
	if s.items == nil {
		s.items = map[T]struct{}{}
	}
}
