package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func remove() *cli.Command {
	return &cli.Command{
		Name:    "remove",
		Aliases: []string{"rm"},
		Usage:   "Removes secret from storage or field from secret",
		Action:  executeRemove,
		BashComplete: func(ctx *cli.Context) {
			if ctx.NArg() > 0 {
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

func executeRemove(ctx *cli.Context) error {
	storePath := optStringFromCtx(ctx, "store")

	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}
	path := ctx.Args().Get(0)

	var key maybe.Maybe[string]
	if ctx.Args().Len() > 1 {
		key = maybe.NewJust(ctx.Args().Get(1))
	}

	service, _ := newStoreService(ctx)

	return service.Remove(ctx.Context, store.RemoveParams{
		CommonParams: store.CommonParams{
			StorePath: storePath,
		},
		Path: path,
		Key:  key,
	})
}
