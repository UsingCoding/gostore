package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"gostore/internal/common/maybe"
	"gostore/internal/gostore/app/store"
)

func remove() *cli.Command {
	return &cli.Command{
		Name:    "remove",
		Aliases: []string{"rm"},
		Usage:   "Removes secret from storage or field from secret",
		Action:  executeRemove,
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