package mgnt

import (
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
)

func unpack() *cli.Command {
	return &cli.Command{
		Name:     "unpack",
		Usage:    "Unpack store for direct edits of secrets",
		Category: cmd.MgmtCategory,
		Action: func(ctx *cli.Context) error {
			s := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

			return s.Unpack(ctx.Context)
		},
	}
}
