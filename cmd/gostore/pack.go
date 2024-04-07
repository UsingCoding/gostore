package main

import (
	"github.com/urfave/cli/v2"
)

func pack() *cli.Command {
	return &cli.Command{
		Name:   "pack",
		Usage:  "Pack store after it was unpacked",
		Action: executePack,
	}
}

func executePack(ctx *cli.Context) error {
	s, _ := newStoreService(ctx)

	return s.Pack(ctx.Context, makeCommonParams(ctx))
}
