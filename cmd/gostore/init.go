package main

import (
	"fmt"
	"os"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/urfave/cli/v2"

	"gostore/internal/common/maybe"
	"gostore/internal/gostore/app/encryption"
	"gostore/internal/gostore/app/storage"
	"gostore/internal/gostore/app/store"
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
			&cli.StringSliceFlag{
				Name:    "recipients",
				Usage:   "Pass public key to store",
				Aliases: []string{"r"},
			},
		},
	}
}

func executeInit(ctx *cli.Context) error {
	storeID := ctx.String("id")
	storePath := optStringFromCtx(ctx, "store")
	recipients := ctx.StringSlice("recipients")

	service, _ := newStoreService(ctx)

	res, err := service.Init(ctx.Context, store.InitParams{
		CommonParams: store.CommonParams{
			StorePath: storePath,
		},
		ID: storeID,
		Recipients: slices.Map(recipients, func(r string) encryption.Recipient {
			return encryption.Recipient(r)
		}),
		StorageType: maybe.Maybe[storage.Type]{},
		Remote:      maybe.Maybe[string]{},
	})
	if err != nil {
		return err
	}

	_, _ = os.Stdout.WriteString(fmt.Sprintf("Created store: %s\n", res.StorePath))
	if maybe.Valid(res.GeneratedIdentity) {
		_, _ = os.Stdout.WriteString("Generated keys:\n")
		_, _ = os.Stdout.WriteString(fmt.Sprintf("Public key: %s\n", maybe.Just(res.GeneratedIdentity).Recipient))
		_, _ = os.Stdout.WriteString(fmt.Sprintf("Private key: %s\n", maybe.Just(res.GeneratedIdentity).PrivateKey))
	}

	return nil
}
