package store

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
)

func list() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List stores",
		Action: func(ctx *cli.Context) error {
			service := clipkg.ContainerScope.MustGet(ctx.Context).C

			stores, err := service.ListStores(ctx.Context)
			if err != nil {
				return err
			}

			for _, store := range stores {
				var currentPtr string
				if store.Current {
					currentPtr = "*"
				}

				msg := fmt.Sprintf("%s%s: %s\n", currentPtr, store.ID, store.Path)
				_, _ = os.Stdout.WriteString(msg)
			}

			return nil
		},
	}
}
