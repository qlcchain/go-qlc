package relation

import (
	"fmt"
	"strings"
)

type schema struct {
	tableName string
	create    string
	insert    string
	delete    string
}

func convertSchemaType(db string, typ string) string {
	switch db {
	case "sqlite", "sqlite3":
		return typ
	case "mysql":
		// TODO mysql type may different
		return typ
	default:
		return "varchar(100)"
	}
}

func create(tableName string, columns map[string]string, key string) string {
	cs := ""
	keyTyp := ""
	for k, v := range columns {
		if k != key {
			cs = cs + fmt.Sprintf(" %s %s ,", k, v)
		} else {
			keyTyp = v
		}
	}
	cs = strings.TrimRight(cs, ",")
	if key == "" {
		return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s)`, tableName, cs)
	} else {
		return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s ( %s %s PRIMARY KEY NOT NULL, %s)`, tableName, key, keyTyp, cs)
	}
}

func insert(tableName string, columns []string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (:%s)", tableName, strings.Join(columns, ","), strings.Join(columns, ", :"))
}
