package service

import (
	"context"
	"path"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type Service interface {
	store.Service
}

func NewService(
	configService config.Service,
	storageManager storage.Manager,
	encryptionManager encryption.Manager,
	manifestSerializer store.ManifestSerializer,
	secretSerializer store.SecretSerializer,
) Service {
	return &service{
		configService:      configService,
		storageManager:     storageManager,
		encryptionManager:  encryptionManager,
		manifestSerializer: manifestSerializer,
		secretSerializer:   secretSerializer,
	}
}

type service struct {
	configService config.Service

	storageManager     storage.Manager
	encryptionManager  encryption.Manager
	manifestSerializer store.ManifestSerializer
	secretSerializer   store.SecretSerializer
}

func (s *service) Init(ctx context.Context, params store.InitParams) (store.InitRes, error) {
	err := s.configService.Init(ctx)
	if err != nil {
		return store.InitRes{}, err
	}

	storeID, ok := maybe.JustValid(params.StoreID)
	if !ok {
		return store.InitRes{}, errors.New("storeID not passed")
	}

	storePath := maybe.MapNone(params.StorePath, func() string {
		return path.Join(s.configService.GostoreLocation(ctx), storeID)
	})

	params.StorePath = maybe.NewJust(storePath)

	initRes, err := s.makeStoreService().Init(ctx, params)
	if err != nil {
		return store.InitRes{}, err
	}

	err = s.configService.AddStore(ctx, config.StoreID(storeID), initRes.StorePath)
	if err != nil {
		return store.InitRes{}, err
	}

	if maybe.Valid(initRes.GeneratedIdentity) {
		identity := maybe.Just(initRes.GeneratedIdentity)
		err = s.configService.AddIdentity(ctx, identity)
		if err != nil {
			return store.InitRes{}, errors.Wrapf(err, "failed to add identity for recipient %s to config", identity.Recipient)
		}
	}

	return initRes, err
}

func (s *service) Clone(ctx context.Context, params store.CloneParams) error {
	err := s.configService.Init(ctx)
	if err != nil {
		return err
	}

	storePath := maybe.MapNone(params.StorePath, func() string {
		return path.Join(s.configService.GostoreLocation(ctx), params.ID)
	})

	params.StorePath = maybe.NewJust(storePath)

	err = s.makeStoreService().Clone(ctx, params)
	if err != nil {
		return err
	}

	err = s.configService.AddStore(ctx, config.StoreID(params.ID), storePath)
	return err
}

func (s *service) Add(ctx context.Context, params store.AddParams) error {
	p, err := s.populateCommonParams(ctx, params.CommonParams)
	if err != nil {
		return err
	}

	params.CommonParams = p

	return s.makeStoreService().
		Add(ctx, params)
}

func (s *service) Copy(ctx context.Context, params store.CopyParams) error {
	p, err := s.populateCommonParams(ctx, params.CommonParams)
	if err != nil {
		return err
	}

	params.CommonParams = p

	return s.makeStoreService().
		Copy(ctx, params)
}

func (s *service) Move(ctx context.Context, params store.MoveParams) error {
	p, err := s.populateCommonParams(ctx, params.CommonParams)
	if err != nil {
		return err
	}

	params.CommonParams = p

	return s.makeStoreService().
		Move(ctx, params)
}

func (s *service) Get(ctx context.Context, params store.GetParams) ([]store.SecretData, error) {
	p, err := s.populateCommonParams(ctx, params.CommonParams)
	if err != nil {
		return nil, err
	}

	params.CommonParams = p

	return s.makeStoreService().
		Get(ctx, params)
}

func (s *service) List(ctx context.Context, params store.ListParams) ([]storage.Entry, error) {
	p, err := s.populateCommonParams(ctx, params.CommonParams)
	if err != nil {
		return nil, err
	}

	params.CommonParams = p

	return s.makeStoreService().
		List(ctx, params)
}

func (s *service) Remove(ctx context.Context, params store.RemoveParams) error {
	p, err := s.populateCommonParams(ctx, params.CommonParams)
	if err != nil {
		return err
	}

	params.CommonParams = p

	return s.makeStoreService().
		Remove(ctx, params)
}

func (s *service) Unpack(ctx context.Context, params store.CommonParams) error {
	p, err := s.populateCommonParams(ctx, params)
	if err != nil {
		return err
	}

	return s.makeStoreService().
		Unpack(ctx, p)
}

func (s *service) Pack(ctx context.Context, params store.CommonParams) error {
	p, err := s.populateCommonParams(ctx, params)
	if err != nil {
		return err
	}

	return s.makeStoreService().
		Pack(ctx, p)
}

func (s *service) Sync(ctx context.Context, params store.SyncParams) error {
	p, err := s.populateCommonParams(ctx, params.CommonParams)
	if err != nil {
		return err
	}

	params.CommonParams = p

	return s.makeStoreService().
		Sync(ctx, params)
}

func (s *service) Rollback(ctx context.Context, params store.CommonParams) error {
	p, err := s.populateCommonParams(ctx, params)
	if err != nil {
		return err
	}

	return s.makeStoreService().
		Rollback(ctx, p)
}

func (s *service) populateCommonParams(ctx context.Context, params store.CommonParams) (store.CommonParams, error) {
	if !maybe.Valid(params.StorePath) {
		var storePath maybe.Maybe[string]
		if id, ok := maybe.JustValid(params.StoreID); ok {
			storeByID, err := s.configService.StoreByID(ctx, config.StoreID(id))
			if err != nil {
				return store.CommonParams{}, err
			}

			foundStore, ok := maybe.JustValid(storeByID)
			if !ok {
				return store.CommonParams{}, errors.Errorf("store by id %s is not found", id)
			}
			storePath = maybe.NewJust(foundStore.Path)
		} else {
			var err error
			storePath, err = s.configService.CurrentStorePath(ctx)
			if err != nil {
				return store.CommonParams{}, errors.Wrap(err, "failed to get current store")
			}
		}

		params.StorePath = storePath
	}

	return params, nil
}

func (s *service) makeStoreService() store.Service {
	return store.NewStoreService(
		s.storageManager,
		s.encryptionManager,
		s.manifestSerializer,
		s.secretSerializer,
		s.configService,
	)
}
