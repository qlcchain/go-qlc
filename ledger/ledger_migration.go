package ledger

import (
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
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

type MigrationV3ToV4 struct {
}

func (m MigrationV3ToV4) Migrate(txn db.StoreTxn) error {
	b, err := checkVersion(m, txn)
	if err != nil {
		return err
	}
	if b {
		fmt.Println("migrating ledger v3 to v4 ... ")
		deleteTable := []byte{idPrefixSender, idPrefixReceiver, idPrefixMessage}
		for _, d := range deleteTable {
			prefix := []byte{d}
			err := txn.Stream(prefix, func(item *badger.Item) bool {
				return true
			}, func(list *pb.KVList) error {
				for _, l := range list.Kv {
					if err := txn.Delete(l.Key); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return updateVersion(m, txn)
	}
	return nil
}

func (m MigrationV3ToV4) StartVersion() int {
	return 3
}

func (m MigrationV3ToV4) EndVersion() int {
	return 4
}

type MigrationV4ToV5 struct {
}

func (m MigrationV4ToV5) Migrate(txn db.StoreTxn) error {
	b, err := checkVersion(m, txn)
	if err != nil {
		return err
	}

	if b {
		fmt.Println("migrating ledger v4 to v5 ... ")

		bas := make(map[types.Address]types.Balance)
		err = txn.Iterator(idPrefixRepresentation, func(key []byte, val []byte, b byte) error {
			var amount types.Balance
			if err := amount.UnmarshalText(val); err == nil {
				address, err := types.BytesToAddress(key[1:])
				if err != nil {
					return err
				}
				bas[address] = amount
			}
			return nil
		})
		if err != nil {
			return err
		}
		for k, v := range bas {
			key := getRepresentationKey(k)
			benefit := new(types.Benefit)
			benefit.Balance = v
			benefit.Storage = types.ZeroBalance
			benefit.Vote = types.ZeroBalance
			benefit.Oracle = types.ZeroBalance
			benefit.Network = types.ZeroBalance
			benefit.Total = v
			val2, err := benefit.MarshalMsg(nil)
			if err != nil {
				return err
			}
			if err := txn.Set(key, val2); err != nil {
				return err
			}
		}
		return updateVersion(m, txn)
	}
	return nil
}

func (m MigrationV4ToV5) StartVersion() int {
	return 4
}

func (m MigrationV4ToV5) EndVersion() int {
	return 5
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
	if err := setVersion(int64(m.EndVersion()), txn); err != nil {
		return err
	}
	fmt.Printf("update ledger version %d to %d successfully\n ", m.StartVersion(), m.EndVersion())
	return nil
}
