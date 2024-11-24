//go:build !windows

package fuse

import (
	"context"
	"os"
	stdslices "slices"
	"syscall"
	"time"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/anacrolix/fuse"
	fusefs "github.com/anacrolix/fuse/fs"

	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

// Dir implements both Node and Handle for the hello file.
type Dir struct {
	r    root
	path string
}

func (d *Dir) Attr(_ context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0o755
	a.Mtime = time.Now()
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	path := name
	if d.path != "" {
		path = d.path + "/" + name
	}

	entries, err := d.r.service.List(ctx, store.ListParams{
		Path: d.path,
	})
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		// d,path not found
		return nil, syscall.ENOENT
	}

	i := stdslices.IndexFunc(entries, func(e storage.Entry) bool {
		return e.Name == name
	})
	if i == -1 {
		return nil, syscall.ENOENT
	}

	entry := entries[i]

	isDir := len(entry.Children) != 0

	if isDir {
		return &Dir{r: d.r, path: path}, nil
	}
	return &File{r: d.r, path: path}, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	entries, err := d.r.service.List(ctx, store.ListParams{
		Path: d.path,
	})
	if err != nil {
		return nil, err
	}

	return slices.Map(entries, func(e storage.Entry) fuse.Dirent {
		direntType := fuse.DT_File
		if len(e.Children) != 0 {
			direntType = fuse.DT_Dir
		}

		return fuse.Dirent{
			Type: direntType,
			Name: e.Name,
		}
	}), nil
}
