package main

import (
	"github.com/urfave/cli/v2"
)

func rollback() *cli.Command {
	return &cli.Command{
		Name:   "rollback",
		Usage:  "Rollback uncommited changes and operations in store",
		Action: executeRollback,
	}
}

func executeRollback(ctx *cli.Context) error {
	s, _ := newStoreService(ctx)

	return s.Rollback(ctx.Context, makeCommonParams(ctx))
}
