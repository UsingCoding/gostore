package viewer

import (
	"context"
	"os"
	"os/exec"
	"path"

	"github.com/UsingCoding/gostore/internal/common/slices"
	"github.com/UsingCoding/gostore/internal/gostore/app/usecase/view"
	"github.com/pkg/errors"
)

func NewViewer() (view.Viewer, error) {
	args, wait, err := appForView()
	if err != nil {
		return nil, err
	}

	return &viewer{args: args, wait: wait}, nil
}

type viewer struct {
	args []string
	wait bool
}

func (v viewer) View(_ context.Context, p string, data []byte) error {
	// Create temporary file for viewing
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

	// v.wait means that view app waits until user close it, so we can remove tmp files
	if v.wait {
		_ = os.Remove(tmpFilePath)
	}

	return nil
}

func (v viewer) run(p string) error {
	cmdName, args := slices.Decompose(v.args)

	//nolint:gosec
	cmd := exec.Command(cmdName, append(args, p)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
