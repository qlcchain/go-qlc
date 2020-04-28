package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
)

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

func (l *Ledger) SelectRelation(dest interface{}, query string) error {
	return l.relation.Select(dest, query)
}

func (l *Ledger) GetRelation(dest interface{}, query string) error {
	return l.relation.Get(dest, query)
}

//func (l *Ledger) Queryx(fn func(v types.Table) error, t types.Table, query string, args ...interface{}) error {
//	rows, err := l.relation.Queryx(query, args)
//	if err != nil {
//		return fmt.Errorf("queryx: %s", err)
//	}
//	for rows.Next() {
//		err := rows.StructScan(t)
//		if err != nil {
//			return fmt.Errorf("struct scan: %s", err)
//		}
//		if err := fn(t); err != nil {
//			return err
//		}
//	}
//	return nil
//}
