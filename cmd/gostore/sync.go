package main

import (
	"os"

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

	err := service.Sync(
		ctx.Context,
		store.SyncParams{},
	)
	if err != nil {
		return err
	}

	_, _ = os.Stdout.WriteString("Synced\n")
	return nil
}
