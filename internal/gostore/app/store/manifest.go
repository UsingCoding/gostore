package store

import (
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
)

const (
	ManifestPath = ".gostore.json"
)

type Manifest struct {
	StorageType storage.Type
	Encryption  encryption.Encryption
	Recipients  []encryption.Recipient
}

type ManifestSerializer interface {
	Serialize(m Manifest) ([]byte, error)
	Deserialize(data []byte) (Manifest, error)
}
