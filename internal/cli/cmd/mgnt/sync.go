package mgnt

import (
	"os"

	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func sync() *cli.Command {
	return &cli.Command{
		Name:     "sync",
		Usage:    "Sync store with remote",
		Category: cmd.MgmtCategory,
		Action: func(ctx *cli.Context) error {
			service := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

			err := service.Sync(ctx.Context)
			if err != nil {
				return err
			}

			o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
			o.OKf("Synced")

			return nil
		},
	}
}
