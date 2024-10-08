package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func sync() *cli.Command {
	return &cli.Command{
		Name:   "sync",
		Usage:  "Sync store with remote",
		Action: executeSync,
	}
}

func executeSync(ctx *cli.Context) error {
	service, _ := newStoreService(ctx)

	err := service.Sync(
		ctx.Context,
		store.SyncParams{
			CommonParams: makeCommonParams(ctx),
		},
	)
	if err != nil {
		return err
	}

	o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
	o.OKf("Synced")

	return nil
}
