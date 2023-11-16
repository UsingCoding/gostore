package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anmitsu/go-shlex"
	"github.com/pkg/errors"
)

func findEditor() ([]string, error) {
	editorExecutable := os.Getenv("EDITOR")

	if editorExecutable == "" {
		editorExecutable = lookupAnyEditor(fallbackEditors)
		if editorExecutable == "" {
			return nil, errors.Errorf(
				"no editor found in env $EDITOR and fallback variants %s",
				strings.Join(fallbackEditors, " "),
			)
		}
		return []string{editorExecutable}, nil
	}

	parts, err := shlex.Split(editorExecutable, false)
	if err != nil {
		return nil, fmt.Errorf("invalid $EDITOR: %s", editorExecutable)
	}

	return parts, nil
}

func lookupAnyEditor(editorNames []string) string {
	for _, editorName := range editorNames {
		editorPath, err := exec.LookPath(editorName)
		if err == nil {
			return editorPath
		}
	}
	return ""
}
