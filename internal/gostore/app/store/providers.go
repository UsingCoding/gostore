package store

import (
	"context"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
)

type DataProvider interface {
	CurrentStorePath(ctx context.Context) (maybe.Maybe[string], error)
	StorePath(ctx context.Context, storeID string) (maybe.Maybe[string], error)
}

type IdentityProvider interface {
	IdentityByRecipient(ctx context.Context, recipient encryption.Recipient) (maybe.Maybe[encryption.Identity], error)
}
