package generics

func NonBlockSend[T any](out chan<- T, v T) bool {
	select {
	case out <- v:
		return true
	default:
		return false
	}
}
