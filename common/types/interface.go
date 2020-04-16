package types

type Serializer interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

type Convert interface {
	ConvertToSchema() ([]Schema, error)
}

type Schema interface {
	IdentityID() string
	DeleteKey() string
}
