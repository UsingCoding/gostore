package totp

import (
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/cli/cmd"
)

func TOTP() []*cli.Command {
	return []*cli.Command{
		{
			Name:     "totp",
			Usage:    "Manage TOTP's",
			Category: cmd.ModuleCategory,
			Subcommands: []*cli.Command{
				add(),
				passcode(),
			},
		},
	}
}
