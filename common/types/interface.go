package types

type Serializer interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

type Convert interface {
	TableConvert() ([]Table, error)
}

type Table interface {
	TableID() string
	DeleteKey() string
}
