package mgnt

import (
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
)

func pack() *cli.Command {
	return &cli.Command{
		Name:     "pack",
		Usage:    "Pack store after it was unpacked",
		Category: cmd.MgmtCategory,
		Action: func(ctx *cli.Context) error {
			s := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

			return s.Pack(ctx.Context)
		},
	}
}
