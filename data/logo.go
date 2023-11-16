package data

import (
	_ "embed"
	"io"

	"github.com/fatih/color"
)

var (
	//go:embed logo.txt
	logo string
)

// Logo prints gopher logo
func Logo(w io.Writer) {
	_, _ = color.New(color.FgBlue).Fprintln(w, logo)
}
