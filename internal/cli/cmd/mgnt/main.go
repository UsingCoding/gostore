package mgnt

import (
	"github.com/urfave/cli/v2"
)

func Main() []*cli.Command {
	return []*cli.Command{
		edit(),
		view(),
		pack(),
		unpack(),
		sync(),
		rollback(),
		mount(),
	}
}
