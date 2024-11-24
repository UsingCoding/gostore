package storage

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// checks that path is relative and has no upper directories
func relativePathForStorage(p string) bool {
	return filepath.IsLocal(p)
}

func move(src, dst string) error {
	dstDir := path.Dir(dst)
	e, err := exists(dstDir)
	if err != nil {
		return err
	}

	if !e {
		err = os.MkdirAll(dstDir, 0o755)
		if err != nil {
			return errors.Wrapf(err, "failed to create dir for path %s", dst)
		}
	}

	return os.Rename(src, dst)
}

type sysStat struct {
	uid, gid uint32
}

func copyPath(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dstPath, _ := strings.CutPrefix(p, src)
		dstPath = path.Join(dst, dstPath)

		srcPath := p

		fileInfo, err := os.Stat(srcPath)
		if err != nil {
			return err
		}
		stat, err := castToSysStat(fileInfo)
		if err != nil {
			return errors.Wrap(err, "copyPath")
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err = createDirIfNotExists(dstPath, 0o755); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err = copySymLink(srcPath, dstPath); err != nil {
				return err
			}
		default:
			if err = copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}

		if err = os.Lchown(dstPath, int(stat.uid), int(stat.gid)); err != nil {
			return err
		}
		fInfo, err := d.Info()
		if err != nil {
			return err
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err = os.Chmod(dstPath, fInfo.Mode()); err != nil {
				return err
			}
		}

		return nil
	})
}

func copyFile(srcFile, dstFile string) error {
	if e, err := exists(path.Dir(dstFile)); !e || err != nil {
		if err != nil {
			return err
		}

		err = os.MkdirAll(path.Dir(dstFile), 0o755)
		if err != nil {
			return errors.Wrapf(err, "failed to create dir for path %s", dstFile)
		}
	}

	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer out.Close()
	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

func createDirIfNotExists(dir string, perm os.FileMode) error {
	if e, err := exists(dir); e || err != nil {
		return err
	}
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}
	return nil
}
func copySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
