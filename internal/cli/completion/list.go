package completion

import (
	"os"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
	"github.com/urfave/cli/v2"
)

func ListCompletion(ctx *cli.Context) {
	container, ok := clipkg.ContainerScope.Get(ctx.Context)
	if !ok {
		return
	}

	tree, err := container.S.List(ctx.Context, store.ListParams{})
	if err != nil {
		return
	}

	o := consoleoutput.New(os.Stdout, consoleoutput.WithNewline(true))
	for _, p := range tree.Inline().Keys() {
		o.Printf(p)
	}
}
