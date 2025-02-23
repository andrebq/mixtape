package generics

func Transform[V any, SV ~[]V, T any, ST ~[]T](in SV, out ST, fn func(v V) T) ST {
	for _, v := range in {
		out = append(out, fn(v))
	}
	return out
}
