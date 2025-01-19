package core

import (
	"io"
	stdos "os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
	"github.com/UsingCoding/gostore/internal/cli/completion"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func add() *cli.Command {
	return &cli.Command{
		Name:         "add",
		Usage:        "Add secret to current store",
		Category:     cmd.CoreCategory,
		Action:       executeAdd,
		BashComplete: completion.ListCompletion(""),
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
	if term.IsTerminal(int(stdos.Stdin.Fd())) {
		o := consoleoutput.New(stdos.Stdin)
		o.Printf("Enter secret:")

		data, err = term.ReadPassword(int(stdos.Stdin.Fd()))
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

	service := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

	return service.Add(ctx.Context, store.AddParams{
		SecretIndex: store.SecretIndex{
			Path: path,
			Key:  key,
		},
		Data: data,
	})
}
