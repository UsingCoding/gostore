package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func copyCmd() *cli.Command {
	return &cli.Command{
		Name:      "copy",
		Aliases:   []string{"cp"},
		Usage:     "Copies path in store",
		UsageText: "cp <src> <dst>",
		Action:    executeCopy,
		BashComplete: func(ctx *cli.Context) {
			if ctx.NArg() > 1 {
				return
			}

			printTree(ctx)
		},
	}
}

func executeCopy(ctx *cli.Context) error {
	if ctx.Args().Len() < 2 {
		return errors.New("expected exactly 2 arguments")
	}

	src := ctx.Args().Get(0)
	dst := ctx.Args().Get(1)

	service, _ := newStoreService(ctx)

	return service.Copy(ctx.Context, store.CopyParams{
		CommonParams: makeCommonParams(ctx),
		Src:          src,
		Dst:          dst,
	})
}
