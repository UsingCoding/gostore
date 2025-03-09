package mgnt

import (
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func pack() *cli.Command {
	return &cli.Command{
		Name:  "pack",
		Usage: "Pack store after it was unpacked",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-changes-check",
				Usage: "Skip changes checking of packed files. Will make pack faster, but commits even unmodified files, since encryption may add timestamp to encrypted payload",
			},
		},
		Category: cmd.MgmtCategory,
		Action: func(ctx *cli.Context) error {
			s := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

			return s.Pack(
				ctx.Context,
				store.PackParams{
					SkipChangesCheck: ctx.Bool("skip-changes-check"),
				},
			)
		},
	}
}
