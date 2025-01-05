package identity

import (
	"errors"
	"os"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func export() *cli.Command {
	return &cli.Command{
		Name:   "export",
		Usage:  "Export identities",
		Action: executeExport,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "recipients",
				Aliases:  []string{"r"},
				Required: true,
			},
		},
	}
}

func executeExport(ctx *cli.Context) error {
	service := clipkg.ContainerScope.MustGet(ctx.Context).C

	recipients := ctx.StringSlice("recipients")
	if len(recipients) == 0 {
		return errors.New("empty slice of recipients")
	}

	identities, err := service.ExportRawIdentity(ctx.Context, slices.Map(recipients, func(r string) encryption.Recipient {
		return encryption.Recipient(r)
	})...)
	if err != nil {
		return err
	}

	o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))

	for i := 0; i < len(identities); i++ {
		identity := identities[i]
		o.Printf(string(identity))
		if i != len(identities)-1 {
			o.Printf("") // empty line separator
		}
	}

	return nil
}
