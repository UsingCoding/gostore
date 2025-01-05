package totp

import (
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	clicompletion "github.com/UsingCoding/gostore/internal/cli/completion"
	apptotp "github.com/UsingCoding/gostore/internal/gostore/app/usecase/totp"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func passcode() *cli.Command {
	return &cli.Command{
		Name:         "passcode",
		Usage:        "Generate totp passcode",
		BashComplete: clicompletion.ListCompletion,
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() < 1 {
				return errors.New("not enough arguments")
			}

			path := ctx.Args().First()

			service := clipkg.ContainerScope.MustGet(ctx.Context).StoreService

			o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))

			p, err := apptotp.NewService(service).
				Passcode(
					ctx.Context,
					path,
				)
			if err != nil {
				return err
			}

			o.Printf("TOTP p: %s", p)

			return nil
		},
	}
}
