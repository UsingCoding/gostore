package store

import (
	"encoding/json"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/app/vars"
)

func NewManifestSerializer() store.ManifestSerializer {
	return &manifestSerializer{}
}

type manifestSerializer struct{}

func (serializer manifestSerializer) Serialize(m store.Manifest) ([]byte, error) {
	data, err := json.Marshal(manifest{
		Kind:        string(vars.StoreKind),
		StorageType: string(m.StorageType),
		Encryption:  string(m.Encryption),
		Recipients: slices.Map(m.Recipients, func(r encryption.Recipient) string {
			return string(r)
		}),
	})
	return data, errors.Wrap(err, "failed to serialize manifest")
}

func (serializer manifestSerializer) Deserialize(data []byte) (store.Manifest, error) {
	var m manifest
	err := json.Unmarshal(data, &m)
	if err != nil {
		return store.Manifest{}, errors.Wrap(err, "failed to deserialize manifest")
	}

	if m.Kind != string(vars.StoreKind) {
		return store.Manifest{}, errors.Errorf("unknown kind %s in manifest", m.Kind)
	}

	return store.Manifest{
		StorageType: storage.Type(m.StorageType),
		Encryption:  encryption.Encryption(m.Encryption),
		Recipients: slices.Map(m.Recipients, func(r string) encryption.Recipient {
			return encryption.Recipient(r)
		}),
	}, nil
}

type manifest struct {
	Kind        string   `json:"kind"`
	StorageType string   `json:"storageType"`
	Encryption  string   `json:"encryption"`
	Recipients  []string `json:"recipients"`
}
