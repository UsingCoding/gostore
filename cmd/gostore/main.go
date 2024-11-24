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

	"github.com/UsingCoding/gostore/internal/common/errors"
	"github.com/UsingCoding/gostore/internal/gostore/app/output"
	"github.com/UsingCoding/gostore/internal/gostore/app/progress"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/app/verbose"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"

	contextcmd "github.com/UsingCoding/gostore/cmd/gostore/context"
	identitycmd "github.com/UsingCoding/gostore/cmd/gostore/identity"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	"github.com/UsingCoding/gostore/internal/gostore/app/service"
	infraconfig "github.com/UsingCoding/gostore/internal/gostore/infrastructure/config"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/storage"
	infrastore "github.com/UsingCoding/gostore/internal/gostore/infrastructure/store"
)

const (
	appID = "gostore"
)

var (
	version = "UNKNOWN"
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

	app := &cli.App{
		Name:    appID,
		Version: version,
		// do not use built-in version flag
		HideVersion:          true,
		Usage:                "Secrets store manager",
		EnableBashCompletion: true,
		Action:               repl,
		Before: func(c *cli.Context) error {
			err = initProgress(c)
			if err != nil {
				return err
			}
			return initOutput(c)
		},
		Commands: []*cli.Command{
			versionCmd(),
			initCmd(),
			clone(),
			add(),
			get(),
			edit(),
			view(),
			move(),
			copyCmd(),
			list(),
			remove(),
			unpack(),
			pack(),
			sync(),
			rollback(),
			contextcmd.Context(),
			identitycmd.Identity(),
			completion(),
			mount(),
		},
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

	return app.RunContext(ctx, args)
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

func newStoreService(ctx *cli.Context) (s service.Service, c config.Service) {
	gostoreBaseDir := ctx.String("gostore-base-path")

	storageManager := storage.NewManager()
	c = config.NewService(
		infraconfig.NewStorage(gostoreBaseDir),
		gostoreBaseDir,
		storageManager,
		encryption.NewManager(),
	)
	s = service.NewService(
		c,
		storageManager,
		encryption.NewManager(),
		infrastore.NewManifestSerializer(),
		infrastore.NewSecretSerializer(),
	)

	return s, c
}

func makeCommonParams(ctx *cli.Context) store.CommonParams {
	return store.CommonParams{
		StorePath: maybe.Maybe[string]{},
		StoreID:   optStringFromCtx(ctx, "store-id"),
	}
}

func initProgress(c *cli.Context) error {
	p, err := progress.Init(progress.Mode(c.String("progress")))
	if err != nil {
		return err
	}

	c.Context = progress.ToCtx(c.Context, p)

	return nil
}

func initOutput(c *cli.Context) error {
	ctx, err := output.InitToCtx(c.Context, c.String("output"))
	if err != nil {
		return err
	}

	c.Context = ctx
	return nil
}

func optStringFromCtx(ctx *cli.Context, key string) maybe.Maybe[string] {
	v := ctx.String(key)
	if v == "" {
		return maybe.Maybe[string]{}
	}

	return maybe.NewJust(v)
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
