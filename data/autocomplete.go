package data

import (
	_ "embed"
	"fmt"
)

var (
	//go:embed bash_autocomplete
	bash string
	//go:embed zsh_autocomplete
	zsh string
)

func Bash(id string) string {
	return fmt.Sprintf(bash, id)
}

func Zsh(id string) string {
	return fmt.Sprintf(zsh, id)
}
