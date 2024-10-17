package tests

import (
	"fmt"
	"os"

	"github.com/gofrs/uuid/v5"

	"github.com/UsingCoding/gostore/cmd/tests/api"
)

func newSuite() (suite, error) {
	prefix := fmt.Sprintf("gostore-%s", uuid.Must(uuid.NewV7()))
	dir, err := os.MkdirTemp(
		"",
		prefix,
	)
	if err != nil {
		return suite{}, err
	}

	return suite{
		api:      api.New(dir),
		basePath: dir,
	}, nil
}

type suite struct {
	api api.API

	basePath string
}

func (s suite) gostore() api.API {
	return s.api
}

func (s suite) cleanup() {
	_ = os.RemoveAll(s.basePath)
}
