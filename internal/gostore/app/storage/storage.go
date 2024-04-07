package storage

import (
	"context"

	"github.com/UsingCoding/gostore/internal/common/maybe"
)

type Storage interface {
	// Store data to storage
	Store(ctx context.Context, path string, data []byte) error
	// Remove path from storage
	Remove(ctx context.Context, path string) error

	// Copy path in storage
	Copy(ctx context.Context, src, dst string) error
	// Move path in storage
	Move(ctx context.Context, src, dst string) error

	// Get data from storage
	Get(ctx context.Context, path string) (maybe.Maybe[[]byte], error)
	// GetLatest reruns latest version of object
	GetLatest(ctx context.Context, p string) (maybe.Maybe[[]byte], error)
	// List storage entries
	List(ctx context.Context, path string) ([]Entry, error)

	// AddRemote to storage. remoteAddr depends on storage implementation
	AddRemote(ctx context.Context, remoteName string, remoteAddr string) error
	// HasRemote reports that Storage has remote
	HasRemote(ctx context.Context) (bool, error)
	// Push storage to remote if there is one
	Push(ctx context.Context) error
	// Pull changes from remote
	Pull(ctx context.Context) error

	// Commit changes to storage. Semantics depends on storage implementation
	Commit(ctx context.Context, msg string) error
	// Rollback all uncommitted changes
	Rollback(ctx context.Context) error
}

type Entry struct {
	Name     string
	Children []Entry
}
