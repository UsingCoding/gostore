package cli

import (
	"github.com/urfave/cli/v2"

	"github.com/UsingCoding/gostore/internal/common/scope"
	"github.com/UsingCoding/gostore/internal/gostore/app/usecase/totp"

	"github.com/UsingCoding/gostore/internal/gostore/app/config"
	"github.com/UsingCoding/gostore/internal/gostore/app/service"
	infraconfig "github.com/UsingCoding/gostore/internal/gostore/infrastructure/config"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/storage"
	infrastore "github.com/UsingCoding/gostore/internal/gostore/infrastructure/store"
)

func ContainerCtx(ctx *cli.Context) {
	ctx.Context = ContainerScope.Set(ctx.Context, NewContainer(ctx))
}

func NewContainer(ctx *cli.Context) Container {
	gostoreBaseDir := ctx.String("gostore-base-path")

	storageManager := storage.NewManager()
	c := config.NewService(
		infraconfig.NewStorage(gostoreBaseDir),
		gostoreBaseDir,
		storageManager,
		encryption.NewManager(),
	)
	s := service.NewService(
		c,
		storageManager,
		encryption.NewManager(),
		infrastore.NewManifestSerializer(),
		infrastore.NewSecretSerializer(),
	)

	return Container{
		S:    s,
		C:    c,
		TOTP: totp.NewService(s),
	}
}

type Container struct {
	S service.Service
	C config.Service

	TOTP totp.Service
}

var (
	ContainerScope scope.Typed[Container]
)
