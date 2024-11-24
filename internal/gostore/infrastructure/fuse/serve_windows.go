package fuse

import (
	"context"

	"github.com/pkg/errors"
)

func (fs fs) Serve(ctx context.Context) (err error) {
	return errors.Errorf("FUSE mount is not supported on Windows")
}
