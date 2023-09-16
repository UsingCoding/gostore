package context

import (
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/config"
)

func remove() *cli.Command {
	return &cli.Command{
		Name:    "remove",
		Aliases: []string{"rm"},
		Usage:   "Removes store local copy",
		Action:  executeRemove,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Usage:    "Store id to remove",
				Required: true,
			},
		},
	}
}

func executeRemove(ctx *cli.Context) error {
	service := newConfigService(ctx)

	id := ctx.String("id")

	return service.RemoveStore(ctx.Context, config.StoreID(id))
}
