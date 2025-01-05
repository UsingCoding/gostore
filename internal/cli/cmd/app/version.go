package app

import (
	"encoding/json"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/cli/cmd"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func version() *cli.Command {
	return &cli.Command{
		Name:     "version",
		Usage:    "Show gostore version",
		Category: cmd.AppCategory,
		Action:   executeVersion,
	}
}

func executeVersion(_ *cli.Context) error {
	v := struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}{
		Version: Version,
		Commit:  Commit,
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	consoleoutput.
		New(os.Stdout, consoleoutput.WithNewline(true)).
		Printf(string(data))

	return nil
}
