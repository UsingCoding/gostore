package main

import (
	"encoding/json"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
	"github.com/urfave/cli/v2"
	"os"
)

func versionCmd() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "Show gostore version",
		Action: executeVersion,
	}
}

func executeVersion(c *cli.Context) error {
	v := struct {
		Version string `json:"version"`
	}{
		Version: version,
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
