package store

import (
	"slices"

	"github.com/UsingCoding/gostore/internal/common/maybe"
)

const (
	defaultKey = "data"
)

type Secret struct {
	// Payload is json object with encrypted values
	Payload map[string][]byte
}

func (s *Secret) addData(key maybe.Maybe[string], data []byte) {
	k := maybe.MapNone(key, func() string {
		return defaultKey
	})

	s.Payload[k] = data
}

func (s *Secret) get(key maybe.Maybe[string]) []SecretData {
	if maybe.Valid(key) {
		k := maybe.Just(key)
		data, exists := s.Payload[k]
		if !exists {
			return nil
		}

		return []SecretData{{
			Name:    k,
			Payload: data,
			Default: k == defaultKey,
		}}
	}

	keys := make([]string, 0, len(s.Payload))
	for name := range s.Payload {
		keys = append(keys, name)
	}

	slices.Sort(keys)

	var secrets []SecretData
	for _, k := range keys {
		secrets = append(secrets, SecretData{
			Name:    k,
			Payload: s.Payload[k],
			Default: k == defaultKey,
		})
	}

	return secrets
}

func (s *Secret) remove(key string) {
	delete(s.Payload, key)
}

func (s *Secret) empty() bool {
	return len(s.Payload) == 0
}

func initSecret() Secret {
	return Secret{
		Payload: map[string][]byte{},
	}
}

type SecretData struct {
	Name    string
	Payload []byte

	Default bool // means that secret data stored in default field
}

// RawSecret - raw secret one of
type RawSecret struct {
	Data    maybe.Maybe[[]byte]
	Payload maybe.Maybe[map[string][]byte]
}

type SecretSerializer interface {
	Serialize(secret Secret) ([]byte, error)
	Deserialize([]byte) (Secret, error)

	RawSerialize(secret RawSecret) ([]byte, error)
	RawDeserialize(data []byte) (RawSecret, error)
}
