package slices

import (
	"github.com/UsingCoding/gostore/internal/common/maybe"
)

func Find[T any](s []T, f func(T) bool) maybe.Maybe[T] {
	for _, v := range s {
		if f(v) {
			return maybe.NewJust(v)
		}
	}

	return maybe.Maybe[T]{}
}
