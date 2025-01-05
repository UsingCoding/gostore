package store

import (
	"os"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storecrud"
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

	storePath := maybe.MapZero(ctx.String("store-path"))
	remote := maybe.MapZero(ctx.String("remote"))
	recipients := slices.Map(ctx.StringSlice("recipients"), func(r string) encryption.Recipient {
		return encryption.Recipient(r)
	})

	service := clipkg.ContainerScope.MustGet(ctx.Context).StoreCRUD

	res, err := service.Init(ctx.Context, storecrud.InitParams{
		StoreID:    storeID,
		StorePath:  storePath,
		Recipients: recipients,
		Remote:     remote,
	})
	if err != nil {
		return err
	}

	output := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
	output.Printf("Created store: %s", res.StorePath)
	if i, ok := maybe.JustValid(res.Identity); ok {
		output.Printf("Generated keys:")
		output.Printf("Public key: %s", i.Recipient)
		output.Printf("Private key: %s", i.PrivateKey)
	}

	return nil
}
