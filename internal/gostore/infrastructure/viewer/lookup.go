package viewer

import (
	"os/exec"
	"runtime"

	"github.com/pkg/errors"
)

func appForView() (string, error) {
	var name string

	switch runtime.GOOS {
	case "linux":
		name = "xdg-open"
	default:
		return "", errors.Errorf("unknown app view for %s can be used", runtime.GOOS)
	}

	path := lookupAny(name)
	if path == "" {
		return "", errors.Errorf("cannot find apps for view, tried %s", path)
	}

	return path, nil
}
func lookupAny(editorNames ...string) string {
	for _, editorName := range editorNames {
		editorPath, err := exec.LookPath(editorName)
		if err == nil {
			return editorPath
		}
	}
	return ""
}
