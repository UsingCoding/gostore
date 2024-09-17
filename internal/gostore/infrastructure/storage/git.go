package storage

import (
	"context"
	"github.com/UsingCoding/gostore/internal/gostore/app/progress"
	"github.com/go-git/go-git/v5/plumbing/object"
	"io"
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

func (storage *gitStorage) Copy(_ context.Context, src, dst string) error {
	if !relativePathForStorage(src) {
		return errors.Errorf("path to secret is not local: %s", src)
	}
	if !relativePathForStorage(dst) {
		return errors.Errorf("path to secret is not local: %s", dst)
	}

	srcPath := path.Join(storage.repoDir, src)
	dstPath := path.Join(storage.repoDir, dst)

	err := copyPath(srcPath, dstPath)
	if err != nil {
		return errors.Wrapf(err, "failed to copy %s to %s", src, dst)
	}

	worktree, err := storage.repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = worktree.Add(dst)
	return errors.Wrapf(err, "failed to add to index copy %s to %s", src, dst)
}

func (storage *gitStorage) Move(_ context.Context, src, dst string) error {
	if !relativePathForStorage(src) {
		return errors.Errorf("path to secret is not local: %s", src)
	}
	if !relativePathForStorage(dst) {
		return errors.Errorf("path to secret is not local: %s", dst)
	}

	srcPath := path.Join(storage.repoDir, src)
	dstPath := path.Join(storage.repoDir, dst)

	err := move(srcPath, dstPath)
	if err != nil {
		return errors.Wrapf(err, "failed to move %s to %s", src, dst)
	}

	worktree, err := storage.repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = worktree.Add(src)
	if err != nil {
		return errors.Wrapf(err, "failed to commit changes in src %s", src)
	}

	_, err = worktree.Add(dst)
	if err != nil {
		return errors.Wrapf(err, "failed to commit changes in dst %s", dst)
	}

	return nil
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

func (storage *gitStorage) GetLatest(_ context.Context, p string) (maybe.Maybe[[]byte], error) {
	if !relativePathForStorage(p) {
		return maybe.NewNone[[]byte](), errors.Errorf("path to secret is not local: %s", p)
	}

	commit, err := storage.getLastCommit(maybe.NewJust(p))
	if err != nil {
		return maybe.Maybe[[]byte]{}, err
	}

	// new file or no commits in repo
	if commit == nil {
		return maybe.Maybe[[]byte]{}, nil
	}

	file, err := commit.File(p)
	if err != nil {
		return maybe.Maybe[[]byte]{}, errors.Wrap(err, "failed to get file from commit")
	}

	content, err := file.Contents()
	if err != nil {
		return maybe.Maybe[[]byte]{}, errors.Wrap(err, "failed to get file content from commit")
	}

	return maybe.NewJust([]byte(content)), nil
}

func (storage *gitStorage) List(_ context.Context, p string) (appstorage.Tree, error) {
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

func (storage *gitStorage) Push(ctx context.Context) error {
	err := storage.repo.PushContext(ctx, &git.PushOptions{
		RemoteName: remoteName,
		Auth:       nil,
		Progress:   defaultProgress(ctx),
	})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}
	return errors.Wrap(err, "failed to push repo")
}

func (storage *gitStorage) Pull(ctx context.Context) error {
	err := storage.repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: remoteName,
		Progress:   defaultProgress(ctx),
	})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}
		return err
	}

	worktree, err := storage.repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	err = worktree.PullContext(ctx, &git.PullOptions{
		RemoteName: remoteName,
		Progress:   defaultProgress(ctx),
	})
	return errors.Wrap(err, "failed to pull from repo")
}

func (storage *gitStorage) Commit(_ context.Context, msg string) error {
	worktree, err := storage.repo.Worktree()
	if err != nil {
		return errors.WithStack(err)
	}

	status, err := worktree.Status()
	if err != nil {
		return errors.Wrap(err, "failed to get worktree status")
	}

	if status.IsClean() {
		// nothing changed
		return nil
	}

	_, err = worktree.Commit(msg, &git.CommitOptions{
		All: true,
	})
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

			if len(children) == 0 {
				// skip empty dirs
				continue
			}
		}

		entries = append(entries, appstorage.Entry{
			Name:     entry.Name(),
			Children: children,
		})
	}

	return entries, nil
}

func (storage *gitStorage) getLastCommit(p maybe.Maybe[string]) (*object.Commit, error) {
	iter, err := storage.repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
		PathFilter: func(s string) bool {
			if !maybe.Valid(p) {
				return true
			}
			return s == maybe.Just(p)
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get last commit with associated file %s", p)
	}

	next, err := iter.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to iterate commits")
	}

	iter.Close()
	return next, nil
}

func defaultProgress(ctx context.Context) progress.Progress {
	return progress.FromCtx(ctx).Alter(
		progress.WithDescription("Packing store"),
		progress.WithBytes(true),
		progress.WithSpinnerType(11),
	)
}
