package scope

import (
	"context"
	"fmt"
)

type Typed[T any] struct{}

func (t Typed[T]) Set(ctx context.Context, v T) context.Context {
	return context.WithValue(
		ctx,
		t, // use t as key
		v,
	)
}

func (t Typed[T]) Get(ctx context.Context) (T, bool) {
	v, ok := ctx.Value(t).(T)
	return v, ok
}

func (t Typed[T]) MustGet(ctx context.Context) T {
	v, ok := ctx.Value(t).(T)
	if !ok {
		panic(fmt.Sprintf("missing value for %T", t))
	}

	return v
}
