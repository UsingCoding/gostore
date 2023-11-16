package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
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

			service, _ := newStoreService(ctx)

			entries, err := service.List(ctx.Context, store.ListParams{})
			if err != nil {
				return
			}

			o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
			for _, p := range inlinePaths(entries) {
				o.Printf(p)
			}
		},
	}
}

func executeCopy(ctx *cli.Context) error {
	storePath := optStringFromCtx(ctx, "store")

	if ctx.Args().Len() < 2 {
		return errors.New("expected exactly 2 arguments")
	}

	src := ctx.Args().Get(0)
	dst := ctx.Args().Get(1)

	service, _ := newStoreService(ctx)

	return service.Copy(ctx.Context, store.CopyParams{
		CommonParams: store.CommonParams{
			StorePath: storePath,
		},
		Src: src,
		Dst: dst,
	})
}
