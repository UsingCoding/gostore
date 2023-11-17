package identity

import (
	"io"
	stdos "os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func importCmd() *cli.Command {
	return &cli.Command{
		Name:      "import",
		Usage:     "Import identity",
		UsageText: "import < identity.plain",
		Action:    executeImport,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "provider",
				Usage:    "Identity provider: age",
				Required: true,
				Aliases:  []string{"p"},
			},
		},
	}
}

func executeImport(ctx *cli.Context) error {
	provider := ctx.String("provider")

	data, err := io.ReadAll(stdos.Stdin)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return errors.New("empty stdin")
	}

	err = newConfigService(ctx).
		ImportRawIdentity(
			ctx.Context,
			encryption.Provider(provider),
			data,
		)
	if err != nil {
		return err
	}

	o := consoleoutput.New(stdos.Stdout, consoleoutput.WithNewline(true))
	o.OKf("Identity imported")

	return nil
}
