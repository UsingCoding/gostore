package encryption

type Provider string

const (
	AgeIdentityProvider = "age"
)

type Identity struct {
	Provider
	Recipient
	PrivateKey
}

type Recipient []byte

func (k Recipient) String() string {
	return string(k)
}

type PrivateKey []byte

func (k PrivateKey) String() string {
	return string(k)
}

type Service interface {
	Encrypt(data []byte, recipients []Recipient) ([]byte, error)
	Decrypt(data []byte, identities []Identity) ([]byte, error)
}
