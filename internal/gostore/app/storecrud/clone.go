package storecrud

import (
	"context"
	"path"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
)

type CloneParams struct {
	StoreID   string
	StorePath maybe.Maybe[string]

	StorageType storage.Type
	Remote      string
}

func (s service) Clone(ctx context.Context, params CloneParams) error {
	err := s.configService.Init(ctx)
	if err != nil {
		return err
	}

	err = s.ensureStoreNotExists(ctx, params.StoreID)
	if err != nil {
		return err
	}

	storePath := maybe.MapNone(params.StorePath, func() string {
		return path.Join(s.configService.GostoreLocation(ctx), params.StoreID)
	})

	_, err = s.storageManager.Clone(
		ctx,
		storePath,
		params.Remote,
		params.StorageType,
	)
	if err != nil {
		return errors.Wrap(err, "failed to clone repo")
	}

	err = s.configService.AddStore(ctx, config.StoreID(params.StoreID), storePath)
	return err
}
