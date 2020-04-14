package db

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
		return openSqlite(cfg)
	case "mysql":
		return openMysql(cfg)
	case "postgres":
		return openPostgres(cfg)
	}
	return nil, errors.New("unsupported driver")
}
