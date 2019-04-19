package db

import (
	"fmt"
	"path"

	"github.com/jmoiron/sqlx"
	"github.com/qlcchain/go-qlc/common/util"
)

type DB interface {
	Create()
}

func openSqlite(dir, conn string) (*sqlx.DB, error) {
	if err := util.CreateDirIfNotExist(dir); err != nil {
		return nil, err
	}
	dataSourceName := fmt.Sprintf("file:%s?_auth&_auth_user=%s&_auth_pass=%s", path.Join(dir, "sqlite3.db"), user, password)
	db, err := sqlx.Connect("sqlite3", dataSourceName)
	if err != nil {
		fmt.Println("connect sqlite error: ", err)
		return nil, err
	}
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(100)

	sqls := []string{
		`CREATE TABLE IF NOT EXISTS BLOCKHASH
		(   id integer PRIMARY KEY AUTOINCREMENT,
			hash char(32),
			type varchar(10),
			address char(32),
			timestamp integer
		)`,
		`CREATE TABLE IF NOT EXISTS BLOCKMESSAGE 
		(	id integer PRIMARY KEY AUTOINCREMENT,
			hash char(32),
			sender varchar(15),
			receiver varchar(15) ,
			message	char(32),
			timestamp integer
		)`,
		`CREATE INDEX IF NOT EXISTS index_sender   ON BLOCKMESSAGE (sender);  `,
		`CREATE INDEX IF NOT EXISTS index_receiver ON BLOCKMESSAGE (receiver);`,
		`CREATE INDEX IF NOT EXISTS index_message  ON BLOCKMESSAGE (message); `,
	}

	for _, sql := range sqls {
		if _, err := db.Exec(sql); err != nil {
			fmt.Printf("exec error, sql: %s, err: %s \n", sql, err.Error())
			return nil, err
		}
	}
	return db, nil
}
