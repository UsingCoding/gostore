package store

import (
	"context"
	stderrors "errors"
	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/pkg/errors"
	"path"
)

func (s *store) unpack(ctx context.Context) (err error) {
	defer func() {
		if err == nil {
			return
		}

		// use background ctx to rollback changes os closed ctx
		err = stderrors.Join(err, s.rollback(context.Background()))
	}()

	const root = ""
	entries, err := s.list(ctx, root)
	if err != nil {
		return err
	}

	for _, entryPath := range inlinePaths(entries) {
		secretData, err2 := s.get(ctx, entryPath, maybe.NewNone[string]())
		if err2 != nil {
			return err2
		}

		var rawSecret RawSecret
		if len(secretData) == 1 && secretData[0].Default {
			rawSecret.Data = maybe.NewJust(secretData[0].Payload)
		} else {
			res := make(map[string][]byte, len(secretData))
			for _, secret := range secretData {
				res[secret.Name] = secret.Payload
			}
			rawSecret.Payload = maybe.NewJust(res)
		}

		data, err2 := s.secretSerializer.RawSerialize(rawSecret)
		if err2 != nil {
			return errors.Wrap(err2, "failed to serialize secret")
		}

		err2 = s.storage.Store(ctx, entryPath, data)
		if err2 != nil {
			return errors.Wrapf(err2, "failed to store secret %s", entryPath)
		}
	}

	return err
}

func (s *store) pack(ctx context.Context) error {
	const root = ""
	entries, err := s.list(ctx, root)
	if err != nil {
		return err
	}

	for _, entryPath := range inlinePaths(entries) {
		err = s.packSecret(ctx, entryPath)
		if err != nil {
			return errors.Wrapf(err, "failed to pack secret %s", entryPath)
		}
	}

	// packing store back may introduce changes in file, mark changedAdded = true to store changes
	s.changedAdded = true

	return nil
}

func (s *store) packSecret(ctx context.Context, entryPath string) (err error) {
	data, err := s.storage.Get(ctx, entryPath)
	if err != nil {
		return err
	}

	if !maybe.Valid(data) {
		return errors.New("secret not found")
	}

	res, err := s.secretSerializer.RawDeserialize(maybe.Just(data))
	if err != nil {
		return errors.Wrapf(err, "failed to deserialize secret")
	}

	secret := Secret{Payload: map[string][]byte{}}
	switch {
	case maybe.Valid(res.Data):
		encryptedData, err2 := s.encrypt.Encrypt(maybe.Just(res.Data), s.manifest.Recipients)
		if err2 != nil {
			return errors.Wrap(err2, "failed to encrypt data")
		}

		secret.addData(maybe.Maybe[string]{}, encryptedData)

	case maybe.Valid(res.Payload):
		for k, v := range maybe.Just(res.Payload) {
			encryptedData, err2 := s.encrypt.Encrypt(v, s.manifest.Recipients)
			if err2 != nil {
				return errors.Wrap(err2, "failed to encrypt data")
			}

			secret.addData(maybe.NewJust(k), encryptedData)
		}
	default:
		return errors.New("unknown secret deserialize result")
	}

	secretBytes, err := s.secretSerializer.Serialize(secret)
	if err != nil {
		return err
	}

	return s.storage.Store(ctx, entryPath, secretBytes)
}

func inlinePaths(entries []storage.Entry) []string {
	var recursiveInlinePath func(e storage.Entry) []string
	recursiveInlinePath = func(e storage.Entry) []string {
		if len(e.Children) == 0 {
			return []string{e.Name}
		}

		var res []string
		for _, child := range e.Children {
			childsPath := recursiveInlinePath(child)
			res = append(res, slices.Map(childsPath, func(p string) string {
				return path.Join(e.Name, p)
			})...)
		}

		return res
	}

	var res []string
	for _, entry := range entries {
		res = append(res, recursiveInlinePath(entry)...)
	}
	return res
}
