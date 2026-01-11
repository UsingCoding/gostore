package tui

import (
	"context"

	ui "github.com/metaspartan/gotui/v5"
	"github.com/pkg/errors"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/gostore/app/usecase/edit"
	infraeditor "github.com/UsingCoding/gostore/internal/gostore/infrastructure/editor"
)

func TUI(ctx context.Context) error {
	container := clipkg.ContainerScope.MustGet(ctx)

	if err := container.C.Init(ctx); err != nil {
		return err
	}

	var editService edit.Service
	editor, err := infraeditor.NewEditor()
	if err == nil {
		editService = edit.NewService(container.StoreService, editor)
	}

	dashboard := newDashboard(ctx, container.C, container.StoreService, editService)
	app := ui.NewApp()
	app.SetRoot(dashboard, true)

	err = app.Run()
	return errors.Wrap(err, "tui")
}
