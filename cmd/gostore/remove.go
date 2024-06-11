package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
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

			printTree(ctx)
		},
	}
}

func executeRemove(ctx *cli.Context) error {
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
		CommonParams: makeCommonParams(ctx),
		Path:         path,
		Key:          key,
	})
}
