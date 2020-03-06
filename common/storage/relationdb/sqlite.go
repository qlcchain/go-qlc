package relationdb

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
)

func openSqlite(cfg *config.Config) (*sqlx.DB, error) {
	if err := util.CreateDirIfNotExist(cfg.SqliteDir()); err != nil {
		return nil, err
	}
	db, err := sqlx.Connect(cfg.DB.Driver, cfg.DB.ConnectionString)
	if err != nil {
		fmt.Println("connect sqlite error: ", err)
		return nil, err
	}
	//DBStore.SetMaxOpenConns(200)
	//DBStore.SetMaxIdleConns(100)

	return db, nil
}

type SqliteDB struct {
	store *sqlx.DB
}

func NewSqliteDB(cfg *config.Config) (*SqliteDB, error) {
	db, err := openSqlite(cfg)
	if err != nil {
		return nil, err
	}
	s := &SqliteDB{
		store: db,
	}
	return s, nil
}

func (s *SqliteDB) CreateTable(tableName string, fields map[string]interface{}, key string) string {
	cf := ""
	for k, v := range fields {
		cf = cf + fmt.Sprintf(" %s %s ,", k, s.ConvertSchemaType(v))
	}
	cf = strings.TrimRight(cf, ",")
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
		( id integer PRIMARY KEY AUTOINCREMENT, %s)`, tableName, cf)
}

func (s *SqliteDB) ConvertSchemaType(typ interface{}) string {
	switch typ.(type) {
	case types.Address:
		return "char(32)"
	case types.Hash:
		return "char(32)"
	case types.BlockType:
		return "varchar(10)"
	case int64:
		return "integer"
	case types.Balance:
		return "integer"
	default:
		return "varchar(30)"
	}
}

func (s *SqliteDB) Set(tableName string, vals map[string]interface{}) (string, []interface{}) {
	str := ""
	val := make([]interface{}, 0)
	index := 0
	indexStr := ""
	//	str = fmt.Sprintf("INSERT INTO %s (hash,type,address,timestamp) VALUES ($1, $2, $3, $4)", b.TableName())
	//	val := []interface{}{b.GetHash().String(), b.GetType().String(), b.GetAddress().String(), b.GetTimestamp()}
	for k, v := range vals {
		str = str + fmt.Sprintf(" %s ,", k)
		val = append(val, v)
		index = index + 1
		indexStr = indexStr + fmt.Sprintf(" $%d ,", index)
	}
	str = fmt.Sprintf("INSERT INTO %s (  %s ) VALUES ( %s )", tableName, strings.TrimRight(str, ","), strings.TrimRight(indexStr, ","))
	return str, val
}

func (s *SqliteDB) Delete(tableName string, vals map[string]interface{}) string {
	//TODO delete combined key
	field := ""
	val := ""
	for k, v := range vals {
		field = k
		val = v.(string)
	}
	// fmt.Sprintf("DELETE FROM %s WHERE hash = '%s'", b.TableName(), b.GetHash().String())
	return fmt.Sprintf("DELETE FROM %s WHERE %s = '%s'", tableName, field, val)
}

func (s *SqliteDB) Store() *sqlx.DB {
	return s.store
}

func (s *SqliteDB) Close() error {
	return s.store.Close()
}
