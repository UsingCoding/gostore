package totp

import (
	"context"
	"path"

	appservice "github.com/UsingCoding/gostore/internal/gostore/app/service"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type Service interface {
	AddIssuer(ctx context.Context, params AddParams) error
	Passcode(ctx context.Context, name string) (string, error)
}

func NewService(s appservice.Service) Service {
	return &service{service: s}
}

type service struct {
	service appservice.Service
}

const (
	totpPathPrefix = "totp"

	// totp metadata
	secretKey = "secret"
	algKey    = "alg"
)

func makeTOTPIndex(index store.SecretIndex) store.SecretIndex {
	// append prefix to path
	index.Path = path.Join(totpPathPrefix, index.Path)
	return index
}
