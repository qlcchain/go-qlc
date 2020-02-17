package types

type Serialize interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

type Schema interface {
	TableName() string
	TableSchema() string
	SetRelation() (string, []interface{})
	RemoveRelation() string
}

func ConvertSchemaType(typ interface{}) string {
	switch typ.(type) {
	case Address:
		return "char(32)"
	case Hash:
		return "char(32)"
	case BlockType:
		return "varchar(10)"
	case int64:
		return "integer"
	case Balance:
		return "integer"
	default:
		return "varchar(30)"
	}
}

type TableSchema struct {
	Name   string
	Index  string
	Column []interface{}
}
