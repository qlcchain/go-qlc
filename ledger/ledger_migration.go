package ledger

import (
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type MigrationV1ToV2 struct {
}

func (m MigrationV1ToV2) Migrate(txn db.StoreTxn) error {
	b, err := checkVersion(m, txn)
	if err != nil {
		return err
	}
	if b {
		fmt.Println("migrate ledger v1 to v2 ")
		err = txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
			blk := new(types.StateBlock)
			_, err := blk.UnmarshalMsg(val)
			if err != nil {
				return err
			}
			addChild(blk, txn)
			return nil
		})
		if err != nil {
			return err
		}
		return updateVersion(m, txn)
	}
	return nil
}

func (m MigrationV1ToV2) StartVersion() int {
	return 1
}

func (m MigrationV1ToV2) EndVersion() int {
	return 2
}

type MigrationV2ToV3 struct {
}

func (m MigrationV2ToV3) Migrate(txn db.StoreTxn) error {
	b, err := checkVersion(m, txn)
	if err != nil {
		return err
	}
	if b {
		fmt.Println("migrate ledger v2 to v3 ")
		key := getKeyOfHash(common.GenesisBlockHash(), idPrefixBlock)
		err := txn.Get(key, func(bytes []byte, b byte) error {
			return nil
		})
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == badger.ErrKeyNotFound {
			if err := txn.Drop(nil); err != nil {
				return err
			}
		}
		return updateVersion(m, txn)
	}
	return nil
}

func (m MigrationV2ToV3) StartVersion() int {
	return 2
}

func (m MigrationV2ToV3) EndVersion() int {
	return 3
}

func checkVersion(m db.Migration, txn db.StoreTxn) (bool, error) {
	v, err := getVersion(txn)
	if err != nil {
		return false, err
	}
	if int(v) >= m.StartVersion() && int(v) < m.EndVersion() {
		return true, nil
	}
	return false, nil
}

func updateVersion(m db.Migration, txn db.StoreTxn) error {
	fmt.Printf("update ledger version %d to %d\n", m.StartVersion(), m.EndVersion())
	if err := setVersion(int64(m.EndVersion()), txn); err != nil {
		return err
	}
	return nil
}
