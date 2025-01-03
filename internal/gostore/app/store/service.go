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
	// Init creates new store
	Init(ctx context.Context, params InitParams) (InitRes, error)
	// Clone store from remote
	Clone(ctx context.Context, params CloneParams) error

	Add(ctx context.Context, params AddParams) error

	Copy(ctx context.Context, params CopyParams) error
	Move(ctx context.Context, params MoveParams) error

	Get(ctx context.Context, params GetParams) ([]SecretData, error)
	List(ctx context.Context, params ListParams) (storage.Tree, error)

	Remove(ctx context.Context, params RemoveParams) error

	Unpack(ctx context.Context, params CommonParams) error
	Pack(ctx context.Context, params CommonParams) error

	Sync(ctx context.Context, params SyncParams) error
	Rollback(ctx context.Context, params CommonParams) error
}

func NewStoreService(
	storageManager storage.Manager,
	encryptionManager encryption.Manager,
	manifestSerializer ManifestSerializer,
	secretSerializer SecretSerializer,
	identityProvider IdentityProvider,
) Service {
	return &storeService{
		storageManager:     storageManager,
		encryptionManager:  encryptionManager,
		manifestSerializer: manifestSerializer,
		secretSerializer:   secretSerializer,
		identityProvider:   identityProvider,
	}
}

type storeService struct {
	storageManager    storage.Manager
	encryptionManager encryption.Manager

	manifestSerializer ManifestSerializer
	secretSerializer   SecretSerializer

	identityProvider IdentityProvider
}

func (service *storeService) Init(ctx context.Context, params InitParams) (InitRes, error) {
	storePath, err := service.storeLocation(params.CommonParams)
	if err != nil {
		return InitRes{}, err
	}

	storageType := maybe.MapNone(params.StorageType, func() storage.Type {
		// use git as default storage type
		return storage.GITType
	})

	s, err := service.storageManager.Init(
		ctx,
		storePath,
		params.Remote,
		storageType,
	)
	if err != nil {
		return InitRes{}, err
	}

	var (
		recipients  []encryption.Recipient
		newIdentity maybe.Maybe[encryption.Identity]
	)
	const enc = encryption.AgeEncryption
	if len(params.Recipients) != 0 {
		recipients = params.Recipients
	} else {
		identity, err2 := service.encryptionManager.GenerateIdentity(enc)
		if err2 != nil {
			return InitRes{}, err2
		}

		recipients = []encryption.Recipient{identity.Recipient}

		newIdentity = maybe.NewJust(identity)
	}

	m := Manifest{
		StorageType: storageType,
		Encryption:  enc,
		Recipients:  recipients,
	}
	data, err := service.manifestSerializer.Serialize(m)
	if err != nil {
		return InitRes{}, errors.Wrapf(err, "failed to serialize store manifest")
	}

	err = s.Store(ctx, ManifestPath, data)
	if err != nil {
		return InitRes{}, errors.Wrapf(err, "failed to store manifest")
	}

	err = s.Commit(ctx, "Initialized store")
	if err != nil {
		return InitRes{}, errors.Wrapf(err, "failed to commit to storage")
	}

	if maybe.Valid(params.Remote) {
		err = s.Push(ctx)
		if err != nil {
			return InitRes{}, errors.Wrapf(err, "failed to sync storage")
		}
	}

	return InitRes{
		StorePath:         storePath,
		GeneratedIdentity: newIdentity,
	}, nil
}

func (service *storeService) Clone(ctx context.Context, params CloneParams) error {
	storePath, err := service.storeLocation(params.CommonParams)
	if err != nil {
		return err
	}

	_, err = service.storageManager.Clone(ctx, storePath, params.Remote, params.StorageType)
	return errors.Wrap(err, "failed to clone repo")
}

func (service *storeService) Add(ctx context.Context, params AddParams) (err error) {
	s, err := service.loadStore(ctx, params.CommonParams)
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
	s, err := service.loadStore(ctx, params.CommonParams)
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
	s, err := service.loadStore(ctx, params.CommonParams)
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
	s, err := service.loadStore(ctx, params.CommonParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load store")
	}

	return s.get(ctx, params.Path, params.Key)
}

func (service *storeService) List(ctx context.Context, params ListParams) (storage.Tree, error) {
	s, err := service.loadStore(ctx, params.CommonParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load store")
	}

	return s.list(ctx, params.Path)
}

func (service *storeService) Remove(ctx context.Context, params RemoveParams) (err error) {
	s, err := service.loadStore(ctx, params.CommonParams)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}
	defer func() {
		err = stderrors.Join(err, s.close())
	}()

	err = s.remove(ctx, params.Path, params.Key)
	return err
}

func (service *storeService) Unpack(ctx context.Context, params CommonParams) (err error) {
	s, err := service.loadStore(ctx, params)
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

func (service *storeService) Pack(ctx context.Context, params CommonParams) (err error) {
	s, err := service.loadStore(ctx, params)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}
	defer func() {
		err = stderrors.Join(err, s.close())
	}()

	err = s.pack(ctx)
	if err == nil {
		s.manifest.Unpacked = false
		err = stderrors.Join(err, service.writeManifest(ctx, s.manifest, s.storage))
	}
	return err
}

func (service *storeService) Sync(ctx context.Context, params SyncParams) error {
	s, err := service.loadStore(ctx, params.CommonParams)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}

	return s.sync(ctx)
}

func (service *storeService) Rollback(ctx context.Context, params CommonParams) error {
	s, err := service.loadStore(ctx, params)
	if err != nil {
		return errors.Wrap(err, "failed to load store")
	}

	return s.rollback(ctx)
}

func (service *storeService) loadStore(ctx context.Context, params CommonParams) (*store, error) {
	storePath, err := service.storeLocation(params)
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

func (service *storeService) storeLocation(params CommonParams) (string, error) {
	if maybe.Valid(params.StorePath) {
		return maybe.Just(params.StorePath), nil
	}

	return "", errors.New("empty store path")
}
