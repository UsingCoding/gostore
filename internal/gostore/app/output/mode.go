package output

import (
	"context"
	"fmt"
)

type Output int

const (
	Plain = Output(iota)
	JSON
)

func InitToCtx(ctx context.Context, output string) (context.Context, error) {
	o, err := convertOutput(output)
	if err != nil {
		return nil, err
	}
	return toCtx(ctx, o), nil
}

func convertOutput(output string) (Output, error) {
	switch output {
	case "", "plain":
		return Plain, nil
	case "json":
		return JSON, nil
	default:
		return 0, fmt.Errorf("invalid output: %s", output)
	}
}
