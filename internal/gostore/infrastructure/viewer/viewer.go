package viewer

import (
	"context"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/gostore/app/view"
)

func NewViewer() (view.Viewer, error) {
	cmd, err := appForView()
	if err != nil {
		return nil, err
	}

	return &viewer{cmd: cmd}, nil
}

type viewer struct {
	cmd string
}

func (v *viewer) View(_ context.Context, p string, data []byte) error {
	// Create temporary file for editing
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		return errors.Wrap(err, "failed to create temp dir for editing")
	}

	// Do no remove temp dir since xdg-open don`t block execution of program that called it.
	// So we can`t remove temporary files

	temp, err := os.Create(path.Join(tmpdir, path.Base(p)))
	if err != nil {
		return errors.Wrap(err, "failed to create temp file for editing")
	}
	defer temp.Close()

	tmpFilePath := temp.Name()

	_, err = temp.Write(data)
	if err != nil {
		return errors.Wrap(err, "failed to write to temp file")
	}

	err = temp.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close temp file before editing")
	}

	err = v.run(tmpFilePath)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 2)

	return nil
}

func (v *viewer) run(p string) error {
	cmd := exec.Command(v.cmd, p)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
