package store

import (
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
)

type CommonParams struct {
	StorePath maybe.Maybe[string]
	StoreID   maybe.Maybe[string]
}

type SecretIndex struct {
	Path string
	Key  maybe.Maybe[string]
}

func (q SecretIndex) String() string {
	s := q.Path
	if k, ok := maybe.JustValid(q.Key); ok {
		s += "->" + k
	}
	return s
}

type WriteManifestParams struct {
	Recipients  []encryption.Recipient
	Encryption  encryption.Encryption
	StorageType storage.Type
}

type InitParams struct {
	CommonParams

	Recipients []encryption.Recipient
	Encryption encryption.Encryption

	StorageType maybe.Maybe[storage.Type]
	Remote      maybe.Maybe[string]
}

type CloneParams struct {
	CommonParams

	ID string

	StorageType storage.Type
	Remote      string
}

type InitRes struct {
	StorePath         string
	GeneratedIdentity maybe.Maybe[encryption.Identity]
}

type AddParams struct {
	SecretIndex

	Data []byte
}

type CopyParams struct {
	Src string
	Dst string
}

type MoveParams struct {
	Src string
	Dst string
}

type GetParams struct {
	SecretIndex
}

type ListParams struct {
	Path string
}

type RemoveParams struct {
	Path string
	Key  maybe.Maybe[string]
}
