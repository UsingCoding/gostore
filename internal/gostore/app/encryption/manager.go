package encryption

type Manager interface {
	// GenerateIdentity creates new identity
	GenerateIdentity(encryption Encryption) (Identity, error)
	// PrivateKey returns private key from host
	PrivateKey(key Recipient) (PrivateKey, error)
	EncryptService(encryption Encryption) (Service, error)

	ImportRawIdentity(provider Provider, data []byte) (Identity, error)
	ExportRawIdentity(identity Identity) ([]byte, error)
}
