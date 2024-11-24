package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
	infrafuse "github.com/UsingCoding/gostore/internal/gostore/infrastructure/fuse"
)

func mount() *cli.Command {
	return &cli.Command{
		Name:      "mount",
		Usage:     "Mount store as filesystem. BETA mode",
		UsageText: "mount <MOUNT_POINT>",
		Action:    executeMount,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "read-only",
				Aliases: []string{"ro"},
				Usage:   "Mount filesystem in read-only mode",
			},
		},
	}
}

func executeMount(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return errors.New("mount point required")
	}

	mountPoint := ctx.Args().Get(0)
	readOnly := ctx.Bool("read-only")

	// Check if mount point exists
	if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
		return errors.Errorf("mount point '%s' does not exist", mountPoint)
	}

	service, _ := newStoreService(ctx)

	fs := infrafuse.New(infrafuse.Config{
		Service:    service,
		MountPoint: mountPoint,
		ReadOnly:   readOnly,
	})

	o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
	o.Printf("Mounted at %s", mountPoint)
	o.Printf("Press Ctrl+C to unmount")

	return fs.Serve(ctx.Context)
}
