package store

import (
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/cli/cmd"
)

func Store() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "store",
			Aliases:  []string{"stores"},
			Usage:    "Manage stores",
			Category: cmd.ModuleCategory,
			Subcommands: []*cli.Command{
				use(),
				list(),
				initCmd(),
				clone(),
				remove(),
			},
		},
	}
}
