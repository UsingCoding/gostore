package store

import (
	"context"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
)

type IdentityProvider interface {
	IdentityByRecipient(ctx context.Context, recipient encryption.Recipient) (maybe.Maybe[encryption.Identity], error)
}
