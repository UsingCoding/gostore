package config

import (
	"bytes"
	"context"
	stderrors "errors"
	stdslices "slices"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type StoreView struct {
	ID      StoreID
	Path    string
	Current bool
}

type Service interface {
	Init(ctx context.Context) error

	SetCurrentStore(ctx context.Context, storeID string) error

	CurrentStoreID(ctx context.Context) (maybe.Maybe[StoreID], error)
	CurrentStorePath(ctx context.Context) (maybe.Maybe[string], error)
	GostoreLocation(ctx context.Context) string

	ListStores(ctx context.Context) ([]StoreView, error)
	StoreByID(ctx context.Context, storeID StoreID) (maybe.Maybe[StoreView], error)

	AddIdentity(ctx context.Context, identities ...encryption.Identity) error
	AddStore(ctx context.Context, storeID StoreID, path string) error

	ImportRawIdentity(ctx context.Context, provider encryption.Provider, data []byte) error
	ExportRawIdentity(ctx context.Context, recipients ...encryption.Recipient) ([][]byte, error)

	RemoveStore(ctx context.Context, storeID StoreID) error

	store.IdentityProvider
}

func NewService(
	s Storage,
	gostoreLocation string,
	storageManager storage.Manager,
	encryptionManager encryption.Manager,
) Service {
	return &service{
		storage:           s,
		gostoreLocation:   gostoreLocation,
		storageManager:    storageManager,
		encryptionManager: encryptionManager,
	}
}

type service struct {
	storage         Storage
	gostoreLocation string

	storageManager    storage.Manager
	encryptionManager encryption.Manager
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

	i := stdslices.IndexFunc(config.Stores, func(s Store) bool {
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

	i := stdslices.IndexFunc(config.Stores, func(s Store) bool {
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

func (s *service) ListStores(ctx context.Context) ([]StoreView, error) {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	return slices.Map(config.Stores, func(s Store) StoreView {
		return StoreView{
			ID:      s.ID,
			Path:    s.Path,
			Current: maybe.Valid(config.Context) && maybe.Just(config.Context) == s.ID,
		}
	}), nil
}

func (s *service) StoreByID(ctx context.Context, storeID StoreID) (maybe.Maybe[StoreView], error) {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return maybe.Maybe[StoreView]{}, err
	}

	i := stdslices.IndexFunc(config.Stores, func(s Store) bool {
		return s.ID == storeID
	})
	if i == -1 {
		return maybe.Maybe[StoreView]{}, err
	}

	foundStore := config.Stores[i]

	return maybe.NewJust(StoreView{
		ID:      foundStore.ID,
		Path:    foundStore.Path,
		Current: maybe.Valid(config.Context) && maybe.Just(config.Context) == foundStore.ID,
	}), nil
}

func (s *service) AddIdentity(ctx context.Context, identities ...encryption.Identity) error {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	for _, identity := range identities {
		i := stdslices.IndexFunc(config.Identities, func(i encryption.Identity) bool {
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

	i := stdslices.IndexFunc(config.Stores, func(s Store) bool {
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

func (s *service) ImportRawIdentity(ctx context.Context, provider encryption.Provider, data []byte) error {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	identity, err := s.encryptionManager.ImportRawIdentity(provider, data)
	if err != nil {
		return err
	}

	i := stdslices.IndexFunc(config.Identities, func(i encryption.Identity) bool {
		return bytes.Equal(i.Recipient, identity.Recipient)
	})

	if i != -1 {
		return errors.Errorf("identity with recipient %s alrteady added", identity.Recipient)
	}
	config.Identities = append(config.Identities, identity)

	return s.storage.Store(ctx, config)
}

func (s *service) ExportRawIdentity(ctx context.Context, recipients ...encryption.Recipient) ([][]byte, error) {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	return slices.MapErr(recipients, func(recipient encryption.Recipient) ([]byte, error) {
		i := stdslices.IndexFunc(config.Identities, func(i encryption.Identity) bool {
			return bytes.Equal(i.Recipient, recipient)
		})

		if i == -1 {
			return nil, errors.Errorf("identity for %s not found", recipient)
		}

		identity := config.Identities[i]
		data, err2 := s.encryptionManager.ExportRawIdentity(identity)
		if err2 != nil {
			return nil, errors.Wrapf(err2, "failed to export raw identity for %s", identity.Recipient)
		}

		return data, nil
	})
}

func (s *service) RemoveStore(ctx context.Context, storeID StoreID) error {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	i := stdslices.IndexFunc(config.Stores, func(s Store) bool {
		return s.ID == storeID
	})

	if i == -1 {
		return errors.Errorf("store %s not found to remove", storeID)
	}

	storeToRemove := config.Stores[i]

	err = s.storageManager.Remove(ctx, storeToRemove.Path)
	if err != nil {
		return errors.Wrap(err, "failed to remove store")
	}

	config.Stores = append(config.Stores[:i], config.Stores[i+1:]...)

	if maybe.Valid(config.Context) && maybe.Just(config.Context) == storeID {
		// reset current context
		config.Context = maybe.NewNone[StoreID]()

		// set as current store first existed store
		if len(config.Stores) > 0 {
			config.Context = maybe.NewJust(config.Stores[0].ID)
		}
	}

	err = s.storage.Store(ctx, config)
	return err
}

func (s *service) IdentityByRecipient(ctx context.Context, recipient encryption.Recipient) (maybe.Maybe[encryption.Identity], error) {
	config, err := s.storage.Load(ctx)
	if err != nil {
		return maybe.Maybe[encryption.Identity]{}, errors.Wrap(err, "failed to load config")
	}

	i := stdslices.IndexFunc(config.Identities, func(i encryption.Identity) bool {
		return bytes.Equal(recipient, i.Recipient)
	})

	if i == -1 {
		return maybe.Maybe[encryption.Identity]{}, nil
	}

	return maybe.NewJust(config.Identities[i]), nil
}
