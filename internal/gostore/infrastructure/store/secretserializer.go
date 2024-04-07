package store

import (
	"encoding/json"
	"github.com/UsingCoding/gostore/internal/common/maybe"

	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/app/vars"
)

func NewSecretSerializer() store.SecretSerializer {
	return &secretSerializer{}
}

type secretSerializer struct{}

func (s secretSerializer) Serialize(sec store.Secret) ([]byte, error) {
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

func (s secretSerializer) Deserialize(b []byte) (store.Secret, error) {
	var sec secret
	decoder := json.Decoder{}
	decoder.DisallowUnknownFields()

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

func (s secretSerializer) RawSerialize(secret store.RawSecret) ([]byte, error) {
	if maybe.Valid(secret.Data) {
		return maybe.Just(secret.Data), nil
	}

	if maybe.Valid(secret.Payload) {
		return s.Serialize(store.Secret{
			Payload: maybe.Just(secret.Payload),
		})
	}

	return nil, errors.Errorf("unknown secret serialier format")
}

func (s secretSerializer) RawDeserialize(data []byte) (store.RawSecret, error) {
	res, err := s.Deserialize(data)
	if err == nil {
		return store.RawSecret{
			Payload: maybe.NewJust(res.Payload),
		}, nil
	}

	// failed to deserialize as JSON secret, interpret as raw data

	return store.RawSecret{
		Data: maybe.NewJust(data),
	}, nil
}

type secret struct {
	Kind    string            `json:"kind"`
	Payload map[string]string `json:"payload"` // convert to map[string]string to avoid unnecessary base64 conversion
}
