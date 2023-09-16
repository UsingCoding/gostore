package config

import (
	"bytes"
	"context"
	stderrors "errors"
	"slices"

	"github.com/pkg/errors"

	"gostore/internal/common/maybe"
	"gostore/internal/gostore/app/encryption"
	"gostore/internal/gostore/app/store"
)

type Service interface {
	Init(ctx context.Context) error

	SetCurrentStore(ctx context.Context, storeID string) error

	CurrentStoreID(ctx context.Context) (maybe.Maybe[StoreID], error)
	CurrentStorePath(ctx context.Context) (maybe.Maybe[string], error)
	GostoreLocation(ctx context.Context) string
	ListStores(ctx context.Context) ([]Store, error)

	AddIdentity(ctx context.Context, identities ...encryption.Identity) error
	AddStore(ctx context.Context, storeID StoreID, path string) error

	store.IdentityProvider
}

func NewService(storage Storage, gostoreLocation string) Service {
	return &service{storage: storage, gostoreLocation: gostoreLocation}
}

type service struct {
	storage         Storage
	gostoreLocation string
}

func (s *service) Init(ctx context.Context) error {
	_, err := s.storage.Load(ctx)
	if err == nil || !stderrors.Is(err, ErrConfigNotFound) {
		return err
	}

	err = s.storage.Store(ctx, Config{
		Context:    maybe.Maybe[StoreID]{},
		Stores:     nil,
		Identities: nil,
	})
	if err != nil {
		return errors.Wrap(err, "failed to store default config")
	}

	return err
}

func (s *service) SetCurrentStore(ctx context.Context, storeID string) error {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	i := slices.IndexFunc(config.Stores, func(s Store) bool {
		return string(s.ID) == storeID
	})
	if i == -1 {
		return errors.Errorf("store with id %s not found", storeID)
	}

	id := config.Stores[i].ID
	config.Context = maybe.NewJust(id)

	return s.storage.Store(ctx, config)
}

func (s *service) CurrentStoreID(ctx context.Context) (maybe.Maybe[StoreID], error) {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return maybe.Maybe[StoreID]{}, errors.Wrap(err, "failed to load config")
	}

	return config.Context, nil
}

func (s *service) CurrentStorePath(ctx context.Context) (maybe.Maybe[string], error) {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return maybe.Maybe[string]{}, errors.Wrap(err, "failed to load config")
	}

	if !maybe.Valid(config.Context) {
		return maybe.Maybe[string]{}, nil
	}

	i := slices.IndexFunc(config.Stores, func(s Store) bool {
		return s.ID == maybe.Just(config.Context)
	})
	if i == -1 {
		return maybe.Maybe[string]{}, err
	}

	return maybe.NewJust(config.Stores[i].Path), nil
}

func (s *service) GostoreLocation(context.Context) string {
	return s.gostoreLocation
}

func (s *service) ListStores(ctx context.Context) ([]Store, error) {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	return config.Stores, nil
}

func (s *service) AddIdentity(ctx context.Context, identities ...encryption.Identity) error {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	for _, identity := range identities {
		i := slices.IndexFunc(config.Identities, func(i encryption.Identity) bool {
			return bytes.Equal(identity.Recipient, i.Recipient)
		})

		if i != -1 {
			return errors.Errorf("identity with recipient %s already exists", identity.Recipient)
		}
	}

	config.Identities = append(config.Identities, identities...)

	err = s.storage.Store(ctx, config)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) AddStore(ctx context.Context, storeID StoreID, path string) error {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	i := slices.IndexFunc(config.Stores, func(s Store) bool {
		return s.ID == storeID
	})

	if i != -1 {
		return nil
	}

	config.Stores = append(config.Stores, Store{
		ID:   storeID,
		Path: path,
	})

	// if there is no store in context set new store to context
	if !maybe.Valid(config.Context) {
		config.Context = maybe.NewJust(storeID)
	}

	err = s.storage.Store(ctx, config)
	return err
}

func (s *service) IdentityByRecipient(ctx context.Context, recipient encryption.Recipient) (maybe.Maybe[encryption.Identity], error) {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return maybe.Maybe[encryption.Identity]{}, errors.Wrap(err, "failed to load config")
	}

	i := slices.IndexFunc(config.Identities, func(i encryption.Identity) bool {
		return bytes.Equal(recipient, i.Recipient)
	})

	if i == -1 {
		return maybe.Maybe[encryption.Identity]{}, nil
	}

	return maybe.NewJust(config.Identities[i]), nil
}
