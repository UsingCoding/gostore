package vars

type Kind string

const (
	ConfigKind = Kind("config")
	StoreKind  = Kind("store")
	SecretKind = Kind("secret")
)
