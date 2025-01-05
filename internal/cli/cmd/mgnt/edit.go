package mgnt

import (
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
	completionpkg "github.com/UsingCoding/gostore/internal/cli/completion"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	appedit "github.com/UsingCoding/gostore/internal/gostore/app/usecase/edit"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/editor"
)

func edit() *cli.Command {
	return &cli.Command{
		Name:         "edit",
		Usage:        "Edit secrets",
		UsageText:    "edit <SECRET_ID> ?<KEY>",
		Category:     cmd.MgmtCategory,
		Action:       executeEdit,
		BashComplete: completionpkg.ListCompletion(""),
	}
}

func executeEdit(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}
	path := ctx.Args().Get(0)

	var key maybe.Maybe[string]
	if ctx.Args().Len() > 1 {
		key = maybe.NewJust(ctx.Args().Get(1))
	}

	s := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

	e, err := editor.NewEditor()
	if err != nil {
		return err
	}

	o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))

	err = appedit.NewService(s, e).Edit(ctx.Context, store.SecretIndex{
		Path: path,
		Key:  key,
	})
	if errors.Is(err, appedit.ErrNoChangesMade) {
		o.Printf("No changes made")
		return nil
	}
	return err
}
