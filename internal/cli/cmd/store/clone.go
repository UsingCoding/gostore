package store

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/storecrud"
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
	storePath := maybe.MapZero(ctx.String("store-path"))
	storageType := ctx.String("storage-type")

	service := clipkg.ContainerScope.MustGet(ctx.Context).StoreCRUD

	return service.Clone(
		ctx.Context,
		storecrud.CloneParams{
			StoreID:     id,
			StorePath:   storePath,
			StorageType: storage.Type(storageType),
			Remote:      address,
		},
	)
}
