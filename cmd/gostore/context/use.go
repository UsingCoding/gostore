package context

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func use() *cli.Command {
	return &cli.Command{
		Name:      "use",
		Usage:     "Switch current store",
		UsageText: "use <STOREPATH>",
		Action:    executeUse,
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
