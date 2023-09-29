package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func move() *cli.Command {
	return &cli.Command{
		Name:      "move",
		Aliases:   []string{"mv"},
		Usage:     "Moves path in store",
		UsageText: "mv <src> <dst>",
		Action:    executeMove,
	}
}

func executeMove(ctx *cli.Context) error {
	storePath := optStringFromCtx(ctx, "store")

	if ctx.Args().Len() < 2 {
		return errors.New("expected exactly 2 arguments")
	}

	src := ctx.Args().Get(0)
	dst := ctx.Args().Get(1)

	service, _ := newStoreService(ctx)

	return service.Move(ctx.Context, store.MoveParams{
		CommonParams: store.CommonParams{
			StorePath: storePath,
		},
		Src: src,
		Dst: dst,
	})
}
