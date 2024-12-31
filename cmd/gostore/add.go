package main

import (
	"io"
	stdos "os"
	"syscall"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func add() *cli.Command {
	return &cli.Command{
		Name:   "add",
		Usage:  "Add secret to current store",
		Action: executeAdd,
		BashComplete: func(ctx *cli.Context) {
			if ctx.NArg() > 0 {
				return
			}

			printTree(ctx)
		},
	}
}

func executeAdd(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}
	path := ctx.Args().Get(0)

	var key maybe.Maybe[string]
	if ctx.Args().Len() > 1 {
		key = maybe.NewJust(ctx.Args().Get(1))
	}

	var (
		data []byte
		err  error
	)
	if term.IsTerminal(syscall.Stdin) {
		o := consoleoutput.New(stdos.Stdout)
		o.Printf("Enter secret:")

		data, err = term.ReadPassword(syscall.Stdin)
		if err != nil {
			return errors.Wrap(err, "failed to read password")
		}
	} else {
		data, err = io.ReadAll(stdos.Stdin)
		if err != nil {
			return errors.Wrap(err, "failed to read from stdin")
		}
	}

	if len(data) == 0 {
		return errors.New("empty data")
	}

	service, _ := newStoreService(ctx)

	return service.Add(ctx.Context, store.AddParams{
		CommonParams: makeCommonParams(ctx),
		Path:         path,
		Key:          key,
		Data:         data,
	})
}
