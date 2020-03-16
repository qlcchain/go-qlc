package relation

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/qlcchain/go-qlc/common/types"
)

func (r *Relation) count(s types.Schema) (uint64, error) {
	sql := fmt.Sprintf("select count (*) as total from %s", s.TableName())
	r.logger.Debug(sql)
	var i uint64
	err := r.db.Store().Get(&i, sql)
	if err != nil {
		return 0, fmt.Errorf("count error, sql: %s, err: %s", sql, err.Error())
	}
	return i, nil
}

func (r *Relation) Blocks(limit int, offset int) ([]types.Hash, error) {
	block := new(types.StateBlock)
	var h []blocksHash
	sql := fmt.Sprintf("select * from %s order by timestamp desc, type desc limit %d offset %d", block.TableName(), limit, offset)
	err := r.db.Store().Select(&h, sql)
	if err != nil {
		return nil, fmt.Errorf("read error, sql: %s, err: %s", sql, err.Error())
	}
	return blockHash(h)
}

func (r *Relation) BlocksByAccount(address types.Address, limit int, offset int) ([]types.Hash, error) {
	block := new(types.StateBlock)
	var h []blocksHash
	sql := fmt.Sprintf("select * from %s where address = '%s' order by timestamp desc, type desc limit %d offset %d", block.TableName(), address.String(), limit, offset)
	err := r.db.Store().Select(&h, sql)
	if err != nil {
		return nil, fmt.Errorf("read error, sql: %s, err: %s", sql, err.Error())
	}
	return blockHash(h)
}

func (r *Relation) BlocksCount() (uint64, error) {
	return r.count(new(types.StateBlock))
}

func (r *Relation) BlocksCountByType() (map[string]uint64, error) {
	block := new(types.StateBlock)
	var t []blocksType
	sql := fmt.Sprintf("select type, count(*) as count from %s  group by type", block.TableName())
	r.logger.Debug(sql)
	err := r.db.Store().Select(&t, sql)
	if err != nil {
		return nil, fmt.Errorf("group error, sql: %s, err: %s", sql, err.Error())
	}
	return blockType(t), nil
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
	tx := r.db.Store().MustBegin()
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

type blocksHash struct {
	Id        int64
	Hash      string
	Type      string
	Address   string
	Timestamp int64
}

func blockHash(bs []blocksHash) ([]types.Hash, error) {
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
