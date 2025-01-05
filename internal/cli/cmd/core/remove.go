package core

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
	"github.com/UsingCoding/gostore/internal/cli/completion"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func remove() *cli.Command {
	return &cli.Command{
		Name:         "remove",
		Aliases:      []string{"rm"},
		Usage:        "Removes secret from storage or field from secret",
		Category:     cmd.CoreCategory,
		Action:       executeRemove,
		BashComplete: completion.ListCompletion(""),
	}
}

func executeRemove(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}
	path := ctx.Args().Get(0)

	var key maybe.Maybe[string]
	if ctx.Args().Len() > 1 {
		key = maybe.NewJust(ctx.Args().Get(1))
	}

	service := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

	return service.Remove(ctx.Context, store.RemoveParams{
		Path: path,
		Key:  key,
	})
}
