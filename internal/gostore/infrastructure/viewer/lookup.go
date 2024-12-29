package viewer

import (
	"os/exec"
	"runtime"

	"github.com/pkg/errors"
)

func appForView() (args []string, wait bool, err error) {
	switch runtime.GOOS {
	case "linux":
		args = []string{"xdg-open"}
	case "darwin":
		args = []string{"open", "-W", "-n"}
		wait = true
	default:
		return args, wait, errors.Errorf("unknown app view for %s can be used", runtime.GOOS)
	}

	path := lookupAny(args[0])
	if path == "" {
		return args, wait, errors.Errorf("cannot find apps for view, tried %s", path)
	}
	args[0] = path

	return args, wait, err
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
