package edit

import (
	"context"
	stderrors "errors"
)

var (
	ErrNoChangesMade = stderrors.New("no changes made")
)

type Editor interface {
	Edit(ctx context.Context, p string, data []byte) ([]byte, error)
}
