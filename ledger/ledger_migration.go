package ledger

import (
	"encoding/json"
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
		//fmt.Println("migrating ledger v4 to v5 ... ")

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
			key, err := getKeyOfParts(idPrefixRepresentation, k)
			if err != nil {
				return err
			}
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

		ams := make([]*types.AccountMeta, 0)
		err = txn.Iterator(idPrefixAccount, func(key []byte, val []byte, b byte) error {
			am := new(types.AccountMeta)
			_, err := am.UnmarshalMsg(val)
			if err != nil {
				return err
			}
			ams = append(ams, am)
			return nil
		})
		if err != nil {
			return err
		}

		for _, am := range ams {
			tm := am.Token(common.ChainToken())
			if tm != nil && am.CoinBalance.Int == nil {
				am.CoinBalance = tm.Balance
				am.CoinNetwork = types.ZeroBalance
				am.CoinStorage = types.ZeroBalance
				am.CoinOracle = types.ZeroBalance
				am.CoinVote = types.ZeroBalance
				amKey, err := getKeyOfParts(idPrefixAccount, am.Address)
				if err != nil {
					return err
				}
				amBytes, err := am.MarshalMsg(nil)
				if err != nil {
					return err
				}
				if err := txn.Set(amKey, amBytes); err != nil {
					return err
				}
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

type MigrationV5ToV6 struct {
}

func (m MigrationV5ToV6) Migrate(txn db.StoreTxn) error {
	b, err := checkVersion(m, txn)
	if err != nil {
		return err
	}
	if b {
		fmt.Println("migrate ledger v5 to v6")
		newChild := make(map[types.Hash]types.Hash)
		err = txn.Iterator(idPrefixChild, func(key []byte, val []byte, b byte) error {
			children := make(map[types.Hash]int)
			if err := json.Unmarshal(val, &children); err == nil {
				keyHash, err := types.BytesToHash(key[1:])
				if err != nil {
					return err
				}
				if len(children) == 1 {
					for k, v := range children {
						if v == 0 {
							newChild[keyHash] = k
						} else {
							newChild[keyHash] = types.ZeroHash
						}
					}
				} else {
					for k, v := range children {
						if v == 0 {
							newChild[keyHash] = k
						}
					}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		for k, v := range newChild {
			pKey := getKeyOfHash(k, idPrefixChild)
			if v == types.ZeroHash {
				if err := txn.Delete(pKey); err != nil {
					return err
				}
			} else {
				val, err := v.MarshalMsg(nil)
				if err != nil {
					return err
				}
				if err := txn.Set(pKey, val); err != nil {
					return err
				}
			}
		}

		return updateVersion(m, txn)
	}
	return nil
}

func (m MigrationV5ToV6) StartVersion() int {
	return 5
}

func (m MigrationV5ToV6) EndVersion() int {
	return 6
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
	//fmt.Printf("update ledger version %d to %d successfully\n ", m.StartVersion(), m.EndVersion())
	return nil
}

type MigrationV1ToV7 struct {
}

func (m MigrationV1ToV7) Migrate(txn db.StoreTxn) error {
	if b, err := checkVersion(m, txn); err == nil && b {
		fmt.Println("migrate ledger to v7")
		if err := txn.Drop(nil); err == nil {
			return updateVersion(m, txn)
		} else {
			return err
		}
	} else {
		return err
	}
}

func (m MigrationV1ToV7) StartVersion() int {
	return 1
}

func (m MigrationV1ToV7) EndVersion() int {
	return 7
}

type MigrationV7ToV8 struct {
}

func (m MigrationV7ToV8) Migrate(txn db.StoreTxn) error {
	b, err := checkVersion(m, txn)
	if err != nil {
		return err
	}
	if b {
		fmt.Println("migrate ledger v7 to v8 ")
		if err := txn.Drop([]byte{idPrefixBlockCache}); err != nil {
			return err
		}
		if err := txn.Drop([]byte{idPrefixUncheckedBlockPrevious}); err != nil {
			return err
		}
		if err := txn.Drop([]byte{idPrefixUncheckedBlockLink}); err != nil {
			return err
		}
		if err := txn.Drop([]byte{idPrefixUncheckedTokenInfo}); err != nil {
			return err
		}
		return updateVersion(m, txn)
	}
	return nil
}

func (m MigrationV7ToV8) StartVersion() int {
	return 7
}

func (m MigrationV7ToV8) EndVersion() int {
	return 8
}
