package store

import (
	"context"

	"gostore/internal/common/maybe"
	"gostore/internal/gostore/app/encryption"
)

type IdentityProvider interface {
	IdentityByRecipient(ctx context.Context, recipient encryption.Recipient) (maybe.Maybe[encryption.Identity], error)
}
