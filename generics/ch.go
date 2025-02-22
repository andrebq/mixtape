package generics

func NonBlockSend[T any](out chan<- T, v T) bool {
	select {
	case out <- v:
		return true
	default:
		return false
	}
}

func NonBlockRecv[T any](in <-chan T) (val T, received bool, open bool) {
	select {
	case val, open = <-in:
		received = open
		return
	default:
		received = false
		open = true
		return
	}
}
