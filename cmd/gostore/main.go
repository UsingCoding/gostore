package main

import (
	"context"
	"encoding/json"
	stdlog "log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/cli/cmd/app"
	"github.com/UsingCoding/gostore/internal/cli/cmd/core"
	"github.com/UsingCoding/gostore/internal/cli/cmd/identity"
	"github.com/UsingCoding/gostore/internal/cli/cmd/mgnt"
	"github.com/UsingCoding/gostore/internal/cli/cmd/store"
	"github.com/UsingCoding/gostore/internal/cli/cmd/totp"
	"github.com/UsingCoding/gostore/internal/common/slices"

	"github.com/UsingCoding/gostore/internal/common/errors"
	"github.com/UsingCoding/gostore/internal/gostore/app/verbose"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func main() {
	ctx := context.Background()

	ctx = subscribeForKillSignals(ctx)

	err := runApp(ctx, os.Args)
	if err != nil {
		stdlog.Fatal(err)
	}
}

func runApp(ctx context.Context, args []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	a := &cli.App{
		Name:    app.ID,
		Version: app.Version,
		// do not use built-in version flag
		HideVersion:          true,
		Usage:                "Secrets store manager",
		EnableBashCompletion: true,
		Before:               BeforeHook,
		Commands: slices.Merge(
			app.Main(),
			core.Main(),
			mgnt.Main(),
			identity.Identity(),
			store.Store(),
			totp.TOTP(),
		),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "gostore-base-path",
				Usage: "Path to gostore dir",
				EnvVars: []string{
					"GOSTORE_STORE_BASE_PATH",
				},
				Value: path.Join(homeDir, ".gostore"),
			},
			&cli.StringFlag{
				Name:  "store-id",
				Usage: "Store ID",
				EnvVars: []string{
					"GOSTORE_STORE_ID",
				},
			},
			&cli.UintFlag{
				Name:    "verbose",
				Usage:   "Verbose mode: 1, 2, 3",
				Aliases: []string{"v"},
				EnvVars: []string{
					"GOSTORE_VERBOSE",
				},
				Action: func(_ *cli.Context, i uint) error {
					return verbose.Valid(i)
				},
			},
			&cli.StringFlag{
				Name:    "progress",
				Usage:   "Progress mode: auto|none",
				Aliases: []string{"p"},
				EnvVars: []string{
					"GOSTORE_PROGRESS",
				},
				Value: "auto",
			},
			&cli.StringFlag{
				Name:    "output",
				Usage:   "Output type: plain|json",
				Aliases: []string{"o"},
				EnvVars: []string{
					"GOSTORE_OUTPUT",
				},
				Value: "plain",
			},
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			defer func() {
				cli.HandleExitCoder(err)
			}()

			v := verbose.Ensure(c.Uint("verbose"))

			if v < verbose.Level1 {
				return
			}

			traces := errors.StackTraces(err)
			if len(traces) == 0 {
				return
			}

			printStackTraces(traces)
		},
	}

	return a.RunContext(ctx, args)
}

func subscribeForKillSignals(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
			signal.Stop(ch)
		case <-ch:
		}
	}()

	return ctx
}

func printStackTraces(traces []errors.Trace) {
	o := consoleoutput.
		New(os.Stdout, consoleoutput.WithNewline(true))

	for i, trace := range traces {
		traceStr, err := json.Marshal(trace)
		if err != nil {
			return
		}

		o.Printf("Trace: %d", i)
		o.Printf(string(traceStr))
	}
}
