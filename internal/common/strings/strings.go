package strings

import (
	"strings"
)

func HasPrefix(p string, entries []string) bool {
	for _, entry := range entries {
		if strings.HasPrefix(p, entry) {
			return true
		}
	}
	return false
}
