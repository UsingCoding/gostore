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

func (manager *encryptionManager) ImportRawIdentity(provider encryption.Provider, data []byte) (encryption.Identity, error) {
	service, err := manager.makeEncServiceForProvider(provider)
	if err != nil {
		return encryption.Identity{}, err
	}

	return service.loadRawIdentity(data)
}

func (manager *encryptionManager) ExportRawIdentity(identity encryption.Identity) ([]byte, error) {
	service, err := manager.makeEncServiceForProvider(identity.Provider)
	if err != nil {
		return nil, err
	}

	return service.exportRawIdentity(identity)
}

func (manager *encryptionManager) makeEncService(e encryption.Encryption) (encService, error) {
	switch e {
	case encryption.AgeEncryption:
		return newAgeService(), nil
	default:
		return nil, errors.Errorf("unknown type of encryption %s", e)
	}
}

func (manager *encryptionManager) makeEncServiceForProvider(p encryption.Provider) (encService, error) {
	switch p {
	case encryption.AgeIdentityProvider:
		return newAgeService(), nil
	default:
		return nil, errors.Errorf("unknown provider %s", p)
	}
}

type encService interface {
	encryption.Service

	generateIdentity() (encryption.Identity, error)

	loadRawIdentity(data []byte) (encryption.Identity, error)
	exportRawIdentity(identity encryption.Identity) ([]byte, error)
}
