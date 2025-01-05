package cli

import (
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/common/scope"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/app/storecrud"
	"github.com/UsingCoding/gostore/internal/gostore/app/usecase/totp"

	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	infraconfig "github.com/UsingCoding/gostore/internal/gostore/infrastructure/config"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/storage"
	infrastore "github.com/UsingCoding/gostore/internal/gostore/infrastructure/store"
)

func ContainerCtx(ctx *cli.Context) {
	ctx.Context = ContainerScope.Set(
		ctx.Context,
		NewContainer(ctx),
	)
}

var (
	ContainerScope scope.Typed[Container]
)

func NewContainer(ctx *cli.Context) Container {
	gostoreBaseDir := ctx.String("gostore-base-path")
	storeID := maybe.MapZero(ctx.String("store-id"))

	storageManager := storage.NewManager()
	encryptionManager := encryption.NewManager()

	c := config.NewService(
		infraconfig.NewStorage(gostoreBaseDir),
		gostoreBaseDir,
		storageManager,
		encryptionManager,
	)

	manifestSerializer := infrastore.NewManifestSerializer()

	storeService := store.NewStoreService(
		storeID,
		storageManager,
		encryptionManager,
		manifestSerializer,
		infrastore.NewSecretSerializer(),
		c,
		c,
	)

	storeCRUD := storecrud.NewService(
		c,
		encryptionManager,
		storageManager,
		manifestSerializer,
	)

	return Container{
		C:            c,
		StoreService: storeService,
		TOTP:         totp.NewService(storeService),
		StoreCRUD:    storeCRUD,
	}
}

type Container struct {
	C            config.Service
	StoreService store.Service

	StoreCRUD storecrud.Service
	TOTP      totp.Service
}
