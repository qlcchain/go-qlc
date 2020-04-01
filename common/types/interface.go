package types

type Serializer interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}
