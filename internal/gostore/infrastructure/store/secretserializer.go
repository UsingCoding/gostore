package store

import (
	"encoding/json"

	"github.com/pkg/errors"

	"gostore/internal/gostore/app/store"
	"gostore/internal/gostore/app/vars"
)

func NewSecretSerializer() store.SecretSerializer {
	return &secretSerializer{}
}

type secretSerializer struct{}

func (s *secretSerializer) Serialize(sec store.Secret) ([]byte, error) {
	p := map[string]string{}
	for k, v := range sec.Payload {
		p[k] = string(v)
	}

	data, err := json.Marshal(secret{
		Kind:    string(vars.SecretKind),
		Payload: p,
	})
	return data, errors.Wrap(err, "failed to serialize secret")
}

func (s *secretSerializer) Deserialize(b []byte) (store.Secret, error) {
	var sec secret
	err := json.Unmarshal(b, &sec)
	if err != nil {
		return store.Secret{}, errors.Wrap(err, "failed to deserialize secret")
	}

	if sec.Kind != string(vars.SecretKind) {
		return store.Secret{}, errors.Errorf("unknown kind %s in secret", sec.Kind)
	}

	p := map[string][]byte{}
	for k, v := range sec.Payload {
		p[k] = []byte(v)
	}

	return store.Secret{
		Payload: p,
	}, nil
}

type secret struct {
	Kind    string            `json:"kind"`
	Payload map[string]string `json:"payload"` // convert to map[string]string to avoid unnecessary base64 conversion
}
