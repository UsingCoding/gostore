package main

import (
	"github.com/urfave/cli/v2"
)

func unpack() *cli.Command {
	return &cli.Command{
		Name:   "unpack",
		Usage:  "Unpack store for direct edits of secrets",
		Action: executeUnpack,
	}
}

func executeUnpack(ctx *cli.Context) error {
	s, _ := newStoreService(ctx)

	return s.Unpack(ctx.Context, makeCommonParams(ctx))
}
