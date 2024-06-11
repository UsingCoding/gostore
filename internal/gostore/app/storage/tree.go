package storage

import (
	"github.com/UsingCoding/gostore/internal/common/orderedmap"
	"path/filepath"
)

type Tree []Entry

type EntryType string

const (
	FileEntryType    EntryType = "file"
	CatalogEntryType EntryType = "file"
)

type Entry struct {
	Name     string
	Type     EntryType
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
