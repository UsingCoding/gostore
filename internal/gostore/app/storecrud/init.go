package storecrud

import (
	"context"
	"path"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/common/slices"
	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type InitParams struct {
	StoreID   string
	StorePath maybe.Maybe[string]

	// if there is no key passed new one will be created
	Recipients []encryption.Recipient

	StorageType maybe.Maybe[storage.Type]
	Encryption  maybe.Maybe[encryption.Encryption]
	Remote      maybe.Maybe[string]
}

type InitRes struct {
	StorePath string
	Identity  maybe.Maybe[encryption.Identity]
}

func (s service) Init(ctx context.Context, params InitParams) (InitRes, error) {
	err := s.configService.Init(ctx)
	if err != nil {
		return InitRes{}, err
	}

	err = s.ensureStoreNotExists(ctx, params.StoreID)
	if err != nil {
		return InitRes{}, err
	}

	storePath := maybe.MapNone(params.StorePath, func() string {
		return path.Join(s.configService.GostoreLocation(ctx), params.StoreID)
	})

	res := InitRes{
		StorePath: storePath,
	}

	if len(params.Recipients) == 0 {
		const enc = encryption.AgeEncryption

		identity, err2 := s.encryptionManager.GenerateIdentity(enc)
		if err2 != nil {
			return InitRes{}, err2
		}

		params.Recipients = []encryption.Recipient{identity.Recipient}

		res.Identity = maybe.NewJust(identity)
	}

	err = s.writeStore(ctx, storePath, params)
	if err != nil {
		return InitRes{}, err
	}

	err = s.configService.AddStore(
		ctx,
		config.StoreID(params.StoreID),
		storePath,
	)
	if err != nil {
		return InitRes{}, err
	}

	// write new identity to config
	if i, ok := maybe.JustValid(res.Identity); ok {
		err = s.configService.AddIdentity(ctx, i)
		if err != nil {
			return InitRes{}, errors.Wrapf(err, "failed to add identity for recipient %s to config", i.Recipient)
		}
	}

	return res, nil
}

func (s service) writeStore(
	ctx context.Context,
	storePath string,
	params InitParams,
) error {
	storageType := maybe.MapNone(params.StorageType, func() storage.Type {
		// use git as default storage type
		return storage.GITType
	})
	enc := maybe.MapNone(params.Encryption, func() encryption.Encryption {
		// use age as default encryption
		return encryption.AgeEncryption
	})

	storeStorage, err := s.storageManager.Init(
		ctx,
		storePath,
		params.Remote,
		storageType,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize storage for store")
	}

	// write store manifest
	m := store.Manifest{
		StorageType: storageType,
		Encryption:  enc,
		Recipients:  params.Recipients,
	}
	data, err := s.manifestSerializer.Serialize(m)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize store manifest")
	}

	err = storeStorage.Store(ctx, store.ManifestPath, data)
	if err != nil {
		return errors.Wrapf(err, "failed to store manifest")
	}

	err = storeStorage.Commit(ctx, "Initialized store")
	if err != nil {
		return errors.Wrapf(err, "failed to commit to storage")
	}

	return nil
}

func (s service) ensureStoreNotExists(ctx context.Context, storeID string) error {
	stores, err := s.configService.ListStores(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get stores")
	}

	_, exists := maybe.JustValid(slices.Find(stores, func(storeView config.StoreView) bool {
		return string(storeView.ID) == storeID
	}))
	if !exists {
		return nil
	}

	return errors.Errorf("store '%s' already exists", storeID)
}
