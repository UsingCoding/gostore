package storage

import (
	"path/filepath"
)

// checks that path is relative and has no upper directories
func relativePathForStorage(p string) bool {
	return filepath.IsLocal(p)
}
