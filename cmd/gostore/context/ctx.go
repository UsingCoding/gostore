package context

import (
	"github.com/urfave/cli/v2"

	"gostore/internal/gostore/app/config"
	infraconfig "gostore/internal/gostore/infrastructure/config"
)

func Context() *cli.Command {
	return &cli.Command{
		Name:  "stores",
		Usage: "Manage stores",
		Subcommands: []*cli.Command{
			use(),
			list(),
		},
	}
}

func newConfigService(ctx *cli.Context) config.Service {
	gostoreBaseDir := ctx.String("gostore-base-path")

	return config.NewService(
		infraconfig.NewStorage(gostoreBaseDir),
		gostoreBaseDir,
	)
}
