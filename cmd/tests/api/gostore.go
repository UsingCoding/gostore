package api

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

func (a api) gostore(in input) (output, error) {
	//nolint:gosec
	c := exec.Command(
		path.Join("..", "..", gostorePath),
		in.args...,
	)

	c.Env = append(
		os.Environ(),
		fmt.Sprintf("GOSTORE_STORE_BASE_PATH=%s", a.basePath),
	)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	if in.stdin != nil {
		c.Stdin = in.stdin
	}
	c.Stdout = stdout
	c.Stderr = stderr

	o := output{
		stdout: stdout,
		stderr: stderr,
	}

	err := c.Run()
	if err != nil {
		err = exitErr{
			err:    err,
			output: o,
		}
	}

	return o, err
}

type input struct {
	args  []string
	stdin io.Reader
}

type output struct {
	stdout, stderr *bytes.Buffer
}

type exitErr struct {
	err    error
	output output
}

func (e exitErr) Error() string {
	msg := []string{
		"err: %s",
		"stdout:", "%s",
		"stderr:", "%s",
	}

	return fmt.Sprintf(
		strings.Join(msg, "\n"),
		e.err.Error(),
		e.output.stdout,
		e.output.stderr,
	)
}
