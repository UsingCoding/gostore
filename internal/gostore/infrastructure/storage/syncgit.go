package storage

import (
	"context"
	"sync"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
)

func newSyncGit(s *gitStorage) storage.Storage {
	return &syncGit{
		gitStorage: s,
	}
}

// sync git storage for several operations due internal bugs leading to panic
type syncGit struct {
	*gitStorage
	m sync.Mutex
}

func (g *syncGit) GetLatest(ctx context.Context, p string) (maybe.Maybe[[]byte], error) {
	g.m.Lock()
	defer g.m.Unlock()

	return g.gitStorage.GetLatest(ctx, p)
}
