package storage

import (
	"context"

	"github.com/UsingCoding/gostore/internal/common/maybe"
)

type Type string

const (
	GITType = Type("git")
)

// Manager manages stores: create, delete, mount
type Manager interface {
	// Init creates storage locally
	Init(
		ctx context.Context,
		path string,
		remote maybe.Maybe[string],
		t Type,
	) (Storage, error)
	// Clone copies Storage from remote to path
	Clone(ctx context.Context, path string, remote string, t Type) (Storage, error)

	// Use local copy of store by path
	Use(ctx context.Context, path string) (Storage, error)

	// Remove local storage copy
	Remove(ctx context.Context, path string) error
}
