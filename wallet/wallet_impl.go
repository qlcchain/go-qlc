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
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
	"io"
)

const (
	idPrefixId byte = iota
	idPrefixIds
	idPrefixVersion
	idPrefixSeed
	idPrefixIndex
	idPrefixRepresentation
	idPrefixWork
)

const (
	Version     = 1
	searchDepth = 100
)

type WalletStore struct {
	io.Closer
	db.Store
	ledger ledger.Ledger
	log    *zap.SugaredLogger
}

type Session struct {
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

func (ws *WalletStore) NewSession(walletId []byte) *Session {
	s := &Session{
		Store:    ws.Store,
		ledger:   ws.ledger,
		log:      log.NewLogger("wallet session" + hex.EncodeToString(walletId)),
		maxDepth: searchDepth,
		walletId: walletId,
		password: []byte{},
	}
	//update database
	err := s.UpdateInTx(func(txn db.StoreTxn) error {
		var migrations []db.Migration
		return txn.Upgrade(migrations)
	})
	if err != nil {
		ws.log.Fatal(err)
	}
	return s
}

func (s *Session) Init() error {
	err := s.SetDeterministicIndex(1)
	if err != nil {
		return err
	}
	_ = s.SetVersion(Version)
	//default password is empty
	_ = s.EnterPassword("")

	seed, err := types.NewSeed()
	if err != nil {
		return err
	}
	err = s.SetSeed(seed[:])

	return err
}

//Remove wallet by id
func (s *Session) Remove() error {
	return s.UpdateInTx(func(txn db.StoreTxn) error {
		for _, val := range []byte{idPrefixId, idPrefixVersion, idPrefixSeed, idPrefixRepresentation} {
			seedKey := []byte{val}
			seedKey = append(seedKey, s.walletId...)
			err := txn.Delete(seedKey)
			if err != nil {
				s.log.Fatal(err)
			}
		}

		return nil
	})
}

func (s *Session) EnterPassword(password string) error {
	s.setPassword(password)
	_, err := s.GetSeed()
	return err
}

func (s *Session) GetWalletId() ([]byte, error) {
	if len(s.walletId) == 0 {
		return nil, EmptyIdErr
	}
	return s.walletId, nil
}

func (s *Session) GetRepresentative() (types.Address, error) {
	var address types.Address
	err := s.ViewInTx(func(txn db.StoreTxn) error {

		key := s.getKey(idPrefixRepresentation)
		return txn.Get(key, func(val []byte, b byte) error {
			addr, err := types.BytesToAddress(val)
			address = addr
			return err
		})
	})

	return address, err
}

func (s *Session) SetRepresentative(address types.Address) error {
	return s.UpdateInTx(func(txn db.StoreTxn) error {
		key := s.getKey(idPrefixRepresentation)
		return txn.Set(key, address[:])
	})
}

func (s *Session) GetSeed() ([]byte, error) {
	var seed []byte
	err := s.ViewInTx(func(txn db.StoreTxn) error {

		key := s.getKey(idPrefixSeed)
		return txn.Get(key, func(val []byte, b byte) error {
			s, err := DecryptSeed(val, s.getPassword())
			seed = append(seed, s...)
			return err
		})
	})

	return seed, err
}

func (s *Session) SetSeed(seed []byte) error {
	encryptSeed, err := EncryptSeed(seed, s.getPassword())

	if err != nil {
		return err
	}

	return s.UpdateInTx(func(txn db.StoreTxn) error {
		key := s.getKey(idPrefixSeed)
		return txn.Set(key, encryptSeed)
	})
}

func (s *Session) ResetDeterministicIndex() error {
	return s.SetDeterministicIndex(0)
}

func (s *Session) GetBalances() ([]*types.AccountMeta, error) {
	index, err := s.GetDeterministicIndex()

	if err != nil {
		return nil, err
	}
	var accounts []*types.AccountMeta

	if index > 0 {
		seed, err := s.GetSeed()
		if err != nil {
			return nil, err
		}

		max := max(uint32(index), uint32(s.maxDepth))
		seedArray := hex.EncodeToString(seed)
		session := s.ledger.NewLedgerSession(false)
		defer session.Close()

		for i := uint32(0); i < max; i++ {
			key, _, err := types.KeypairFromSeed(seedArray, uint32(i))
			if err != nil {
				s.log.Fatal(err)
			}
			ac, err := session.GetAccountMeta(types.PubToAddress(key))
			if err != nil {
				s.log.Fatal(err)
			} else {
				accounts = append(accounts, ac)
			}
		}
	}

	return accounts, nil
}

func (s *Session) GetBalance(addr types.Address) (*types.AccountMeta, error) {
	index, err := s.GetDeterministicIndex()

	if err != nil {
		return nil, err
	}

	if index > 0 {
		seed, err := s.GetSeed()
		if err != nil {
			return nil, err
		}

		max := max(uint32(index), uint32(s.maxDepth))
		seedArray := hex.EncodeToString(seed)
		session := s.ledger.NewLedgerSession(false)
		defer func() {
			err := session.Close()
			if err != nil {
				s.log.Fatal(err)
			}
		}()

		for i := uint32(0); i < max; i++ {
			key, _, err := types.KeypairFromSeed(seedArray, uint32(i))
			if err != nil {
				s.log.Fatal(err)
			}
			address := types.PubToAddress(key)
			if address == addr {
				ac, err := session.GetAccountMeta(address)
				if err != nil {
					s.log.Fatal(err)
				} else {
					return ac, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("can not find account(%s) balance", addr.String())
}

func (s *Session) SearchPending() error {
	panic("implement me")
}

func (s *Session) Send(source types.Address, token types.Hash, to types.Address, amount types.Amount) (*types.Block, error) {
	panic("implement me")
}

func (s *Session) Receive(block types.Hash) (*types.Block, error) {
	panic("implement me")
}

func (s *Session) Change(account types.Address, representative types.Address) (*types.Block, error) {
	if exist := s.IsAccountExist(account); exist {
		session := s.ledger.NewLedgerSession(false)
		defer session.Close()
		//get latest chain token block
		hash := session.Latest(account, common.ChainTokenType)

		if !hash.IsZero() {
			block, err := session.GetBlock(hash)
			if err != nil {
				return nil, err
			}
			if sb, ok := block.(*types.StateBlock); ok {
				newBlock, err := types.NewBlock(types.State)
				if err != nil {
					return nil, err
				}
				if newSb, ok := newBlock.(*types.StateBlock); ok {
					newSb.Balance = sb.GetBalance()
				}
				//newBlock.
				return &newBlock, nil
			}

			return nil, fmt.Errorf("invalid block (%s) of account[%s]", hash, account.String())
		}
	}
	return nil, fmt.Errorf("account[%s] is not exist", account.String())
}

func (s *Session) Import(content string, password string) error {
	panic("implement me")
}

func (s *Session) Export(path string) error {
	panic("implement me")
}

func (s *Session) GetVersion() (int64, error) {
	var i int64
	err := s.ViewInTx(func(txn db.StoreTxn) error {

		key := s.getKey(idPrefixVersion)
		return txn.Get(key, func(val []byte, b byte) error {
			i, _ = binary.Varint(val)
			return nil
		})
	})

	return i, err
}

func (s *Session) SetVersion(version int64) error {
	return s.UpdateInTx(func(txn db.StoreTxn) error {
		key := s.getKey(idPrefixVersion)
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, version)
		return txn.Set(key, buf[:n])
	})
}

func (s *Session) GetDeterministicIndex() (int64, error) {
	var i int64
	err := s.ViewInTx(func(txn db.StoreTxn) error {

		key := s.getKey(idPrefixIndex)
		return txn.Get(key, func(val []byte, b byte) error {
			i, _ = binary.Varint(val)
			return nil
		})
	})

	return i, err
}

func (s *Session) SetDeterministicIndex(index int64) error {
	return s.UpdateInTx(func(txn db.StoreTxn) error {
		key := s.getKey(idPrefixIndex)
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, index)
		return txn.Set(key, buf[:n])
	})
}

func (s *Session) GetWork(hash types.Hash) (types.Work, error) {
	var work types.Work
	err := s.ViewInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixWork}
		key = append(key, hash[:]...)
		return txn.Get(key, func(val []byte, b byte) error {
			return work.UnmarshalBinary(val)
		})
	})

	if err != nil {
		worker, err := types.NewWorker(work, hash)
		work = worker.NewWork()
		if err != nil {
			return work, err
		}

		//cache to db
		_ = s.setWork(hash, work)
	}

	return work, nil
}

func (s *Session) setWork(hash types.Hash, work types.Work) error {
	return s.UpdateInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixWork}
		key = append(key, hash[:]...)
		buf := make([]byte, work.Len())
		err := work.MarshalBinaryTo(buf)
		if err != nil {
			return err
		}
		return txn.Set(key, buf)
	})
}

func (s *Session) IsAccountExist(addr types.Address) bool {
	index, err := s.GetDeterministicIndex()

	if err != nil {
		return false
	}

	if index > 0 {
		seed, err := s.GetSeed()
		if err != nil {
			return false
		}

		max := max(uint32(index), uint32(s.maxDepth))
		seedArray := hex.EncodeToString(seed)
		session := s.ledger.NewLedgerSession(false)
		defer session.Close()

		for i := uint32(0); i < max; i++ {
			key, _, err := types.KeypairFromSeed(seedArray, uint32(i))
			if err != nil {
				s.log.Fatal(err)
			}
			address := types.PubToAddress(key)
			if address == addr {
				_, err := session.GetAccountMeta(address)
				if err != nil {
					s.log.Fatal(err)
				} else {
					return true
				}
			}
		}
	}

	return false
}

func (s *Session) ValidPassword() bool {
	_, err := s.GetSeed()
	return err == nil
}

func (s *Session) ChangePassword(password string) error {
	seed, err := s.GetSeed()
	if err != nil {
		return nil
	}
	//set new password
	s.setPassword(password)
	return s.SetSeed(seed)
}

func (s *Session) getKey(t byte) []byte {
	var key []byte
	key = append(key, t)
	key = append(key, s.walletId...)
	return key[:]
}

func (s *Session) getPassword() []byte {
	return s.password
}

//TODO: implement password fan
func (s *Session) setPassword(password string) {
	s.password = []byte(password)
}

// max returns the larger of x or y.
func max(x, y uint32) uint32 {
	if x < y {
		return y
	}
	return x
}
