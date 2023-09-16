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

	storePath := maybe.MapNone(params.StorePath, func() string {
		return path.Join(s.configService.GostoreLocation(ctx), params.ID)
	})

	params.StorePath = maybe.NewJust(storePath)

	initRes, err := s.makeStoreService().Init(ctx, params)
	if err != nil {
		return store.InitRes{}, err
	}

	err = s.configService.AddStore(ctx, config.StoreID(params.ID), initRes.StorePath)
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
	p, err := s.populateCommonParams(ctx, params.CommonParams)
	if err != nil {
		return err
	}

	params.CommonParams = p

	return s.makeStoreService().
		Clone(ctx, params)
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

func (s *service) populateCommonParams(ctx context.Context, params store.CommonParams) (store.CommonParams, error) {
	if !maybe.Valid(params.StorePath) {
		storePath, err := s.configService.CurrentStorePath(ctx)
		if err != nil {
			return store.CommonParams{}, errors.Wrap(err, "failed to get current store")
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
