package store

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/completion"
	"github.com/UsingCoding/gostore/internal/gostore/app/config"
)

func remove() *cli.Command {
	return &cli.Command{
		Name:         "remove",
		Aliases:      []string{"rm"},
		Usage:        "Removes store local copy",
		UsageText:    "rm <STORE_ID>",
		BashComplete: completion.ListStoresCompletion,
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() < 1 {
				return errors.New("not enough arguments")
			}

			storeID := ctx.Args().Get(0)

			service := clipkg.ContainerScope.MustGet(ctx.Context).C

			return service.RemoveStore(ctx.Context, config.StoreID(storeID))
		},
	}
}
