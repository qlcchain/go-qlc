package relation

import (
	"github.com/jmoiron/sqlx"

	"github.com/qlcchain/go-qlc/common/types"
)

type Store interface {
	Blocks(limit int, offset int) ([]types.Hash, error)
	BlocksByAccount(address types.Address, limit int, offset int) ([]types.Hash, error)
	BlocksCount() (uint64, error)
	BlocksCountByType() (map[string]uint64, error)

	Select(dest interface{}, query string) error
	Get(dest interface{}, query string) error
	Queryx(query string) (*sqlx.Rows, error)
	DB() *sqlx.DB
	Close() error
	EmptyStore() error
}
