package completion

import (
	"os"

	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func ListCompletion(prefix string) func(ctx *cli.Context) {
	return func(ctx *cli.Context) {
		// create own container since common context is not initialized yet
		c := clipkg.NewContainer(ctx)

		tree, err := c.StoreService.List(
			ctx.Context,
			store.ListParams{Path: prefix},
		)
		if err != nil {
			return
		}

		o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
		for _, p := range tree.Inline().Keys() {
			o.Printf(p)
		}
	}
}

func ListStoresCompletion(ctx *cli.Context) {
	if ctx.NArg() > 0 {
		return
	}

	service := clipkg.NewContainer(ctx).C

	stores, err := service.ListStores(ctx.Context)
	if err != nil {
		return
	}

	o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
	for _, s := range stores {
		var currentPrefix string
		if s.Current {
			currentPrefix = "*"
		}
		o.Printf(currentPrefix + string(s.ID))
	}
}
