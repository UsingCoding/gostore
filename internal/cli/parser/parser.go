package parser

import (
	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

func ParseSecretIndex(args []string) (store.SecretIndex, error) {
	if len(args) < 1 {
		return store.SecretIndex{}, errors.New("not enough arguments")
	}

	path := args[0]

	var key maybe.Maybe[string]
	if len(args) > 1 {
		key = maybe.NewJust(args[1])
	}

	return store.SecretIndex{
		Path: path,
		Key:  key,
	}, nil
}
