package context

import (
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func remove() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Removes store local copy",
		UsageText: "rm <STORE_ID>",
		Action:    executeRemove,
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
				o.Printf(string(store.ID))
			}
		},
	}
}

func executeRemove(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}

	storeID := ctx.Args().Get(0)

	service := newConfigService(ctx)

	return service.RemoveStore(ctx.Context, config.StoreID(storeID))
}
