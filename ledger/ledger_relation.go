package ledger

import "github.com/qlcchain/go-qlc/common/types"

type Relation interface {
	Blocks(limit int, offset int) ([]types.Hash, error)
	BlocksByAccount(address types.Address, limit int, offset int) ([]types.Hash, error)
	BlocksCount() (uint64, error)
	BlocksCountByType() (map[string]uint64, error)

	EmptyRelation() error
}

func (l *Ledger) Blocks(limit int, offset int) ([]types.Hash, error) {
	return l.relation.Blocks(limit, offset)
}

func (l *Ledger) BlocksByAccount(address types.Address, limit int, offset int) ([]types.Hash, error) {
	return l.relation.BlocksByAccount(address, limit, offset)
}

func (l *Ledger) BlocksCount() (uint64, error) {
	return l.relation.BlocksCount()
}

func (l *Ledger) BlocksCountByType() (map[string]uint64, error) {
	return l.relation.BlocksCountByType()
}

func (l *Ledger) EmptyRelation() error {
	return l.relation.EmptyStore()
}
