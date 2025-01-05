package app

import (
	"github.com/urfave/cli/v2"
)

const (
	ID = "gostore"
)

var (
	Version = "dev"
	Commit  = "none"
)

func Main() []*cli.Command {
	return []*cli.Command{
		version(),
		completion(),
	}
}
