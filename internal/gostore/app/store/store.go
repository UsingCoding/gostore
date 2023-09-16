package store

import (
	"context"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/pkg/errors"

	"gostore/internal/common/maybe"
	commonstrings "gostore/internal/common/strings"
	"gostore/internal/gostore/app/encryption"
	"gostore/internal/gostore/app/storage"
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
	err := allowedPath(path)
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

func (s *store) get(ctx context.Context, path string, key maybe.Maybe[string]) ([]SecretData, error) {
	err := allowedPath(path)
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
	err := allowedPath(path)
	if err != nil {
		return err
	}

	if !maybe.Valid(key) {
		return s.storage.Remove(ctx, path)
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
		return s.storage.Remove(ctx, path)
	}

	secretData, err := s.secretSerializer.Serialize(secret)
	if err != nil {
		return err
	}

	err = s.storage.Store(ctx, path, secretData)
	if err != nil {
		return err
	}

	return nil
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

// checks that path is not store internal object
func allowedPath(path string) error {
	if !commonstrings.HasPrefix(path, reservedPaths) {
		return nil
	}

	return errors.Errorf("access to store internal objects")
}
