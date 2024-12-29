package slices

func Decompose[T any](s []T) (head T, tail []T) {
	if len(s) == 0 {
		panic("Decompose on empty slice")
	}

	head, tail = s[0], s[1:]
	return
}
