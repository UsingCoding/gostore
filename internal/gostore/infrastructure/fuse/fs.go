package fuse

import (
	"context"
	stderrors "errors"

	"github.com/anacrolix/fuse"
	fusefs "github.com/anacrolix/fuse/fs"
	"github.com/pkg/errors"

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

func (fs fs) Serve(ctx context.Context) (err error) {
	const (
		fsID = "gostore"
	)

	options := []fuse.MountOption{
		fuse.FSName(fsID),
		fuse.Subtype(fsID),
	}
	if fs.c.ReadOnly {
		options = append(options, fuse.ReadOnly())
	}

	// Initialize FUSE connection
	c, err := fuse.Mount(
		fs.c.MountPoint,
		options...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to mount filesystem")
	}

	go func() {
		<-ctx.Done()
		err = stderrors.Join(err, fs.shutdown())
	}()

	err = fusefs.Serve(c, root{
		service: fs.c.Service,
	})

	// closing conn *can panic* on specific os and systems due to internal error
	// close connection but ignore the panic
	defer func() {
		_ = recover()
	}()
	_ = c.Close()

	err = errors.Wrap(err, "failed to serve filesystem")
	return err
}

func (fs fs) shutdown() error {
	return fuse.Unmount(fs.c.MountPoint)
}

type root struct {
	service store.Service
}

func (r root) Root() (fusefs.Node, error) {
	return &Dir{r: r, path: ""}, nil
}
