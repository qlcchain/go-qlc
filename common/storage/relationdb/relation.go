package relationdb

import (
	"errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/qlcchain/go-qlc/config"
)

func NewDB(cfg *config.Config) (RelationDB, error) {
	dbStr := cfg.DB.Driver
	switch dbStr {
	case "sqlite", "sqlite3":
		return NewSqliteDB(cfg)
	case "mysql":
		return NewMySqlDB(cfg)
	}
	return nil, errors.New("unsupported driver")
}

type RelationDB interface {
	Store() *sqlx.DB
	CreateTable(string, map[string]interface{}, string) string
	Set(string, map[string]interface{}) (string, []interface{})
	Delete(string, map[string]interface{}) string
	ConvertSchemaType(typ interface{}) string

	Close() error
}
