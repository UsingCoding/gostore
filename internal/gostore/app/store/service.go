package store

import (
	"context"
	stderrors "errors"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
)

type Service interface {
	Add(ctx context.Context, params AddParams) error

	Copy(ctx context.Context, params CopyParams) error
	Move(ctx context.Context, params MoveParams) error

	Get(ctx context.Context, params GetParams) ([]SecretData, error)
	List(ctx context.Context, params ListParams) (storage.Tree, error)

	Remove(ctx context.Context, params RemoveParams) error

	Unpack(ctx context.Context) error
	Pack(ctx context.Context, params PackParams) error

	Sync(ctx context.Context) error
	Rollback(ctx context.Context) error
}

func NewStoreService(
	storeID maybe.Maybe[string],
	storageManager storage.Manager,
	encryptionManager encryption.Manager,
	manifestSerializer ManifestSerializer,
	secretSerializer SecretSerializer,
	dataProvider DataProvider,
	identityProvider IdentityProvider,
) Service {
	return &storeService{
		storeID:            storeID,
		storageManager:     storageManager,
		encryptionManager:  encryptionManager,
		manifestSerializer: manifestSerializer,
		secretSerializer:   secretSerializer,
		dataProvider:       dataProvider,
		identityProvider:   identityProvider,
	}
}

type storeService struct {
	storeID maybe.Maybe[string]

	storageManager    storage.Manager
	encryptionManager encryption.Manager

	manifestSerializer ManifestSerializer
	secretSerializer   SecretSerializer

	dataProvider     DataProvider
	identityProvider IdentityProvider
}

func (service *storeService) Add(ctx context.Context, params AddParams) (err error) {
	s, err := service.loadStore(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}
	defer func() {
		err = stderrors.Join(err, s.close())
	}()

	err = s.add(
		ctx,
		params.Path,
		params.Key,
		params.Data,
	)
	return err
}

func (service *storeService) Copy(ctx context.Context, params CopyParams) (err error) {
	s, err := service.loadStore(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}
	defer func() {
		err = stderrors.Join(err, s.close())
	}()

	err = s.copy(ctx, params.Src, params.Dst)
	return err
}

func (service *storeService) Move(ctx context.Context, params MoveParams) (err error) {
	s, err := service.loadStore(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}
	defer func() {
		err = stderrors.Join(err, s.close())
	}()

	err = s.move(ctx, params.Src, params.Dst)
	return err
}

func (service *storeService) Get(ctx context.Context, params GetParams) ([]SecretData, error) {
	s, err := service.loadStore(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load store")
	}

	return s.get(ctx, params.Path, params.Key)
}

func (service *storeService) List(ctx context.Context, params ListParams) (storage.Tree, error) {
	s, err := service.loadStore(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load store")
	}

	return s.list(ctx, params.Path)
}

func (service *storeService) Remove(ctx context.Context, params RemoveParams) (err error) {
	s, err := service.loadStore(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}
	defer func() {
		err = stderrors.Join(err, s.close())
	}()

	err = s.remove(ctx, params.Path, params.Key)
	return err
}

func (service *storeService) Unpack(ctx context.Context) (err error) {
	s, err := service.loadStore(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}
	defer func() {
		err = stderrors.Join(err, s.close())
	}()

	err = s.unpack(ctx)
	if err == nil {
		s.manifest.Unpacked = true
		err = service.writeManifest(ctx, s.manifest, s.storage)
	}
	return err
}

func (service *storeService) Pack(ctx context.Context, params PackParams) (err error) {
	s, err := service.loadStore(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}
	defer func() {
		err = stderrors.Join(err, s.close())
	}()

	err = s.pack(ctx, params)
	if err == nil {
		s.manifest.Unpacked = false
		err = stderrors.Join(err, service.writeManifest(ctx, s.manifest, s.storage))
	}
	return err
}

func (service *storeService) Sync(ctx context.Context) error {
	s, err := service.loadStore(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}

	return s.sync(ctx)
}

func (service *storeService) Rollback(ctx context.Context) error {
	s, err := service.loadStore(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}

	return s.rollback(ctx)
}

func (service *storeService) loadStore(ctx context.Context) (*store, error) {
	storePath, err := service.resolveStoreLocation(ctx)
	if err != nil {
		return nil, err
	}

	s, err := service.storageManager.Use(ctx, storePath)
	if err != nil {
		return nil, err
	}

	manifestData, err := s.Get(ctx, ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manifest from storage")
	}

	if !maybe.Valid(manifestData) {
		return nil, errors.Wrap(err, "manifest not found in storage")
	}

	manifest, err := service.manifestSerializer.Deserialize(maybe.Just(manifestData))
	if err != nil {
		return nil, errors.Wrap(err, "failed to deserialize manifest")
	}

	encryptService, err := service.encryptionManager.EncryptService(manifest.Encryption)
	if err != nil {
		return nil, err
	}

	return &store{
		manifest:         manifest,
		storage:          s,
		encryption:       encryptService,
		secretSerializer: service.secretSerializer,
		identityProvider: service.identityProvider,
	}, nil
}

func (service *storeService) writeManifest(ctx context.Context, m Manifest, s storage.Storage) error {
	data, err := service.manifestSerializer.Serialize(m)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize store manifest")
	}

	err = s.Store(ctx, ManifestPath, data)
	return errors.Wrapf(err, "failed to store manifest")
}

func (service *storeService) resolveStoreLocation(ctx context.Context) (string, error) {
	if id, ok := maybe.JustValid(service.storeID); ok {
		storePath, err := service.dataProvider.StorePath(ctx, id)
		if err != nil {
			return "", err
		}
		if p, ok := maybe.JustValid(storePath); ok {
			return p, nil
		}

		return "", errors.Errorf("failed to resolve store location for %s", id)
	}

	storePath, err := service.dataProvider.CurrentStorePath(ctx)
	if err != nil {
		return "", err
	}

	if s, ok := maybe.JustValid(storePath); ok {
		return s, nil
	}

	return "", errors.New("failed to resolve store location: no current store path")
}
