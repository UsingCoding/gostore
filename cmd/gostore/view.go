package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	appview "github.com/UsingCoding/gostore/internal/gostore/app/view"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/viewer"
)

func view() *cli.Command {
	return &cli.Command{
		Name:    "view",
		Aliases: nil,
		Usage:   "View secret in store",
		Description: `View secret in store via default system apps. For linux, gostore runs xdg-open to view secret
Can solve specific cases like view picture from store and e.t.c
NOTE: before run viewer app gostore put UNENCRYPTED secret data in tmp file and does not clean it after exiting.
Since apps like xdg-open does not blocking programs that called it or provide some info about opened resource`,
		Action: executeView,
		BashComplete: func(ctx *cli.Context) {
			if ctx.NArg() > 0 {
				return
			}

			printTree(ctx)
		},
	}
}

func executeView(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}
	path := ctx.Args().Get(0)

	var key maybe.Maybe[string]
	if ctx.Args().Len() > 1 {
		key = maybe.NewJust(ctx.Args().Get(1))
	}

	s, _ := newStoreService(ctx)

	v, err := viewer.NewViewer()
	if err != nil {
		return err
	}

	return appview.NewService(s, v).
		View(
			ctx.Context,
			path,
			key,
		)
}
