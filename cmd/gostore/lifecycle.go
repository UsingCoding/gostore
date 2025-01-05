package main

import (
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/gostore/app/output"
	"github.com/UsingCoding/gostore/internal/gostore/app/progress"
)

func BeforeHook(c *cli.Context) error {
	funcs := []func(c *cli.Context) error{
		func(c *cli.Context) error {
			clipkg.ContainerCtx(c)
			return nil
		},
		initProgress,
		initOutput,
	}

	for _, f := range funcs {
		err := f(c)
		if err != nil {
			return err
		}
	}
	return nil
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
