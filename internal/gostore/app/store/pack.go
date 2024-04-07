package store

import (
	"bytes"
	"context"
	stderrors "errors"
	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
	"github.com/pkg/errors"
	"path"
)

func (s *store) unpack(ctx context.Context) (err error) {
	err = s.assertPacked()
	if err != nil {
		return err
	}

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
	err := s.assertUnpacked()
	if err != nil {
		return err
	}

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

	secret := initSecret()
	switch {
	case maybe.Valid(res.Data):
		secret.addData(maybe.Maybe[string]{}, maybe.Just(res.Data))

	case maybe.Valid(res.Payload):
		for k, v := range maybe.Just(res.Payload) {
			secret.addData(maybe.NewJust(k), v)
		}
	default:
		return errors.New("unknown secret deserialize result")
	}

	latest, err := s.storage.GetLatest(ctx, entryPath)
	if err != nil {
		return err
	}

	if maybe.Valid(latest) {
		secret, err = s.mergeWithLatest(ctx, secret, maybe.Just(latest))
		if err != nil {
			return errors.Wrapf(err, "failed to merge secret %s", entryPath)
		}
	} else {
		err = secret.encrypt(func(data []byte) ([]byte, error) {
			return s.encryption.Encrypt(data, s.manifest.Recipients)
		})
		if err != nil {
			return errors.Wrapf(err, "failed to encrypt secret %s", entryPath)
		}
	}

	secretBytes, err := s.secretSerializer.Serialize(secret)
	if err != nil {
		return err
	}

	return s.storage.Store(ctx, entryPath, secretBytes)
}

func (s *store) mergeWithLatest(ctx context.Context, secret Secret, latestData []byte) (Secret, error) {
	latest, err := s.secretSerializer.Deserialize(latestData)
	if err != nil {
		return Secret{}, errors.Wrap(err, "failed to deserialize latest secret")
	}

	err = secret.iterate(func(k string, v []byte) error {
		latestEncryptedV, ok := maybe.JustValid(latest.getByKey(k))
		if !ok {
			// new key in secret, nothing to compare
			return nil
		}

		latestV, err2 := s.decrypt(ctx, latestEncryptedV)
		if err2 != nil {
			return err2
		}

		eq := bytes.Equal(v, latestV)
		if eq {
			secret.addData(maybe.NewJust(k), latestEncryptedV)

			return nil
		}

		// key value is updated
		encryptedV, err2 := s.encrypt(v)
		if err2 != nil {
			return err2
		}

		secret.addData(maybe.NewJust(k), encryptedV)

		return nil
	})

	return secret, err
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

func (s *store) assertPacked() error {
	if !s.manifest.Unpacked {
		return nil
	}

	return errors.New("store is unpacked")
}

func (s *store) assertUnpacked() error {
	if s.manifest.Unpacked {
		return nil
	}

	return errors.New("store is packed")
}