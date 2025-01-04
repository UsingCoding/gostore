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

type InitParams struct {
	CommonParams

	// if there is no key passed new one will be created
	Recipients []encryption.Recipient

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
	CommonParams
	SecretIndex

	Data []byte
}

type CopyParams struct {
	CommonParams

	Src string
	Dst string
}

type MoveParams struct {
	CommonParams

	Src string
	Dst string
}

type GetParams struct {
	CommonParams
	SecretIndex
}

type ListParams struct {
	CommonParams

	Path string
}

type RemoveParams struct {
	CommonParams

	Path string
	Key  maybe.Maybe[string]
}

type SyncParams struct {
	CommonParams
}
