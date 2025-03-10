package store

import (
	"bytes"
	"context"
	stderrors "errors"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/progress"
)

const (
	packWorkers = 50
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
	tree, err := s.list(ctx, root)
	if err != nil {
		return err
	}

	inlinedTree := tree.Inline()

	p := progress.FromCtx(ctx).Alter(
		progress.WithMax(int64(len(inlinedTree.Keys()))),
		progress.WithDescription("Unpacking store"),
		progress.WithIts(),
	)
	defer p.Finish()

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(packWorkers)

	for _, entryPath := range inlinedTree.Keys() {
		eg.Go(func() error {
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

			p.Inc()

			return nil
		})
	}

	return eg.Wait()
}

func (s *store) pack(ctx context.Context, params PackParams) error {
	err := s.assertUnpacked()
	if err != nil {
		return err
	}

	const root = ""
	tree, err := s.list(ctx, root)
	if err != nil {
		return err
	}

	inlinedTree := tree.Inline()

	p := progress.FromCtx(ctx).Alter(
		progress.WithMax(int64(len(inlinedTree.Keys()))),
		progress.WithDescription("Packing store"),
		progress.WithIts(),
	)
	defer p.Finish()

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(packWorkers)

	for _, entryPath := range inlinedTree.Keys() {
		eg.Go(func() error {
			err = s.packSecret(ctx, entryPath, params.SkipChangesCheck)
			if err != nil {
				return errors.Wrapf(err, "failed to pack secret %s", entryPath)
			}

			p.Inc()

			return nil
		})
	}

	// packing store back may introduce changes in objects
	s.operations.add(packOperation())

	return eg.Wait()
}

func (s *store) packSecret(ctx context.Context, entryPath string, skipChangesCheck bool) (err error) {
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

	var latest maybe.Maybe[[]byte]
	if !skipChangesCheck {
		latest, err = s.storage.GetLatest(ctx, entryPath)
		if err != nil {
			return err
		}
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
