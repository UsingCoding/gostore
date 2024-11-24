package fuse

import (
	"context"

	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type FS interface {
	Serve(ctx context.Context) error
}

func New(
	config Config,
) FS {
	return &fs{
		c: config,
	}
}

type Config struct {
	Service store.Service

	MountPoint string
	ReadOnly   bool
}

type fs struct {
	c Config
}
