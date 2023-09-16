package main

import (
	"io"
	stdos "os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func add() *cli.Command {
	return &cli.Command{
		Name:   "add",
		Usage:  "Add secret to current store",
		Action: executeAdd,
	}
}

func executeAdd(ctx *cli.Context) error {
	storePath := optStringFromCtx(ctx, "store")

	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}
	path := ctx.Args().Get(0)

	data, err := io.ReadAll(stdos.Stdin)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return errors.New("empty stdin")
	}

	var key maybe.Maybe[string]
	if ctx.Args().Len() > 1 {
		key = maybe.NewJust(ctx.Args().Get(1))
	}

	service, _ := newStoreService(ctx)

	return service.Add(ctx.Context, store.AddParams{
		CommonParams: store.CommonParams{
			StorePath: storePath,
		},
		Path: path,
		Key:  key,
		Data: data,
	})
}
