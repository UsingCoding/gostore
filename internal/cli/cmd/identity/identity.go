package identity

import (
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/cli/cmd"
)

func Identity() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "identity",
			Usage:    "Manage identities",
			Category: cmd.ModuleCategory,
			Subcommands: []*cli.Command{
				export(),
				importCmd(),
			},
		},
	}
}
