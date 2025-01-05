package store

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/completion"
)

func use() *cli.Command {
	return &cli.Command{
		Name:         "use",
		Usage:        "Switch current store",
		UsageText:    "use <STOREPATH>",
		BashComplete: completion.ListStoresCompletion,
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() < 1 {
				return errors.New("not enough arguments")
			}

			storeID := ctx.Args().Get(0)

			service := clipkg.ContainerScope.MustGet(ctx.Context).C
			return service.SetCurrentStore(ctx.Context, storeID)
		},
	}
}
