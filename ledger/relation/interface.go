package relation

import "github.com/qlcchain/go-qlc/common/types"

type Store interface {
	Blocks(limit int, offset int) ([]types.Hash, error)
	BlocksByAccount(address types.Address, limit int, offset int) ([]types.Hash, error)
	BlocksCount() (uint64, error)
	BlocksCountByType() (map[string]uint64, error)

	Select(dest interface{}, query string, args ...interface{}) error
	Close() error
	EmptyStore() error
}
