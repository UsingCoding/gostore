package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func get() *cli.Command {
	return &cli.Command{
		Name:    "get",
		Aliases: []string{"cat"},
		Usage:   "Get secret from storage",
		Action:  executeGet,
	}
}

func executeGet(ctx *cli.Context) error {
	storePath := optStringFromCtx(ctx, "store")

	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}
	path := ctx.Args().Get(0)

	var key maybe.Maybe[string]
	if ctx.Args().Len() > 1 {
		key = maybe.NewJust(ctx.Args().Get(1))
	}

	service, _ := newStoreService(ctx)

	secretsData, err := service.Get(ctx.Context, store.GetParams{
		CommonParams: store.CommonParams{
			StorePath: storePath,
		},
		Path: path,
		Key:  key,
	})
	if err != nil {
		return err
	}

	if len(secretsData) == 0 {
		return errors.New("no secret payload found")
	}

	// if there is only one data in secret print it without kv formatting
	if len(secretsData) == 1 && secretsData[0].Default {
		s := secretsData[0]
		_, _ = os.Stdout.Write(s.Payload)
		return nil
	}

	for _, data := range secretsData {
		msg := fmt.Sprintf("%s: %s", data.Name, data.Payload)
		_, _ = os.Stdout.WriteString(msg)
	}

	return nil
}
