package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func clone() *cli.Command {
	return &cli.Command{
		Name:      "clone",
		Aliases:   nil,
		Usage:     "Clone store locally",
		UsageText: "clone <ADDRESS>",
		Action:    executeClone,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Usage:    "store id",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "store-path",
				Usage: "Clone store into store-path",
			},
			&cli.StringFlag{
				Name:  "storage-type",
				Usage: "Storage type to detect clone strategy",
				Value: string(storage.GITType),
			},
		},
	}
}

func executeClone(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}
	address := ctx.Args().Get(0)

	id := ctx.String("id")
	storageType := ctx.String("storage-type")

	service, _ := newStoreService(ctx)

	return service.Clone(ctx.Context, store.CloneParams{
		CommonParams: makeCommonParams(ctx),
		ID:           id,
		StorageType:  storage.Type(storageType),
		Remote:       address,
	})
}
