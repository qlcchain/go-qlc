package migration

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
)

type MigrationV1ToV15 struct {
}

func (m MigrationV1ToV15) Migrate(store storage.Store) error {
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

func (m MigrationV1ToV15) StartVersion() int {
	return 1
}

func (m MigrationV1ToV15) EndVersion() int {
	return 15
}

type bytesKV struct {
	key   []byte
	value []byte
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
		nep5PendingKvs := make([]*pendingKV, 0)

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

			pi := new(types.PendingInfo)
			if err := pi.Deserialize(value); err != nil {
				return fmt.Errorf("pendingInfo deserialize err: %s", err)
			}
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

			// if pending is from nep5 contract
			if pi.Source == contractaddress.NEP5PledgeAddress {
				nep5PendingKvs = append(nep5PendingKvs, pkv)
			}
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
				for _, pkv := range nep5PendingKvs {
					pKey, err := storage.GetKeyOfParts(storage.KeyPrefixPending, pkv.key)
					if err != nil {
						return err
					}
					if err := batch.Delete(pKey); err != nil {
						return err
					}
				}
				return updateVersion(m, batch)
			})
		} else {
			return store.BatchWrite(false, func(batch storage.Batch) error {
				for _, pkv := range nep5PendingKvs {
					pKey, err := storage.GetKeyOfParts(storage.KeyPrefixPending, pkv.key)
					if err != nil {
						return err
					}
					if err := batch.Delete(pKey); err != nil {
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
