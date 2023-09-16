package context

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func list() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List stores",
		Action:  executeList,
	}
}

func executeList(ctx *cli.Context) error {
	service := newConfigService(ctx)

	stores, err := service.ListStores(ctx.Context)
	if err != nil {
		return err
	}

	for _, store := range stores {
		msg := fmt.Sprintf("%s: %s\n", store.ID, store.Path)
		_, _ = os.Stdout.WriteString(msg)
	}

	return nil
}
