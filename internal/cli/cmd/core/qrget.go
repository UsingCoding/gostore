package core

import (
	"os"

	"github.com/mdp/qrterminal/v3"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/cli/cmd"
	"github.com/UsingCoding/gostore/internal/cli/completion"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func qrget() *cli.Command {
	return &cli.Command{
		Name:         "qrget",
		Usage:        "Prints QR code with secret payload (supports displaying only one payload)",
		Category:     cmd.CoreCategory,
		BashComplete: completion.ListCompletion(""),
		Action:       executeQRGet,
	}
}

func executeQRGet(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return errors.New("not enough arguments")
	}
	path := ctx.Args().Get(0)

	var key maybe.Maybe[string]
	if ctx.Args().Len() > 1 {
		key = maybe.NewJust(ctx.Args().Get(1))
	}

	service := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

	secretsData, err := service.Get(ctx.Context, store.GetParams{
		SecretIndex: store.SecretIndex{
			Path: path,
			Key:  key,
		},
	})
	if err != nil {
		return err
	}

	if len(secretsData) == 0 {
		return errors.New("no secret payload found")
	}

	if len(secretsData) != 1 {
		return errors.New("qrget not supports multiple secret payloads")
	}

	qrterminal.Generate(
		string(secretsData[0].Payload),
		qrterminal.L,
		os.Stdout,
	)

	return nil
}
