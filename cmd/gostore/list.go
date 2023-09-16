package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"github.com/xlab/treeprint"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func list() *cli.Command {
	return &cli.Command{
		Name:   "ls",
		Usage:  "List secrets in current store",
		Action: executeList,
	}
}

func executeList(ctx *cli.Context) error {
	storePath := optStringFromCtx(ctx, "store")

	var path string
	if ctx.Args().Len() > 0 {
		path = ctx.Args().Get(0)
	}

	service, configService := newStoreService(ctx)

	entries, err := service.List(ctx.Context, store.ListParams{
		CommonParams: store.CommonParams{
			StorePath: storePath,
		},
		Path: path,
	})
	if err != nil {
		return err
	}

	currentStoreID, err := configService.CurrentStoreID(ctx.Context)
	if err != nil {
		return err
	}

	// just value without check since to use service.List we already has store in context
	root := string(maybe.Just(currentStoreID))
	if path != "" {
		root = path
	}

	tree := treeprint.NewWithRoot(root)

	recursiveList(tree, entries)
	_, _ = os.Stdout.WriteString(tree.String())

	return nil
}

func recursiveList(tree treeprint.Tree, entries []storage.Entry) {
	for _, entry := range entries {
		if len(entry.Children) == 0 {
			tree.AddNode(entry.Name)
			continue
		}

		recursiveList(tree.AddBranch(entry.Name), entry.Children)
	}
}
