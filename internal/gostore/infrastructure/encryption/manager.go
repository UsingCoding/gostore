package encryption

import (
	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
)

func NewManager() encryption.Manager {
	return &encryptionManager{}
}

type encryptionManager struct{}

func (manager *encryptionManager) GenerateIdentity(e encryption.Encryption) (encryption.Identity, error) {
	service, err := manager.makeEncService(e)
	if err != nil {
		return encryption.Identity{}, err
	}

	return service.generateIdentity()
}

func (manager *encryptionManager) PrivateKey(key encryption.Recipient) (encryption.PrivateKey, error) {
	return nil, nil
}

func (manager *encryptionManager) EncryptService(e encryption.Encryption) (encryption.Service, error) {
	return manager.makeEncService(e)
}

func (manager *encryptionManager) makeEncService(e encryption.Encryption) (encService, error) {
	switch e {
	case encryption.AgeEncryption:
		return newAgeService(), nil
	default:
		return nil, errors.Errorf("unknown type of encryption %s", e)
	}
}

type encService interface {
	encryption.Service

	generateIdentity() (encryption.Identity, error)
}
