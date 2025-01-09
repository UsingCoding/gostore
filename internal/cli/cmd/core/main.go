package core

import (
	"github.com/urfave/cli/v2"
)

func Main() []*cli.Command {
	return []*cli.Command{
		add(),
		copyCmd(),
		get(),
		qrget(),
		list(),
		move(),
		remove(),
	}
}
