package core

import (
	"encoding/json"
	"os"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/xlab/treeprint"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/output"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func list() *cli.Command {
	return &cli.Command{
		Name:     "list",
		Aliases:  []string{"ls", "la"},
		Category: cmd.CoreCategory,
		Usage:    "List secrets in current store",
		Action:   executeList,
	}
}

func executeList(ctx *cli.Context) error {
	var path string
	if ctx.Args().Len() > 0 {
		path = ctx.Args().Get(0)
	}

	service := clipkg.ContainerScope.MustGet(ctx.Context).StoreService
	configService := clipkg.ContainerScope.MustGet(ctx.Context).C

	tree, err := service.List(ctx.Context, store.ListParams{
		Path: path,
	})
	if err != nil {
		return err
	}

	currentStoreID, err := configService.CurrentStoreID(ctx.Context)
	if err != nil {
		return err
	}

	o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))

	// just value without check since to use service.List we already has store in context
	root := string(maybe.Just(currentStoreID))
	if path != "" {
		root = path
	}

	switch output.FromCtx(ctx.Context) {
	case output.JSON:
		rootNode := jsonTreeNode{
			Name:  root,
			Elems: recursiveJSONList(tree),
		}

		data, err2 := json.Marshal(rootNode)
		if err2 != nil {
			return errors.Wrap(err2, "failed to marshal storage tree")
		}

		o.Printf(string(data))
	default:
		treePrinter := treeprint.NewWithRoot(root)

		recursiveList(treePrinter, tree)
		_, _ = os.Stdout.WriteString(treePrinter.String())
	}

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

func recursiveJSONList(tree storage.Tree) []jsonTreeNode {
	return slices.Map(tree, func(e storage.Entry) jsonTreeNode {
		return jsonTreeNode{
			Name:  e.Name,
			Elems: recursiveJSONList(e.Children),
		}
	})
}

type jsonTreeNode struct {
	Name  string         `json:"name"`
	Elems []jsonTreeNode `json:"children,omitempty"`
}
