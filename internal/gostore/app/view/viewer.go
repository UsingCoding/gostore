package view

import (
	"context"
)

type Viewer interface {
	View(ctx context.Context, p string, data []byte) error
}
