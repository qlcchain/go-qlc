package relation

import (
	"github.com/qlcchain/go-qlc/common/types"
)

type Store interface {
	AccountBlocks(address types.Address, limit int, offset int) ([]types.Hash, error)
	BlocksCount() (uint64, error)
	BlocksCountByType() (map[string]uint64, error)
	Blocks(limit int, offset int) ([]types.Hash, error)
	PhoneBlocks(phone []byte, sender bool, limit int, offset int) ([]types.Hash, error)
	MessageBlocks(hash types.Hash, limit int, offset int) ([]types.Hash, error)
	AddBlock(block *types.StateBlock) error
	DeleteBlock(hash types.Hash) error
	Close() error
}
