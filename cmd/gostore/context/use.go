package context

import (
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func use() *cli.Command {
	return &cli.Command{
		Name:      "use",
		Usage:     "Switch current store",
		UsageText: "use <STOREPATH>",
		Action:    executeUse,
		BashComplete: func(ctx *cli.Context) {
			if ctx.NArg() > 0 {
				return
			}

			stores, err := newConfigService(ctx).ListStores(ctx.Context)
			if err != nil {
				return
			}

			o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
			for _, store := range stores {
				var currentPrefix string
				if store.Current {
					currentPrefix = "*"
				}
				o.Printf(currentPrefix + string(store.ID))
			}
		},
	}
}

func executeUse(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}

	storeID := ctx.Args().Get(0)

	service := newConfigService(ctx)

	return service.SetCurrentStore(ctx.Context, storeID)
}
