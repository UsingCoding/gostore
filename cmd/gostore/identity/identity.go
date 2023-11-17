package identity

import (
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	infraconfig "github.com/UsingCoding/gostore/internal/gostore/infrastructure/config"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/storage"
)

func Identity() *cli.Command {
	return &cli.Command{
		Name:  "identity",
		Usage: "Manage identities",
		Subcommands: []*cli.Command{
			export(),
			importCmd(),
		},
	}
}

func newConfigService(ctx *cli.Context) config.Service {
	gostoreBaseDir := ctx.String("gostore-base-path")

	return config.NewService(
		infraconfig.NewStorage(gostoreBaseDir),
		gostoreBaseDir,
		storage.NewManager(),
		encryption.NewManager(),
	)
}
