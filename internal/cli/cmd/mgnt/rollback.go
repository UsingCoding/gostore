package mgnt

import (
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
)

func rollback() *cli.Command {
	return &cli.Command{
		Name:     "rollback",
		Usage:    "Rollback uncommitted changes and operations in store",
		Category: cmd.MgmtCategory,
		Action: func(ctx *cli.Context) error {
			s := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

			return s.Rollback(ctx.Context)
		},
	}
}
