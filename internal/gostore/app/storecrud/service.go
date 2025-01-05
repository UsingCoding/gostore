package storecrud

import (
	"context"

	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type Service interface {
	Init(ctx context.Context, params InitParams) (InitRes, error)
	Clone(ctx context.Context, params CloneParams) error
}

func NewService(
	configService config.Service,
	encryptionManager encryption.Manager,
	storageManager storage.Manager,
	manifestSerializer store.ManifestSerializer,
) Service {
	return &service{
		configService:      configService,
		encryptionManager:  encryptionManager,
		storageManager:     storageManager,
		manifestSerializer: manifestSerializer,
	}
}

type service struct {
	configService config.Service

	encryptionManager  encryption.Manager
	storageManager     storage.Manager
	manifestSerializer store.ManifestSerializer
}
