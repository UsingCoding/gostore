package view

import (
	"context"
	stderrors "errors"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	appservice "github.com/UsingCoding/gostore/internal/gostore/app/service"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

var (
	ErrSecretNotFound = stderrors.New("secret not found")
)

type Service interface {
	View(ctx context.Context, p string, key maybe.Maybe[string]) error
}

func NewService(s appservice.Service, viewer Viewer) Service {
	return &service{service: s, viewer: viewer}
}

type service struct {
	service appservice.Service
	viewer  Viewer
}

func (s *service) View(ctx context.Context, p string, key maybe.Maybe[string]) error {
	data, err := s.service.Get(ctx, store.GetParams{
		Path: p,
	})
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return errors.Wrap(ErrSecretNotFound, p)
	}

	var secret maybe.Maybe[store.SecretData]
	//nolint:nestif
	if maybe.Valid(key) {
		for _, d := range data {
			if d.Name == maybe.Just(key) {
				secret = maybe.NewJust(d)
			}
		}
		if !maybe.Valid(secret) {
			return errors.Errorf("key %s not found in %s", maybe.Just(key), p)
		}
	} else {
		for _, d := range data {
			if d.Default {
				secret = maybe.NewJust(d)
			}
		}
		if !maybe.Valid(secret) {
			return errors.Errorf("no default record found in %s", p)
		}
	}

	sec := maybe.Just(secret)
	pathForView := p
	if !sec.Default {
		pathForView = p + sec.Name
	}

	return s.viewer.View(ctx, pathForView, sec.Payload)
}
