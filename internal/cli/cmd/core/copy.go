package core

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
	"github.com/UsingCoding/gostore/internal/cli/completion"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func copyCmd() *cli.Command {
	return &cli.Command{
		Name:         "copy",
		Aliases:      []string{"cp"},
		Usage:        "Copies path in store",
		UsageText:    "cp <src> <dst>",
		Category:     cmd.CoreCategory,
		Action:       executeCopy,
		BashComplete: completion.ListCompletion(""),
	}
}

func executeCopy(ctx *cli.Context) error {
	if ctx.Args().Len() < 2 {
		return errors.New("expected exactly 2 arguments")
	}

	src := ctx.Args().Get(0)
	dst := ctx.Args().Get(1)

	service := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

	return service.Copy(ctx.Context, store.CopyParams{
		Src: src,
		Dst: dst,
	})
}
