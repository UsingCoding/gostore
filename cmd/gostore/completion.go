package main

import (
	"fmt"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/data"
)

func completion() *cli.Command {
	return &cli.Command{
		Name:  "completion",
		Usage: "Generate autocompletion",
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
	_, _ = fmt.Fprintln(os.Stdout, data.Bash(appID))
	return nil
}

func executeCompletionZsh(*cli.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, data.Zsh(appID))
	return nil
}

func printTree(ctx *cli.Context) {
	service, _ := newStoreService(ctx)

	tree, err := service.List(ctx.Context, store.ListParams{})
	if err != nil {
		return
	}

	o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
	for _, p := range tree.Inline().Keys() {
		o.Printf(p)
	}
}
