package storage

import (
	"context"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	commonstrings "github.com/UsingCoding/gostore/internal/common/strings"
	appstorage "github.com/UsingCoding/gostore/internal/gostore/app/storage"
)

var (
	storagePaths = []string{
		".git",
	}
)

type gitStorage struct {
	repo    *git.Repository
	repoDir string
}

func (storage *gitStorage) Store(_ context.Context, p string, data []byte) error {
	if !relativePathForStorage(p) {
		return errors.Errorf("path to secret is not local: %s", p)
	}

	fullPath := path.Join(storage.repoDir, p)

	// check existence of parent directory
	if e, err2 := exists(path.Dir(fullPath)); !e || err2 != nil {
		if err2 != nil {
			return errors.WithStack(err2)
		}

		err2 = os.MkdirAll(path.Dir(fullPath), os.ModePerm)
		if err2 != nil {
			return errors.Wrapf(err2, "failed to create subdirectories for path %s", p)
		}
	}

	err := os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write file to %s", fullPath)
	}

	worktree, err := storage.repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = worktree.Add(p)
	if err != nil {
		return errors.Wrapf(err, "failed to add file to git index: %s", fullPath)
	}

	return nil
}

func (storage *gitStorage) Remove(_ context.Context, p string) error {
	if !relativePathForStorage(p) {
		return errors.Errorf("path to secret is not local: %s", p)
	}

	fullPath := path.Join(storage.repoDir, p)

	if e, err := exists(fullPath); !e || err != nil {
		return errors.Wrapf(err, "failed to find a path in repo %s", p)
	}

	worktree, err := storage.repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = worktree.Remove(p)
	return errors.Wrapf(err, "failed to remove path from git index: %s", fullPath)
}

func (storage *gitStorage) Get(_ context.Context, p string) (maybe.Maybe[[]byte], error) {
	if !relativePathForStorage(p) {
		return maybe.NewNone[[]byte](), errors.Errorf("path to secret is not local: %s", p)
	}

	fullPath := path.Join(storage.repoDir, p)

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return maybe.NewNone[[]byte](), nil
		}

		return maybe.NewNone[[]byte](), errors.Wrapf(err, "failed to find a path in repo %s", p)
	}

	if stat.IsDir() {
		return maybe.NewNone[[]byte](), errors.Errorf("find directory from storage, not a file: %s", p)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return maybe.Maybe[[]byte]{}, errors.WithStack(err)
	}

	return maybe.NewJust(data), nil
}

func (storage *gitStorage) List(_ context.Context, p string) ([]appstorage.Entry, error) {
	fixedPath := storage.repoDir
	if p != "" {
		if !relativePathForStorage(p) {
			return nil, errors.Errorf("path to list is not local: %s", p)
		}

		fixedPath = path.Join(storage.repoDir, p)
	}

	stat, err := os.Stat(fixedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to find a path in repo %s", p)
	}

	if !stat.IsDir() {
		return nil, nil
	}

	entries, err := storage.listEntriesRecursively(fixedPath)
	return entries, errors.Wrap(err, "failed to list storage entries")
}

func (storage *gitStorage) AddRemote(_ context.Context, remoteName string, remoteAddr string) error {
	_, err := storage.repo.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{remoteAddr},
	})
	return errors.Wrapf(err, "failed to add remote %s %s to repo", remoteName, remoteAddr)
}

func (storage *gitStorage) HasRemote(context.Context) (bool, error) {
	remotes, err := storage.repo.Remotes()
	if err != nil {
		return false, errors.Wrap(err, "failed to get remotes from repo")
	}

	return len(remotes) != 0, nil
}

func (storage *gitStorage) Sync(ctx context.Context) error {
	err := storage.repo.PushContext(ctx, &git.PushOptions{
		RemoteName: remoteName,
		Auth:       nil,
	})
	return errors.Wrap(err, "failed to sync repo")
}

func (storage *gitStorage) Commit(_ context.Context, msg string) error {
	worktree, err := storage.repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = worktree.Commit(msg, &git.CommitOptions{})
	return errors.Wrap(err, "failed to commit changes")
}

func (storage *gitStorage) Rollback(_ context.Context) error {
	worktree, err := storage.repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	err = worktree.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
	return errors.Wrap(err, "failed to rollback storage")
}

func (storage *gitStorage) listEntriesRecursively(p string) ([]appstorage.Entry, error) {
	dirEntries, err := os.ReadDir(p)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read dir %s", p)
	}

	var entries []appstorage.Entry

	for _, entry := range dirEntries {
		var children []appstorage.Entry

		if commonstrings.HasPrefix(entry.Name(), storagePaths) {
			continue
		}

		if entry.IsDir() {
			children, err = storage.listEntriesRecursively(path.Join(p, entry.Name()))
			if err != nil {
				return nil, err
			}
		}

		entries = append(entries, appstorage.Entry{
			Name:     entry.Name(),
			Children: children,
		})
	}

	return entries, nil
}
