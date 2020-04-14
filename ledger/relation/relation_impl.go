package relation

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/qlcchain/go-qlc/common/types"
)

func (r *Relation) count(s types.Schema) (uint64, error) {
	sql := fmt.Sprintf("select count (*) as total from %s", r.tables[s.IdentityID()].tableName)
	r.logger.Debug(sql)
	var i uint64
	err := r.db.Get(&i, sql)
	if err != nil {
		return 0, fmt.Errorf("count error, sql: %s, err: %s", sql, err.Error())
	}
	return i, nil
}

func (r *Relation) Blocks(limit int, offset int) ([]types.Hash, error) {
	block := new(types.BlockHash)
	var h []types.BlockHash
	sql := fmt.Sprintf("select * from %s order by timestamp desc, type desc limit %d offset %d", r.tables[block.IdentityID()].tableName, limit, offset)
	err := r.db.Select(&h, sql)
	if err != nil {
		return nil, fmt.Errorf("read error, sql: %s, err: %s", sql, err.Error())
	}
	return blockHash(h)
}

func (r *Relation) BlocksByAccount(address types.Address, limit int, offset int) ([]types.Hash, error) {
	block := new(types.BlockHash)
	var h []types.BlockHash
	sql := fmt.Sprintf("select * from %s where address = '%s' order by timestamp desc, type desc limit %d offset %d", r.tables[block.IdentityID()].tableName, address.String(), limit, offset)
	err := r.db.Select(&h, sql)
	if err != nil {
		return nil, fmt.Errorf("read error, sql: %s, err: %s", sql, err.Error())
	}
	return blockHash(h)
}

func (r *Relation) BlocksCount() (uint64, error) {
	return r.count(new(types.BlockHash))
}

func (r *Relation) BlocksCountByType() (map[string]uint64, error) {
	block := new(types.BlockHash)
	var t []blocksType
	sql := fmt.Sprintf("select type, count(*) as count from %s  group by type", r.tables[block.IdentityID()].tableName)
	r.logger.Debug(sql)
	err := r.db.Select(&t, sql)
	if err != nil {
		return nil, fmt.Errorf("group error, sql: %s, err: %s", sql, err.Error())
	}
	return blockType(t), nil
}

func (r *Relation) Select(dest interface{}, query string) error {
	r.logger.Debug(query)
	return r.db.Select(dest, query)
}

func (r *Relation) Get(dest interface{}, query string) error {
	r.logger.Debug(query)
	return r.db.Get(dest, query)
}

func (r *Relation) Queryx(query string) (*sqlx.Rows, error) {
	r.logger.Debug(query)
	return r.db.Queryx(query)
}

type blocksType struct {
	Type  string
	Count uint64
}

func blockType(bs []blocksType) map[string]uint64 {
	t := make(map[string]uint64)
	for _, b := range bs {
		t[b.Type] = b.Count
	}
	return t
}

func (r *Relation) BatchUpdate(fn func(txn *sqlx.Tx) error) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("tx begin: %s", err)
	}
	if err := fn(tx); err != nil {
		return fmt.Errorf("tx fn: %s", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx commit: %s", err)
	}
	return nil
}

//
//func (r *Relation) AddBlocks(txn *sqlx.Tx, blocks []*types.StateBlock) error {
//	if len(blocks) == 0 {
//		return nil
//	}
//	blksHashes := make([]*blocksHash, 0)
//	//r.logger.Info("batch block count: ", len(blocks))
//	for _, block := range blocks {
//		blksHashes = append(blksHashes, &blocksHash{
//			Hash:      block.GetHash().String(),
//			Type:      block.GetType().String(),
//			Address:   block.GetAddress().String(),
//			Timestamp: block.GetTimestamp(),
//		})
//	}
//	if len(blksHashes) > 0 {
//		if _, err := txn.NamedExec("INSERT INTO BLOCKHASH(hash, type,address,timestamp) VALUES (:hash,:type,:address,:timestamp) ", blksHashes); err != nil {
//			r.logger.Errorf("insert block hash, ", err)
//			return err
//		}
//	}
//	return nil
//}

func blockHash(bs []types.BlockHash) ([]types.Hash, error) {
	hs := make([]types.Hash, 0)
	for _, b := range bs {
		var h types.Hash
		if err := h.Of(b.Hash); err != nil {
			return nil, err
		}
		hs = append(hs, h)
	}
	return hs, nil
}
