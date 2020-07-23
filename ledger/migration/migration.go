package migration

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
)

type MigrationV1ToV11 struct {
}

func (m MigrationV1ToV11) Migrate(store storage.Store) error {
	return store.BatchWrite(true, func(batch storage.Batch) error {
		if b, err := checkVersion(m, store); err == nil && b {
			fmt.Println("migrate ledger to v11")
			if err := store.Drop(nil); err == nil {
				return updateVersion(m, batch)
			} else {
				return err
			}
		} else {
			return err
		}
	})
}

func (m MigrationV1ToV11) StartVersion() int {
	return 1
}

func (m MigrationV1ToV11) EndVersion() int {
	return 11
}

type MigrationV11ToV12 struct {
}

func (m MigrationV11ToV12) Migrate(store storage.Store) error {
	return store.BatchWrite(true, func(batch storage.Batch) error {
		b, err := checkVersion(m, store)
		if err != nil {
			return err
		}
		if b {
			fmt.Println("migrate ledger v11 to v12 ")
			representMap := make(map[types.Address]*types.Benefit)
			err = batch.Iterator([]byte{byte(storage.KeyPrefixAccount)}, nil, func(key []byte, val []byte) error {
				am := new(types.AccountMeta)
				if err := am.Deserialize(val); err != nil {
					return err
				}
				tm := am.Token(config.ChainToken())
				if tm != nil {
					if _, ok := representMap[tm.Representative]; !ok {
						representMap[tm.Representative] = &types.Benefit{
							Balance: types.ZeroBalance,
							Vote:    types.ZeroBalance,
							Network: types.ZeroBalance,
							Storage: types.ZeroBalance,
							Oracle:  types.ZeroBalance,
							Total:   types.ZeroBalance,
						}
					}
					representMap[tm.Representative].Balance = representMap[tm.Representative].Balance.Add(am.CoinBalance)
					representMap[tm.Representative].Vote = representMap[tm.Representative].Vote.Add(am.CoinVote)
					representMap[tm.Representative].Network = representMap[tm.Representative].Network.Add(am.CoinNetwork)
					representMap[tm.Representative].Total = representMap[tm.Representative].Total.Add(am.VoteWeight())
				}
				return nil
			})
			if err := batch.Drop([]byte{byte(storage.KeyPrefixRepresentation)}); err != nil {
				return err
			}
			for address, benefit := range representMap {
				key, err := storage.GetKeyOfParts(storage.KeyPrefixRepresentation, address)
				if err != nil {
					return err
				}
				val, err := benefit.MarshalMsg(nil)
				if err != nil {
					return err
				}
				if err := batch.Put(key, val); err != nil {
					return err
				}
			}
			// delete cache
			if err := batch.Drop([]byte{byte(storage.KeyPrefixRepresentationCache)}); err != nil {
				return err
			}
			return updateVersion(m, batch)
		}
		return nil
	})
}

func (m MigrationV11ToV12) StartVersion() int {
	return 11
}

func (m MigrationV11ToV12) EndVersion() int {
	return 12
}

type MigrationV12ToV13 struct {
}

func (m MigrationV12ToV13) Migrate(store storage.Store) error {
	return store.BatchWrite(false, func(batch storage.Batch) error {
		b, err := checkVersion(m, store)
		if err != nil {
			return err
		}
		if b {
			fmt.Println("migrate ledger v12 to v13 ")
			var frontiers []*types.Frontier
			prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixFrontier)
			err := store.Iterator(prefix, nil, func(key []byte, val []byte) error {
				var frontier types.Frontier
				copy(frontier.HeaderBlock[:], key[1:])
				copy(frontier.OpenBlock[:], val)
				frontiers = append(frontiers, &frontier)
				return nil
			})
			if err != nil {
				return err
			}
			for _, f := range frontiers {
				k, err := storage.GetKeyOfParts(storage.KeyPrefixFrontier, f.HeaderBlock)
				if err != nil {
					return err
				}
				v, err := f.OpenBlock.Serialize()
				if err != nil {
					return err
				}
				if err := batch.Put(k, v); err != nil {
					return err
				}
			}
			return updateVersion(m, batch)
		}
		return nil
	})
}

func (m MigrationV12ToV13) StartVersion() int {
	return 12
}

func (m MigrationV12ToV13) EndVersion() int {
	return 13
}

type MigrationV13ToV14 struct {
}

func (m MigrationV13ToV14) Migrate(store storage.Store) error {
	return store.BatchWrite(false, func(batch storage.Batch) error {
		b, err := checkVersion(m, store)
		if err != nil {
			return err
		}
		if b {
			fmt.Println("migrate ledger v13 to v14 ")
			if err := store.Upgrade(m.StartVersion()); err != nil {
				return fmt.Errorf("migrate to %d error: %s", m.EndVersion(), err)
			}
			return updateVersion(m, batch)
		}
		return nil
	})
}

func (m MigrationV13ToV14) StartVersion() int {
	return 13
}

func (m MigrationV13ToV14) EndVersion() int {
	return 14
}

type bytesKV struct {
	key   []byte
	value []byte
}

type MigrationV14ToV15 struct {
}

func (m MigrationV14ToV15) Migrate(store storage.Store) error {
	return store.BatchWrite(false, func(batch storage.Batch) error {
		b, err := checkVersion(m, store)
		if err != nil {
			return err
		}
		if b {
			fmt.Println("migrate ledger v14 to v15 ")
			cs := make([]bytesKV, 0)
			prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixTrieVMStorage)
			if err := store.Iterator(prefix, nil, func(k, v []byte) error {
				//fmt.Println("==key ", k)
				key := make([]byte, len(k))
				copy(key, k)
				value := make([]byte, len(v))
				copy(value, v)
				c := bytesKV{
					key:   key,
					value: value,
				}
				cs = append(cs, c)
				return nil
			}); err != nil {
				return err
			}

			for _, c := range cs {
				contractAddrByte := c.key[1 : 1+types.AddressSize]
				contractAddr, err := types.BytesToAddress(contractAddrByte)
				if err != nil {
					return fmt.Errorf("%s is not address ", contractAddrByte)
				}

				newKey := make([]byte, 0)
				newKey = append(newKey, byte(storage.KeyPrefixVMStorage))
				if !contractaddress.IsContractAddress(contractAddr) {
					newKey = append(newKey, contractaddress.SettlementAddress.Bytes()...)
					newKey = append(newKey, c.key[1:]...)
					if err := batch.Put(newKey, c.value); err != nil {
						return err
					}
					fmt.Printf("%s is not contract address \n", contractAddr.String())
				} else {
					newKey = append(newKey, contractAddr.Bytes()...)
					newKey = append(newKey, c.key[1:]...)
					if err := batch.Put(newKey, c.value); err != nil {
						return err
					}
				}
			}
			return updateVersion(m, batch)
		}
		return nil
	})
}

func (m MigrationV14ToV15) StartVersion() int {
	return 14
}

func (m MigrationV14ToV15) EndVersion() int {
	return 15
}

type pendingKV struct {
	key   *types.PendingKey
	value []byte
}

type MigrationV15ToV16 struct {
}

func (m MigrationV15ToV16) Migrate(store storage.Store) error {
	b, err := checkVersion(m, store)
	if err != nil {
		return err
	}
	if b {
		fmt.Println("migrate ledger v15 to v16 ")
		count := 0
		reset := false
		bs := make([]bytesKV, 0)
		pendingKvs := make([]*pendingKV, 0)

		// get all pending infos from db
		prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPending)
		err := store.Iterator(prefix, nil, func(k, v []byte) error {
			key := make([]byte, len(k))
			copy(key, k)
			value := make([]byte, len(v))
			copy(value, v)
			c := bytesKV{
				key:   key,
				value: value,
			}
			bs = append(bs, c)

			pk := new(types.PendingKey)
			if err := pk.Deserialize(key[1:]); err != nil {
				if _, err := pk.UnmarshalMsg(key[1:]); err != nil {
					return fmt.Errorf("pendingKey unmarshalMsg err: %s", err)
				}
				reset = true
			}
			pkv := &pendingKV{
				key:   pk,
				value: value,
			}
			pendingKvs = append(pendingKvs, pkv)
			return nil
		})
		if err != nil {
			return fmt.Errorf("get pendings info err: %s", err)
		}

		if reset {
			// copy all pending infos to another table
			err = store.BatchWrite(false, func(batch storage.Batch) error {
				for _, b := range bs {
					nk := make([]byte, 0)
					nk = append(nk, storage.KeyPrefixPendingBackup)
					nk = append(nk, b.key[1:]...)
					if err := batch.Put(nk, b.value); err != nil {
						return err
					}
					count++
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("backup pending table error: %s", err)
			}

			// reset pending table
			if count != len(pendingKvs) {
				return fmt.Errorf("pending count err: %d, %d", count, len(pendingKvs))
			}
			return store.BatchWrite(false, func(batch storage.Batch) error {
				for _, b := range bs {
					if err := batch.Delete(b.key); err != nil {
						return err
					}
				}

				for _, pkv := range pendingKvs {
					pKey, err := storage.GetKeyOfParts(storage.KeyPrefixPending, pkv.key)
					if err != nil {
						return err
					}
					if err := batch.Put(pKey, pkv.value); err != nil {
						return err
					}
				}
				return updateVersion(m, batch)
			})
		}
	}
	return nil
}

func (m MigrationV15ToV16) StartVersion() int {
	return 15
}

func (m MigrationV15ToV16) EndVersion() int {
	return 16
}

func checkVersion(m Migration, s storage.Store) (bool, error) {
	v, err := getVersion(s)
	if err != nil {
		return false, err
	}
	if int(v) >= m.StartVersion() && int(v) < m.EndVersion() {
		return true, nil
	}
	return false, nil
}

func updateVersion(m Migration, batch storage.Batch) error {
	if err := setVersion(int64(m.EndVersion()), batch); err != nil {
		return err
	}
	//fmt.Printf("update ledger version %d to %d successfully\n ", m.StartVersion(), m.EndVersion())
	return nil
}

func getVersion(s storage.Store) (int64, error) {
	var i int64
	key := []byte{byte(storage.KeyPrefixVersion)}
	val, err := s.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return 0, errors.New("version not found")
		}
		return i, err
	}
	i, _ = binary.Varint(val)
	return i, nil
}

func setVersion(version int64, batch storage.Batch) error {
	key := []byte{byte(storage.KeyPrefixVersion)}
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, version)
	return batch.Put(key, buf[:n])
}
