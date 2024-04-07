package store

import (
	"context"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	commonstrings "github.com/UsingCoding/gostore/internal/common/strings"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/storage"
)

var (
	reservedPaths = []string{
		ManifestPath,
	}
)

type store struct {
	manifest Manifest

	storage          storage.Storage
	encrypt          encryption.Service
	identityProvider IdentityProvider

	secretSerializer SecretSerializer

	changedAdded bool
}

func (s *store) add(
	ctx context.Context,
	path string,
	key maybe.Maybe[string],
	data []byte,
) error {
	err := s.assertPacked()
	if err != nil {
		return err
	}

	err = allowedPaths(path)
	if err != nil {
		return err
	}

	existedSecret, err := s.storage.Get(ctx, path)
	if err != nil {
		return err
	}

	encryptedData, err := s.encrypt.Encrypt(data, s.manifest.Recipients)
	if err != nil {
		return errors.Wrap(err, "failed to encrypt data")
	}

	var secret Secret
	if maybe.Valid(existedSecret) {
		secret, err = s.secretSerializer.Deserialize(maybe.Just(existedSecret))
		if err != nil {
			return err
		}
	} else {
		secret = initSecret()
	}

	secret.addData(key, encryptedData)

	secretBytes, err := s.secretSerializer.Serialize(secret)
	if err != nil {
		return err
	}

	err = s.storage.Store(ctx, path, secretBytes)
	if err != nil {
		return err
	}

	s.changedAdded = true

	return nil
}

func (s *store) copy(ctx context.Context, src, dst string) error {
	err := s.assertPacked()
	if err != nil {
		return err
	}

	err = allowedPaths(src, dst)
	if err != nil {
		return err
	}

	err = s.storage.Copy(ctx, src, dst)
	if err != nil {
		return err
	}

	s.changedAdded = true

	return nil
}

func (s *store) move(ctx context.Context, src, dst string) error {
	err := s.assertPacked()
	if err != nil {
		return err
	}

	err = allowedPaths(src, dst)
	if err != nil {
		return err
	}

	err = s.storage.Move(ctx, src, dst)
	if err != nil {
		return err
	}

	s.changedAdded = true

	return nil
}

func (s *store) get(ctx context.Context, path string, key maybe.Maybe[string]) ([]SecretData, error) {
	err := s.assertPacked()
	if err != nil {
		return nil, err
	}

	err = allowedPaths(path)
	if err != nil {
		return nil, err
	}

	secretBytes, err := s.storage.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if !maybe.Valid(secretBytes) {
		return nil, nil
	}

	secret, err := s.secretSerializer.Deserialize(maybe.Just(secretBytes))
	if err != nil {
		return nil, err
	}

	var availableIdentities []encryption.Identity
	for _, recipient := range s.manifest.Recipients {
		i, err2 := s.identityProvider.IdentityByRecipient(ctx, recipient)
		if err2 != nil {
			return nil, errors.Wrap(err2, "failed to get identity")
		}

		if maybe.Valid(i) {
			availableIdentities = append(availableIdentities, maybe.Just(i))
		}
	}

	if len(availableIdentities) == 0 {
		return nil, errors.New("no available identities found")
	}

	secretsData := secret.get(key)
	if len(secretsData) == 0 {
		return nil, nil
	}

	return slices.MapErr(secretsData, func(secret SecretData) (SecretData, error) {
		decryptedData, err2 := s.encrypt.Decrypt(secret.Payload, availableIdentities)
		if err2 != nil {
			return SecretData{}, err2
		}

		secret.Payload = decryptedData
		return secret, nil
	})
}

func (s *store) list(ctx context.Context, path string) ([]storage.Entry, error) {
	entries, err := s.storage.List(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extreact entries from storage")
	}

	entries = slices.Filter(entries, func(entry storage.Entry) bool {
		return !commonstrings.HasPrefix(entry.Name, reservedPaths)
	})

	return entries, nil
}

func (s *store) remove(ctx context.Context, path string, key maybe.Maybe[string]) error {
	err := s.assertPacked()
	if err != nil {
		return err
	}

	err = allowedPaths(path)
	if err != nil {
		return err
	}

	if !maybe.Valid(key) {
		err = s.storage.Remove(ctx, path)
		if err != nil {
			return err
		}
		s.changedAdded = true
		return nil
	}

	secretBytes, err := s.storage.Get(ctx, path)
	if err != nil {
		return err
	}

	if !maybe.Valid(secretBytes) {
		return nil
	}

	secret, err := s.secretSerializer.Deserialize(maybe.Just(secretBytes))
	if err != nil {
		return err
	}

	secret.remove(maybe.Just(key))

	// if secret empty - remove from storage
	if secret.empty() {
		err = s.storage.Remove(ctx, path)
		if err != nil {
			return err
		}
		s.changedAdded = true
	}

	secretData, err := s.secretSerializer.Serialize(secret)
	if err != nil {
		return err
	}

	err = s.storage.Store(ctx, path, secretData)
	if err != nil {
		return err
	}

	s.changedAdded = true

	return nil
}

func (s *store) sync(ctx context.Context) error {
	err := s.assertPacked()
	if err != nil {
		return err
	}

	err = s.storage.Pull(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to pull storage")
	}

	err = s.storage.Push(ctx)
	return errors.Wrap(err, "failed to push storage")
}

func (s *store) close() error {
	if !s.changedAdded {
		return nil
	}
	return s.storage.Commit(context.Background(), "Changes committed")
}

func (s *store) rollback(ctx context.Context) error {
	return s.storage.Rollback(ctx)
}

func (s *store) assertPacked() error {
	if !s.manifest.Unpacked {
		return nil
	}

	return errors.New("store is unpacked")
}

// checks that path is not store internal object
func allowedPaths(paths ...string) error {
	for _, p := range paths {
		if commonstrings.HasPrefix(p, reservedPaths) {
			return errors.Errorf("access to store internal objects in %s", p)
		}
	}

	return nil
}
