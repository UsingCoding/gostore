package store

import (
	"fmt"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"strings"
)

type operations []string

func (s *operations) add(o string) {
	*s = append(*s, o)
}

func (s *operations) len() int {
	return len(*s)
}

func (s *operations) String() string {
	return strings.Join(*s, "; ")
}

type operation string

func addOperation(path string, key maybe.Maybe[string]) string {
	txt := "Add %s"
	args := []any{path}

	if k, ok := maybe.JustValid(key); ok {
		txt = "Add secret at %s to %s"
		args = []any{k, path}
	}

	return fmt.Sprintf(txt, args...)
}

func copyOperation(src, dst string) string {
	txt := "Copy %s to %s"
	args := []any{src, dst}

	return fmt.Sprintf(txt, args...)
}

func moveOperation(src, dst string) string {
	txt := "Move %s to %s"
	args := []any{src, dst}

	return fmt.Sprintf(txt, args...)
}

func removeOperation(path string, key maybe.Maybe[string]) string {
	txt := "Remove %s"
	args := []any{path}

	if k, ok := maybe.JustValid(key); ok {
		txt += fmt.Sprintf(" at %s", k)
	}

	return fmt.Sprintf(txt, args...)
}

func removeEmptyOperation(path string) string {
	txt := "Remove %s, since it's empty"
	args := []any{path}

	return fmt.Sprintf(txt, args...)
}

func packOperation() string {
	return "Pack store"
}
