package relationdb

import (
	"errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/qlcchain/go-qlc/config"
)

func NewDB(cfg *config.Config) (*sqlx.DB, error) {
	dbStr := cfg.DB.Driver
	switch dbStr {
	case "sqlite", "sqlite3":
		db, err := openSqlite(cfg)
		if err != nil {
			return nil, err
		}
		return db, nil
	}
	return nil, errors.New("unsupported driver")
}
