package types

type Serializer interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

type Schema interface {
	TableName() string
	TableSchema() (map[string]interface{}, string)
	SetRelation() map[string]interface{}
	RemoveRelation() map[string]interface{}
}
