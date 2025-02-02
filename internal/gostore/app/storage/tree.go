package storage

import (
	"path/filepath"

	"github.com/UsingCoding/gostore/internal/common/orderedmap"
)

type Tree []Entry

type Entry struct {
	Name     string
	Children []Entry
}

func (tree Tree) Inline() *orderedmap.Map[string, Entry] {
	res := orderedmap.NewStable[string, Entry]()

	const root = ""
	collectPaths(res, root, tree)

	return res
}

func collectPaths(res *orderedmap.Map[string, Entry], base string, entries []Entry) {
	for _, entry := range entries {
		p := filepath.Join(base, entry.Name)
		if len(entry.Children) == 0 {
			res.Add(p, entry)
			continue
		}

		collectPaths(
			res,
			p,
			entry.Children,
		)
	}
}
