package relationdb

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
)

func openMysql(cfg *config.Config) (*sqlx.DB, error) {
	if err := util.CreateDirIfNotExist(cfg.SqliteDir()); err != nil {
		return nil, err
	}
	//c := mysql.Config{
	//	User:   "root",
	//	Passwd: "123456",
	//	Net:    "tcp",
	//	Addr:   "127.0.0.1:3306",
	//	DBName: "goqlc",
	//}
	//fmt.Println(c.FormatDSN())
	donnctionString := "root:123456@tcp(127.0.0.1:3306)/goqlc?charset=utf8"
	//db, err := sqlx.Connect(cfg.DB.Driver, cfg.DB.ConnectionString)
	db, err := sqlx.Connect("mysql", donnctionString)
	if err != nil {
		fmt.Println("connect mysql error: ", err)
		return nil, err
	}

	return db, nil
}
