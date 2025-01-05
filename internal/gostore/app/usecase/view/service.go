package view

import (
	"context"
	stderrors "errors"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

var (
	ErrSecretNotFound = stderrors.New("secret not found")
)

type Service interface {
	View(ctx context.Context, index store.SecretIndex) error
}

func NewService(s store.Service, viewer Viewer) Service {
	return &service{service: s, viewer: viewer}
}

type service struct {
	service store.Service
	viewer  Viewer
}

func (s *service) View(ctx context.Context, index store.SecretIndex) error {
	data, err := s.service.Get(ctx, store.GetParams{
		SecretIndex: index,
	})
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return errors.Wrap(ErrSecretNotFound, index.Path)
	}

	var secret maybe.Maybe[store.SecretData]
	//nolint:nestif
	if maybe.Valid(index.Key) {
		for _, d := range data {
			if d.Name == maybe.Just(index.Key) {
				secret = maybe.NewJust(d)
			}
		}
		if !maybe.Valid(secret) {
			return errors.Errorf("key %s not found in %s", maybe.Just(index.Key), index.Path)
		}
	} else {
		for _, d := range data {
			if d.Default {
				secret = maybe.NewJust(d)
			}
		}
		if !maybe.Valid(secret) {
			return errors.Errorf("no default record found in %s", index.Path)
		}
	}

	sec := maybe.Just(secret)
	pathForView := index.Path
	if !sec.Default {
		pathForView = index.Path + sec.Name
	}

	return s.viewer.View(ctx, pathForView, sec.Payload)
}
