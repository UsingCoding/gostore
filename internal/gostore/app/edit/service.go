package edit

import (
	"context"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	appservice "github.com/UsingCoding/gostore/internal/gostore/app/service"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type Service interface {
	Edit(ctx context.Context, path string, key maybe.Maybe[string]) error
}

func NewService(s appservice.Service, editor Editor) Service {
	return &service{service: s, editor: editor}
}

type service struct {
	service appservice.Service
	editor  Editor
}

func (s *service) Edit(ctx context.Context, p string, key maybe.Maybe[string]) error {
	data, err := s.service.Get(ctx, store.GetParams{
		Path: p,
	})
	if err != nil {
		return err
	}

	var payload []byte
	if len(data) != 0 {
		payload, err = payloadFromData(data, p, key)
		if err != nil {
			return err
		}
	}

	edited, err := s.editor.Edit(ctx, p, payload)
	if err != nil {
		return err
	}

	return s.service.Add(ctx, store.AddParams{
		Path: p,
		Key:  key,
		Data: edited,
	})
}

func payloadFromData(data []store.SecretData, p string, key maybe.Maybe[string]) ([]byte, error) {
	if maybe.Valid(key) {
		for _, d := range data {
			if d.Name == maybe.Just(key) {
				return d.Payload, nil
			}
		}

		// allow to create new key in secret
		return nil, nil
	}

	for _, d := range data {
		if d.Default {
			return d.Payload, nil
		}
	}

	return nil, errors.Errorf("no default record found in %s", p)
}
