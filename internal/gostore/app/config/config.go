package config

import (
	"context"
	stderrors "errors"

	"gostore/internal/common/maybe"
	"gostore/internal/gostore/app/encryption"
)

type StoreID string

var (
	ErrConfigNotFound = stderrors.New("config not found")
)

type Config struct {
	Context maybe.Maybe[StoreID] // store ID
	Stores  []Store              // paths to stores

	Identities []encryption.Identity
}

type Store struct {
	ID   StoreID
	Path string
}

type Storage interface {
	Load(ctx context.Context) (Config, error)
	Store(ctx context.Context, config Config) error
}
