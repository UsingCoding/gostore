package main

import (
	"context"
	stdlog "log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/urfave/cli/v2"

	contextcmd "github.com/UsingCoding/gostore/cmd/gostore/context"
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
		Name:                 appID,
		Version:              version,
		Usage:                "Secrets store manager",
		EnableBashCompletion: true,
		Action:               repl,
		Commands: []*cli.Command{
			initCmd(),
			clone(),
			add(),
			get(),
			move(),
			copyCmd(),
			list(),
			remove(),
			sync(),
			contextcmd.Context(),
			completion(),
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

func newStoreService(ctx *cli.Context) (service.Service, config.Service) {
	gostoreBaseDir := ctx.String("gostore-base-path")

	storageManager := storage.NewManager()
	configService := config.NewService(
		infraconfig.NewStorage(gostoreBaseDir),
		gostoreBaseDir,
		storageManager,
	)

	return service.NewService(
		configService,
		storageManager,
		encryption.NewManager(),
		infrastore.NewManifestSerializer(),
		infrastore.NewSecretSerializer(),
	), configService
}

func optFromCtx[T any](ctx *cli.Context, key string) maybe.Maybe[T] {
	v := ctx.Generic(key)
	if v == nil {
		return maybe.Maybe[T]{}
	}

	t, ok := v.(T)
	if !ok {
		return maybe.Maybe[T]{}
	}

	return maybe.NewJust(t)
}

func optStringFromCtx(ctx *cli.Context, key string) maybe.Maybe[string] {
	v := ctx.String(key)
	if v == "" {
		return maybe.Maybe[string]{}
	}

	return maybe.NewJust(v)
}
