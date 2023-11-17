package editor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/gostore/app/edit"
)

func NewEditor() (edit.Editor, error) {
	args, err := findEditor()
	if err != nil {
		return nil, err
	}

	return &editor{
		args: args,
	}, nil
}

var (
	fallbackEditors = []string{"vim", "nano", "vi"}
)

type editor struct {
	args []string
}

func (e *editor) Edit(_ context.Context, p string, data []byte) ([]byte, error) {
	// Create temporary file for editing
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp dir for editing")
	}
	defer os.RemoveAll(tmpdir)

	temp, err := os.Create(path.Join(tmpdir, path.Base(p)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp file for editing")
	}
	defer temp.Close()

	tmpFilePath := temp.Name()

	_, err = temp.Write(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write to temp file")
	}

	srcHash, err := hashFile(tmpFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash source file")
	}

	err = temp.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to close temp file before editing")
	}

	err = e.run(tmpFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to run editor")
	}

	editedHash, err := hashFile(tmpFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash edited file")
	}

	if bytes.Equal(srcHash, editedHash) {
		return nil, errors.WithStack(edit.ErrNoChangesMade)
	}

	edited, err := os.ReadFile(tmpFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read edited file")
	}

	return edited, nil
}

func (e *editor) run(p string) error {
	args := append(e.args, p)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func hashFile(p string) ([]byte, error) {
	var result []byte
	file, err := os.Open(p)
	if err != nil {
		return result, err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}
	return hash.Sum(result), nil
}
