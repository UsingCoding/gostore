package fuse

import (
	"context"
	"slices"
	"time"

	"github.com/anacrolix/fuse"
	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

// File implements both Node and Handle for the hello file.
type File struct {
	r    root
	path string

	// Buffer for write operations
	buffer []byte
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	data, err := f.r.service.Get(ctx, store.GetParams{
		Path: f.path,
	})
	if err != nil {
		return err
	}

	a.Mode = 0644
	if len(data) > 0 {
		i := slices.IndexFunc(data, func(s store.SecretData) bool {
			return s.Default
		})
		if i != -1 {
			// set size only from default payload
			a.Size = uint64(len(data[i].Payload))
		}

	}
	a.Mtime = time.Now()
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	data, err := f.r.service.Get(ctx, store.GetParams{
		Path: f.path,
	})
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []byte{}, nil
	}

	i := slices.IndexFunc(data, func(s store.SecretData) bool {
		return s.Default
	})
	if i == -1 {
		return []byte{}, nil
	}

	return data[i].Payload, nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	// Extend buffer if needed
	necessaryBufferSize := req.Offset + int64(len(req.Data))

	if int64(len(f.buffer)) < necessaryBufferSize {
		newBuf := make([]byte, necessaryBufferSize)
		copy(newBuf, f.buffer)
		f.buffer = newBuf
	}

	// Write data at offset
	copy(f.buffer[req.Offset:], req.Data)
	resp.Size = len(req.Data)
	return nil
}

func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	if len(f.buffer) > 0 {
		err := f.r.service.Add(ctx, store.AddParams{
			Path: f.path,
			Data: f.buffer,
		})
		if err != nil {
			return errors.Wrap(err, "failed to save file")
		}
		f.buffer = nil
	}
	return nil
}
