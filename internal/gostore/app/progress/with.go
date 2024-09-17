package progress

import "context"

type ctxKey struct{}

func FromCtx(ctx context.Context) Progress {
	v, ok := ctx.Value(ctxKey{}).(Progress)
	if !ok {
		return nopProgress{}
	}

	return v
}

func ToCtx(ctx context.Context, v Progress) context.Context {
	return context.WithValue(ctx, ctxKey{}, v)
}
