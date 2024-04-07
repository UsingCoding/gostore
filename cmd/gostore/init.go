package main

import (
	"os"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func initCmd() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Usage:  "Initialize store",
		Action: executeInit,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Usage:    "Local store id",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "store-path",
				Usage: "Path to new store",
			},
			&cli.StringSliceFlag{
				Name:    "recipients",
				Usage:   "Pass public key to store",
				Aliases: []string{"r"},
			},
			&cli.StringFlag{
				Name:  "remote",
				Usage: "Remote address for store",
			},
		},
	}
}

func executeInit(ctx *cli.Context) error {
	storeID := ctx.String("id")
	storePath := optStringFromCtx(ctx, "store-path")
	remote := optStringFromCtx(ctx, "remote")
	recipients := ctx.StringSlice("recipients")

	service, _ := newStoreService(ctx)

	res, err := service.Init(ctx.Context, store.InitParams{
		CommonParams: store.CommonParams{
			StorePath: storePath,
			StoreID:   maybe.NewJust(storeID),
		},
		Recipients: slices.Map(recipients, func(r string) encryption.Recipient {
			return encryption.Recipient(r)
		}),
		StorageType: maybe.Maybe[storage.Type]{},
		Remote:      remote,
	})
	if err != nil {
		return err
	}

	output := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
	output.Printf("Created store: %s", res.StorePath)
	if maybe.Valid(res.GeneratedIdentity) {
		output.Printf("Generated keys:")
		output.Printf("Public key: %s", maybe.Just(res.GeneratedIdentity).Recipient)
		output.Printf("Private key: %s", maybe.Just(res.GeneratedIdentity).PrivateKey)
	}

	return nil
}
