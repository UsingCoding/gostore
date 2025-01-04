package totp

import (
	"github.com/urfave/cli/v2"
)

func TOTP() *cli.Command {
	return &cli.Command{
		Name:  "totp",
		Usage: "Manage TOTP's",
		Subcommands: []*cli.Command{
			add(),
			passcode(),
		},
	}
}
