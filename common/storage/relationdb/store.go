package relationdb

import "github.com/jmoiron/sqlx"

type Store interface {
	Insert(table string, condition map[string]interface{}) error
	Read(table string, condition map[string]interface{}, offset int, limit int, order map[string]bool, dest interface{}) error
	Update(table string, condition map[string]interface{}) error
	Delete(table string, condition map[string]interface{}) error
	Count(table string, condition map[string]interface{}, dest interface{}) error
	Group(table string, column string, dest interface{}) error
	BatchUpdate(fn func(tx *sqlx.Tx) error) error
}
