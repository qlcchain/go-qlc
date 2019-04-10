package db

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type DBSQL struct {
	db     *sqlx.DB
	logger *zap.SugaredLogger
}

func NewSQLDB(path string) (*DBSQL, error) {
	db, err := createDBBySqlite(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	dbsql := DBSQL{db: db, logger: log.NewLogger("relation/dbsql")}
	if err := dbsql.initDB(); err != nil {
		return nil, err
	}
	return &dbsql, nil
}

func createDBBySqlite(path string) (*sqlx.DB, error) {
	return sqlx.Connect("sqlite3", path)
}

func (s *DBSQL) initDB() error {
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS BLOCKHASH
		(   id integer PRIMARY KEY AUTOINCREMENT,
			hash char(32),
			type char(10),
			address char(32),
			timestamp integer
		)`,
		`CREATE TABLE IF NOT EXISTS BLOCKMESSAGE 
		(	id integer PRIMARY KEY AUTOINCREMENT,
			hash char(32),
			sender text ,
			receiver text ,
			message	char(32),
			timestamp integer
		)`,
		`CREATE INDEX IF NOT EXISTS index_sender   ON BLOCKMESSAGE (sender);  `,
		`CREATE INDEX IF NOT EXISTS index_receiver ON BLOCKMESSAGE (receiver);`,
		`CREATE INDEX IF NOT EXISTS index_message  ON BLOCKMESSAGE (message); `,
	}

	for _, sql := range sqls {
		if _, err := s.db.Exec(sql); err != nil {
			s.logger.Error(err)
			return err
		}
	}
	return nil
}

func (s *DBSQL) Create(table TableName, condition map[Column]interface{}) error {
	sql := createSql(table, condition)
	s.logger.Debug(sql)
	if _, err := s.db.Exec(sql); err != nil {
		s.logger.Error(err)
		return err
	}
	return nil
}

func (s *DBSQL) Read(table TableName, condition map[Column]interface{}, offset int, limit int, order Column, dest interface{}) error {
	sql := readSql(table, condition, offset, limit, order)
	s.logger.Debug(sql)
	err := s.db.Select(dest, sql)
	if err != nil {
		s.logger.Error(err)
	}
	return err
}

func (s *DBSQL) Update(table TableName, condition map[Column]interface{}) error {
	panic("implement me")
}

func (s *DBSQL) Delete(table TableName, condition map[Column]interface{}) error {
	sql := deleteSql(table, condition)
	s.logger.Debug(sql)
	_, err := s.db.Exec(sql)
	if err != nil {
		s.logger.Error(err)
	}
	return err
}

func (s *DBSQL) Count(table TableName, dest interface{}) error {
	sql := fmt.Sprintf("select count (*) as count from %s", table.String())
	s.logger.Debug(sql)
	err := s.db.Get(&dest, sql)
	if err != nil {
		s.logger.Error(err)
		return err
	}
	return nil
}

func (s *DBSQL) Group(table TableName, column Column, dest interface{}) error {
	sql := fmt.Sprintf("select %s, count(*) as count from %s  group by %s", column.String(), table.String(), column.String())
	s.logger.Debug(sql)
	err := s.db.Select(dest, sql)
	if err != nil {
		s.logger.Error(err)
		return err
	}
	return nil
}

func createSql(table TableName, condition map[Column]interface{}) string {
	var key []string
	var value []string
	for k, v := range condition {
		key = append(key, k.String())
		switch v.(type) {
		case string:
			value = append(value, fmt.Sprintf("'%s'", v.(string)))
		case int64:
			value = append(value, strconv.FormatInt(v.(int64), 10))
		}
	}
	sql := fmt.Sprintf("insert into %s (%s) values (%s)", table.String(), strings.Join(key, ","), strings.Join(value, ","))
	return sql
}

func readSql(table TableName, condition map[Column]interface{}, offset int, limit int, order Column) string {
	var sql string
	var para []string
	if condition != nil {
		for k, v := range condition {
			switch v.(type) {
			case string:
				para = append(para, k.String()+" = '"+v.(string)+"' ")
			case int64:
				para = append(para, k.String()+" = "+strconv.FormatInt(v.(int64), 10))
			}
		}
	}
	if len(para) != 0 {
		sql = fmt.Sprintf("select * from %s  where %s ", table.String(), strings.Join(para, " or "))
	} else {
		sql = fmt.Sprintf("select * from %s ", table.String())
	}
	if order != ColumnNoNeed {
		sql = sql + " order by  " + order.String()
	}
	if limit != -1 {
		sql = sql + " limit " + strconv.Itoa(limit)
	}
	if offset != -1 {
		sql = sql + " offset " + strconv.Itoa(offset)
	}
	return sql
}

func deleteSql(table TableName, condition map[Column]interface{}) string {
	var sql string
	var para []string
	if condition != nil {
		for k, v := range condition {
			switch v.(type) {
			case string:
				para = append(para, k.String()+" = '"+v.(string)+"' ")
			case int64:
				para = append(para, k.String()+" = "+strconv.FormatInt(v.(int64), 10))
			}
		}
	}
	if len(para) != 0 {
		sql = fmt.Sprintf("delete from %s  where %s ", table.String(), strings.Join(para, " or "))
	} else {
		sql = fmt.Sprintf("delete from %s ", table.String())
	}
	return sql
}
