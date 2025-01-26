package generics

func ShiftHead[T any, E ~[]T](items E) (head T, found bool, tail E) {
	if len(items) == 0 {
		return
	}
	head = items[0]
	found = true
	tail = items[1:]
	return
}
