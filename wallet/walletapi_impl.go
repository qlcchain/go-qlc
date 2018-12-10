/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"go.uber.org/zap"
	"hash"
	"io"
)

const (
	idPrefixId byte = iota
	idPrefixIds
	idPrefixVersion
	idPrefixSeed
	idPrefixIndex
	idPrefixRepresentation
	idPrefixCheck
	idPrefixWork
)

const WalletVersion = 1

type WalletStore struct {
	io.Closer
	db     db.Store
	ledger ledger.Ledger
	log    *zap.SugaredLogger
}

type WalletSession struct {
	db.Store
	ledger   ledger.Ledger
	log      *zap.SugaredLogger
	maxDepth uint64
	walletId []byte
	password []byte // TODO: password fan
}

var (
	EmptyIdErr = errors.New("empty wallet id")
)

func NewWalletStore(config config.Config) *WalletStore {
	return &WalletStore{}
}

func (ws *WalletStore) WalletIds() ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := ws.db.ViewInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixIds}
		return txn.Get(key, func(val []byte, b byte) error {
			err := jsoniter.Unmarshal(val, &ids)
			return err
		})
	})

	return ids, err
}

//NewWallet create new wallet and save to db
// TODO: handle error
func (ws *WalletStore) NewWallet() (uuid.UUID, error) {
	id := uuid.New()

	ids, err := ws.WalletIds()
	if err != nil {
		return id, err
	}
	key, _ := id.MarshalBinary()
	session := ws.NewSession(key)
	err = session.Init()
	if err != nil {
		return id, err
	}
	ids = append(ids, id)
	err = ws.db.UpdateInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixIds}
		bytes, err := jsoniter.Marshal(&ids)
		if err != nil {
			return err
		}
		return txn.Set(key, bytes)
	})
	walletId, _ := id.MarshalBinary()
	err = ws.setCurrentId(walletId)

	return id, err
}

func (ws *WalletStore) CurrentId() (uuid.UUID, error) {
	var id uuid.UUID

	err := ws.db.ViewInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixId}
		return txn.Get(key, func(val []byte, b byte) error {
			if len(val) == 0 {
				return errors.New("can not find any wallet id")
			}
			return id.UnmarshalBinary(val)
		})
	})

	return id, err
}

func (ws *WalletStore) RemoveWallet(id uuid.UUID) error {
	ids, err := ws.WalletIds()
	if err != nil {
		return err
	}

	index := -1

	for i, _id := range ids {
		if id == _id {
			index = i
			break
		}
	}

	if index < 0 {
		return fmt.Errorf("can not find id(%s)", id.String())
	}

	ids = append(ids[:index], ids[index+1:]...)
	walletId, _ := id.MarshalBinary()
	session := ws.NewSession(walletId)
	err = session.Remove()

	if err != nil {
		return err
	}
	var newId []byte
	if len(ids) > 0 {
		newId, _ = ids[0].MarshalBinary()
	} else {
		newId, _ = hex.DecodeString("")
	}
	err = ws.setCurrentId(newId)
	if err != nil {
		return err
	}

	return err
}

func (ws *WalletStore) setCurrentId(walletId []byte) error {
	return ws.db.UpdateInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixId}
		return txn.Set(key, walletId)
	})
}

func (ws *WalletStore) NewSession(walletId []byte) *WalletSession {
	return &WalletSession{
		Store:    ws.db,
		ledger:   ws.ledger,
		log:      common.NewLogger("walletSession" + hex.EncodeToString(walletId)),
		maxDepth: 100,
		walletId: walletId,
		password: []byte{},
	}
}

func (ws *WalletSession) Init() error {
	err := ws.SetDeterministicIndex(1)
	_ = ws.SetVersion(WalletVersion)
	seed, err := types.NewSeed()
	if err != nil {
		return err
	}
	err = ws.SetSeed(seed[:])

	return err
}

//Remove wallet by id
//TODO: Remove all data
func (ws *WalletSession) Remove() error {
	return ws.UpdateInTx(func(txn db.StoreTxn) error {
		seedKey := []byte{idPrefixSeed}
		seedKey = append(seedKey, ws.walletId...)
		err := txn.Delete(seedKey)

		return err
	})
}

func (ws *WalletSession) GetWalletId() ([]byte, error) {
	if len(ws.walletId) == 0 {
		return nil, EmptyIdErr
	}
	return ws.walletId, nil
}

func (ws *WalletSession) GetRepresentative() (types.Address, error) {
	var address types.Address
	err := ws.ViewInTx(func(txn db.StoreTxn) error {

		key := ws.getKey(idPrefixRepresentation)
		return txn.Get(key, func(val []byte, b byte) error {
			addr, err := types.BytesToAddress(val)
			address = addr
			return err
		})
	})

	return address, err
}

func (ws *WalletSession) SetRepresentative(address types.Address) error {
	return ws.UpdateInTx(func(txn db.StoreTxn) error {
		key := ws.getKey(idPrefixRepresentation)
		return txn.Set(key, address[:])
	})
}

func (ws *WalletSession) GetSeed() ([]byte, error) {
	var seed []byte
	err := ws.ViewInTx(func(txn db.StoreTxn) error {

		key := ws.getKey(idPrefixSeed)
		return txn.Get(key, func(val []byte, b byte) error {
			s, err := DecryptSeed(val, []byte(ws.password))
			seed = append(seed, s...)
			return err
		})
	})

	return seed, err
}

func (ws *WalletSession) SetSeed(seed []byte) error {
	encryptSeed, err := EncryptSeed(seed, ws.password)

	if err != nil {
		return err
	}

	return ws.UpdateInTx(func(txn db.StoreTxn) error {
		key := ws.getKey(idPrefixSeed)
		return txn.Set(key, encryptSeed)
	})
}

func (ws *WalletSession) ResetDeterministicIndex() error {
	return ws.SetDeterministicIndex(0)
}

func (ws *WalletSession) GetBalances() ([]*types.AccountMeta, error) {
	index, err := ws.GetDeterministicIndex()

	if err != nil {
		return nil, err
	}
	var accounts []*types.AccountMeta

	if index > 0 {
		seed, err := ws.GetSeed()
		if err != nil {
			return nil, err
		}

		max := max(uint32(index), uint32(ws.maxDepth))
		s := hex.EncodeToString(seed)
		session := ws.ledger.NewLedgerSession(false)
		defer session.Close()

		for i := uint32(0); i < max; i++ {
			key, _, err := types.KeypairFromSeed(s, uint32(i))
			if err != nil {
				ws.log.Fatal(err)
			}
			ac, err := session.GetAccountMeta(types.PubToAddress(key))
			if err != nil {
				ws.log.Fatal(err)
			} else {
				accounts = append(accounts, ac)
			}
		}
	}

	return accounts, nil
}

func (ws *WalletSession) GetBalance(addr types.Address) (*types.AccountMeta, error) {
	index, err := ws.GetDeterministicIndex()

	if err != nil {
		return nil, err
	}

	if index > 0 {
		seed, err := ws.GetSeed()
		if err != nil {
			return nil, err
		}

		max := max(uint32(index), uint32(ws.maxDepth))
		s := hex.EncodeToString(seed)
		session := ws.ledger.NewLedgerSession(false)
		defer session.Close()

		for i := uint32(0); i < max; i++ {
			key, _, err := types.KeypairFromSeed(s, uint32(i))
			if err != nil {
				ws.log.Fatal(err)
			}
			address := types.PubToAddress(key)
			if address == addr {
				ac, err := session.GetAccountMeta(address)
				if err != nil {
					ws.log.Fatal(err)
				} else {
					return ac, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("can not find account(%s) balance", addr.String())
}

func (ws *WalletSession) SearchPending() {
	panic("implement me")
}

func (ws *WalletSession) Open(source types.Hash, token hash.Hash, representative types.Address) (*types.Block, error) {
	panic("implement me")
}

func (ws *WalletSession) Send(source types.Address, token types.Hash, to types.Address, amount types.Amount) (*types.Block, error) {
	panic("implement me")
}

func (ws *WalletSession) Receive(account types.Address, token types.Hash) (*types.Block, error) {
	panic("implement me")
}

func (ws *WalletSession) Change(addr types.Address, representative types.Address) (*types.Block, error) {
	panic("implement me")
}

func (ws *WalletSession) Import(content string, password string) error {
	panic("implement me")
}

func (ws *WalletSession) Export(path string) error {
	panic("implement me")
}

func (ws *WalletSession) GetVersion() (int64, error) {
	var i int64
	err := ws.ViewInTx(func(txn db.StoreTxn) error {

		key := ws.getKey(idPrefixVersion)
		return txn.Get(key, func(val []byte, b byte) error {
			i, _ = binary.Varint(val)
			return nil
		})
	})

	return i, err
}

func (ws *WalletSession) SetVersion(version int64) error {
	return ws.UpdateInTx(func(txn db.StoreTxn) error {
		key := ws.getKey(idPrefixVersion)
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, version)
		return txn.Set(key, buf[:n])
	})
}

func (ws *WalletSession) GetDeterministicIndex() (int64, error) {
	var i int64
	err := ws.ViewInTx(func(txn db.StoreTxn) error {

		key := ws.getKey(idPrefixIndex)
		return txn.Get(key, func(val []byte, b byte) error {
			i, _ = binary.Varint(val)
			return nil
		})
	})

	return i, err
}

func (ws *WalletSession) SetDeterministicIndex(index int64) error {
	return ws.UpdateInTx(func(txn db.StoreTxn) error {
		key := ws.getKey(idPrefixIndex)
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, index)
		return txn.Set(key, buf[:n])
	})
}

func (ws *WalletSession) GetWork() (types.Work, error) {
	panic("implement me")
}

func (ws *WalletSession) SetWork(work types.Work) error {
	panic("implement me")
}

func (ws *WalletSession) IsAccountExist(addr types.Address) bool {
	index, err := ws.GetDeterministicIndex()

	if err != nil {
		return false
	}

	if index > 0 {
		seed, err := ws.GetSeed()
		if err != nil {
			return false
		}

		max := max(uint32(index), uint32(ws.maxDepth))
		s := hex.EncodeToString(seed)
		session := ws.ledger.NewLedgerSession(false)
		defer session.Close()

		for i := uint32(0); i < max; i++ {
			key, _, err := types.KeypairFromSeed(s, uint32(i))
			if err != nil {
				ws.log.Fatal(err)
			}
			address := types.PubToAddress(key)
			if address == addr {
				_, err := session.GetAccountMeta(address)
				if err != nil {
					ws.log.Fatal(err)
				} else {
					return true
				}
			}
		}
	}

	return false
}

func (ws *WalletSession) ValidPassword() bool {
	panic("implement me")
}

func (ws *WalletSession) AttemptPassword(password string) error {
	panic("implement me")
}

func (ws *WalletSession) ChangePassword(password string) error {
	panic("implement me")
}

func (ws *WalletSession) getKey(t byte) []byte {
	var key []byte
	key = append(key, t)
	key = append(key, ws.walletId...)
	return key[:]
}

// max returns the larger of x or y.
func max(x, y uint32) uint32 {
	if x < y {
		return y
	}
	return x
}
