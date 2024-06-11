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
	var path string
	if ctx.Args().Len() > 0 {
		path = ctx.Args().Get(0)
	}

	service, configService := newStoreService(ctx)

	tree, err := service.List(ctx.Context, store.ListParams{
		CommonParams: makeCommonParams(ctx),
		Path:         path,
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

	treePrinter := treeprint.NewWithRoot(root)

	recursiveList(treePrinter, tree)
	_, _ = os.Stdout.WriteString(treePrinter.String())

	return nil
}

func recursiveList(treePrinter treeprint.Tree, tree storage.Tree) {
	for _, entry := range tree {
		if len(entry.Children) == 0 {
			treePrinter.AddNode(entry.Name)
			continue
		}

		recursiveList(treePrinter.AddBranch(entry.Name), entry.Children)
	}
}
