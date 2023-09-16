package context

import (
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	infraconfig "github.com/UsingCoding/gostore/internal/gostore/infrastructure/config"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/storage"
)

func Context() *cli.Command {
	return &cli.Command{
		Name:  "stores",
		Usage: "Manage stores",
		Subcommands: []*cli.Command{
			use(),
			list(),
			remove(),
		},
	}
}

func newConfigService(ctx *cli.Context) config.Service {
	gostoreBaseDir := ctx.String("gostore-base-path")

	return config.NewService(
		infraconfig.NewStorage(gostoreBaseDir),
		gostoreBaseDir,
		storage.NewManager(),
	)
}
