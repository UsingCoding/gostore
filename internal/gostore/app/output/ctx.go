package output

import "context"

func FromCtx(ctx context.Context) Output {
	v, ok := ctx.Value(ctxKey{}).(Output)
	if !ok {
		return Plain
	}

	return v
}

func toCtx(ctx context.Context, o Output) context.Context {
	return context.WithValue(ctx, ctxKey{}, o)
}

type ctxKey struct{}
