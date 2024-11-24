package main

import (
	"encoding/json"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func versionCmd() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "Show gostore version",
		Action: executeVersion,
	}
}

func executeVersion(_ *cli.Context) error {
	v := struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}{
		Version: version,
		Commit:  commit,
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
