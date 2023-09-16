package storage

import (
	"context"
	stdos "os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/pkg/errors"

	"gostore/internal/common/maybe"
	"gostore/internal/gostore/app/storage"
)

const (
	remoteName = "origin"
)

func NewManager() storage.Manager {
	return &manager{}
}

type manager struct{}

func (m *manager) Init(
	_ context.Context,
	p string,
	remote maybe.Maybe[string],
	t storage.Type,
) (storage.Storage, error) {
	ok, err := exists(p)
	if err != nil {
		return nil, err
	}

	if ok {
		return nil, errors.Errorf("path %s already exists", p)
	}

	// check existence of parent directory
	if e, err2 := exists(path.Dir(p)); !e || err2 != nil {
		if err2 != nil {
			return nil, err2
		}

		err2 = stdos.MkdirAll(path.Dir(p), stdos.ModePerm)
		if err2 != nil {
			return nil, errors.Wrapf(err2, "failed to create subdirectories for path %s", p)
		}
	}

	switch t {
	case storage.GITType:
		return m.initGitStorage(p, remote)
	default:
		return nil, errors.Errorf("unsupported storage type %s", t)
	}
}

func (m *manager) Clone(ctx context.Context, p string, remote string, t storage.Type) (storage.Storage, error) {
	ok, err := exists(p)
	if err != nil {
		return nil, err
	}

	if ok {
		return nil, errors.Errorf("path %s already exists", p)
	}

	// check existence of parent directory
	if e, err2 := exists(path.Dir(p)); !e || err2 != nil {
		if err2 != nil {
			return nil, err2
		}

		err2 = stdos.MkdirAll(path.Dir(p), stdos.ModePerm)
		if err2 != nil {
			return nil, errors.Wrapf(err2, "failed to create subdirectories for path %s", p)
		}
	}

	switch t {
	case storage.GITType:
		return m.cloneGitStorage(ctx, p, remote)
	default:
		return nil, errors.Errorf("unsupported storage type %s", t)
	}
}

func (m *manager) Use(_ context.Context, p string) (storage.Storage, error) {
	ok, err := exists(p)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.Errorf("path %s not exists", p)
	}

	repo, err := git.PlainOpen(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open repo")
	}

	return &gitStorage{
		repo:    repo,
		repoDir: p,
	}, nil
}

func (m *manager) Remove(ctx context.Context, p string) error {
	_, err := m.Use(ctx, p)
	if err != nil {
		return err
	}

	err = stdos.RemoveAll(p)
	return errors.Wrapf(err, "failed to remote storage at %s", p)
}

func (m *manager) initGitStorage(p string, remote maybe.Maybe[string]) (storage.Storage, error) {
	repo, err := git.PlainInit(p, false)
	if err != nil {
		return nil, errors.New("failed to init store repo")
	}

	if maybe.Valid(remote) {
		_, err2 := repo.CreateRemote(&config.RemoteConfig{
			Name: remoteName,
			URLs: []string{maybe.Just(remote)},
		})
		if err2 != nil {
			return nil, errors.Wrapf(err2, "failed to add remote %s to repo", maybe.Just(remote))
		}
	}

	return &gitStorage{
		repo:    repo,
		repoDir: p,
	}, nil
}

func (m *manager) cloneGitStorage(ctx context.Context, p string, remote string) (storage.Storage, error) {
	repo, err := git.PlainCloneContext(
		ctx,
		p,
		false,
		&git.CloneOptions{
			URL:        remote,
			Auth:       nil, // use ssh-agent as auth
			RemoteName: remoteName,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to clone store repo")
	}

	return &gitStorage{
		repo:    repo,
		repoDir: p,
	}, nil
}

func exists(path string) (bool, error) {
	_, err := stdos.Stat(path)
	if err == nil {
		return true, nil
	}
	if stdos.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
