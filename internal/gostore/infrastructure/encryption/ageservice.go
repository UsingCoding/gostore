package encryption

import (
	"bytes"
	"io"

	"filippo.io/age"
	"filippo.io/age/armor"
	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/pkg/errors"

	"gostore/internal/gostore/app/encryption"
)

func newAgeService() encService {
	return &ageService{}
}

type ageService struct{}

func (s *ageService) Encrypt(data []byte, recipients []encryption.Recipient) ([]byte, error) {
	var buffer bytes.Buffer
	w := armor.NewWriter(&buffer)
	recps, err := slices.MapErr(recipients, func(r encryption.Recipient) (age.Recipient, error) {
		return age.ParseX25519Recipient(string(r))
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse recipients")
	}

	encryptedWriter, err := age.Encrypt(w, recps...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_, err = encryptedWriter.Write(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt secret")
	}

	err = encryptedWriter.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to close writer after encryption with rcpts %s", recipients)
	}

	err = w.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to close armored writer after encryption with rpcts %s", recipients)
	}

	return buffer.Bytes(), nil
}

func (s *ageService) Decrypt(data []byte, identities []encryption.Identity) ([]byte, error) {
	mappedIdentities, err := slices.MapErr(identities, s.mapIdentity)
	if err != nil {
		return nil, err
	}

	src := bytes.NewReader(data)
	ar := armor.NewReader(src)

	reader, err := age.Decrypt(ar, mappedIdentities...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt data")
	}

	decodedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read decoded data")
	}

	return decodedData, nil
}

func (s *ageService) generateIdentity() (encryption.Identity, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return encryption.Identity{}, errors.WithStack(err)
	}

	return encryption.Identity{
		Recipient:  encryption.Recipient(identity.Recipient().String()),
		PrivateKey: encryption.PrivateKey(identity.String()),
	}, nil
}

func (s *ageService) mapIdentity(identity encryption.Identity) (age.Identity, error) {
	i, err := age.ParseX25519Identity(string(identity.PrivateKey))
	return i, errors.Wrapf(err, "failed to map identity for recipient %s", identity.Recipient)
}
