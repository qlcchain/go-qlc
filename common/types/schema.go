package types

import (
	"fmt"
	"strings"
)

type Serializer interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

type Schema interface {
	TableName() string
	TableSchema(dbType string) string
	SetRelation() (string, []interface{})
	RemoveRelation() string
}

func CreateSchema(tablename string, fields map[string]interface{}, key string, dbType string) string {
	switch dbType {
	case "sqlite", "sqlite3":
		cf := ""
		for k, v := range fields {
			cf = cf + fmt.Sprintf(" %s %s ,", k, ConvertSchemaType(v, dbType))
		}
		cf = strings.TrimRight(cf, ",")
		return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
		( id integer PRIMARY KEY AUTOINCREMENT, %s)`, tablename, cf)
	case "mysql":
		cf := ""
		for k, v := range fields {
			cf = fmt.Sprintf(" %s %s ,", k, ConvertSchemaType(v, dbType))
		}
		cf = strings.TrimRight(cf, ",")
		return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
		( id integer PRIMARY KEY AUTO_INCREMENT, %s)`, tablename, cf)
	default:
		return ""
	}
}

func ConvertSchemaType(typ interface{}, dbType string) string {
	switch dbType {
	case "sqlite", "sqlite3":
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
	case "mysql":
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
	default:
		return ""
	}
}

//type TableSchema struct {
//	Name   string
//	Index  string
//	Column []interface{}
//}
