package main

import (
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/store"
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

	return service.Sync(
		ctx.Context,
		store.SyncParams{},
	)
}
