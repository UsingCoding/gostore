package core

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
	"github.com/UsingCoding/gostore/internal/cli/completion"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func move() *cli.Command {
	return &cli.Command{
		Name:         "move",
		Aliases:      []string{"mv"},
		Usage:        "Moves path in store",
		UsageText:    "mv <src> <dst>",
		Category:     cmd.CoreCategory,
		Action:       executeMove,
		BashComplete: completion.ListCompletion,
	}
}

func executeMove(ctx *cli.Context) error {
	if ctx.Args().Len() < 2 {
		return errors.New("expected exactly 2 arguments")
	}

	src := ctx.Args().Get(0)
	dst := ctx.Args().Get(1)

	service := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

	return service.Move(ctx.Context, store.MoveParams{
		Src: src,
		Dst: dst,
	})
}
