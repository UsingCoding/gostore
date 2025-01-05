package app

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/data"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
)

func completion() *cli.Command {
	return &cli.Command{
		Name:     "completion",
		Usage:    "Generate autocompletion",
		Category: cmd.AppCategory,
		Subcommands: []*cli.Command{
			{
				Name:   "bash",
				Usage:  "Generate autocompletion for bash",
				Action: executeCompletionBash,
			},
			{
				Name:   "zsh",
				Usage:  "Generate autocompletion for bash",
				Action: executeCompletionZsh,
			},
		},
	}
}

func executeCompletionBash(*cli.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, data.Bash(ID))
	return nil
}

func executeCompletionZsh(*cli.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, data.Zsh(ID))
	return nil
}
