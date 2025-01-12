package totp

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	clicompletion "github.com/UsingCoding/gostore/internal/cli/completion"
	apptotp "github.com/UsingCoding/gostore/internal/gostore/app/usecase/totp"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func passcode() *cli.Command {
	return &cli.Command{
		Name:         "passcode",
		Usage:        "Generate totp passcode",
		BashComplete: clicompletion.ListCompletion("totp"),
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() < 1 {
				return errors.New("not enough arguments")
			}

			service := clipkg.ContainerScope.MustGet(ctx.Context).StoreService
			pv, err := apptotp.
				NewService(service).
				PasscodeView(
					ctx.Context,
					ctx.Args().First(),
				)
			if err != nil {
				return err
			}

			o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
			o.Printf("Time-based One Time Password")

			switch term.IsTerminal(int(os.Stdout.Fd())) {
			case true:
				return drawCountdown(ctx.Context, pv)
			default:
				code, err2 := pv.GeneratePasscode()
				if err2 != nil {
					return err2
				}

				o.Printf(
					"Code: %s Countdown: %ds",
					code,
					pv.LastCountdown,
				)

				return nil
			}
		},
	}
}

func drawCountdown(ctx context.Context, pv apptotp.PasscodeView) error {
	countdown := pv.LastCountdown
	code, err := pv.GeneratePasscode()
	if err != nil {
		return err
	}

	o := consoleoutput.New(os.Stdout)
	firstIteration := true

	for {
		// \r brings cursor to beginning of the line
		// \033[K clear line
		escape := "\r\033[K"

		// first iteration
		if firstIteration {
			escape = ""
			firstIteration = false
		}

		if countdown == 0 {
			code, err = pv.GeneratePasscode()
			if err != nil {
				return err
			}
			countdown = pv.Period
		}

		o.Printf("%s%s - %d", escape, code, countdown)

		select {
		case <-ctx.Done():
			// print new line
			o.Printf("\n")

			return nil
		case <-time.After(time.Second):
			countdown--
		}
	}
}
