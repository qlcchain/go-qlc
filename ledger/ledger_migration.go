package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type BlockPosterior struct {
}

func (m BlockPosterior) Migrate(txn db.StoreTxn) error {
	v, err := getVersion(txn)
	if err != nil {
		return err
	}
	if m.StartVersion() > int(v) || m.EndVersion() <= int(v) {
		return nil
	}
	err = txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		_, err := blk.UnmarshalMsg(val)
		if err != nil {
			return err
		}
		pre := blk.Root()
		if !pre.IsZero() {
			pKey := getKeyOfHash(pre, idPrefixPosterior)
			if err := txn.Set(pKey, key[1:]); err != nil {
				return err
			}
		}
		return nil
	})
	if err := setVersion(int64(m.EndVersion()), txn); err != nil {
		return err
	}
	return nil
}

func (m BlockPosterior) StartVersion() int {
	return 1
}

func (m BlockPosterior) EndVersion() int {
	return 2
}
