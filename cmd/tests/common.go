package tests

import (
	"io"
	"strings"

	"github.com/gofrs/uuid/v5"
)

func generateData() io.Reader {
	const c = 10
	res := make([]string, 0, c)
	for range 10 {
		res = append(
			res,
			uuid.Must(uuid.NewV7()).String(),
		)
	}
	return strings.NewReader(strings.Join(res, ":"))
}
