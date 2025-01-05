package totp

import (
	"context"
	"path"

	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type Service interface {
	AddIssuer(ctx context.Context, params AddParams) error
	PasscodeView(ctx context.Context, name string) (PasscodeView, error)
}

func NewService(s store.Service) Service {
	return &service{service: s}
}

type service struct {
	service store.Service
}

const (
	totpPathPrefix = "totp"

	// totp metadata keys
	secretKey = "secret"
	algKey    = "alg"
)

func makeTOTPIndex(index store.SecretIndex) store.SecretIndex {
	// append prefix to path
	index.Path = path.Join(totpPathPrefix, index.Path)
	return index
}
