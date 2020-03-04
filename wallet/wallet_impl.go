/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/monitor"
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
	Version            = 1
	searchAccountCount = 100
)

type WalletStore struct {
	io.Closer
	storage.Store
	dir string
	//ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

type Session struct {
	storage.Store
	lock            sync.RWMutex
	logger          *zap.SugaredLogger
	maxAccountCount uint64
	walletId        []byte
	password        *crypto.SecureString
}

var (
	ErrEmptyId = errors.New("empty wallet id")
)

func (ws *WalletStore) NewSession(walletId types.Address) *Session {
	s := &Session{
		Store:           ws.Store,
		logger:          log.NewLogger("wallet session: " + walletId.String()),
		maxAccountCount: searchAccountCount,
		walletId:        walletId.Bytes(),
	}
	return s
}

func (s *Session) removeWallet(batch storage.Batch) error {
	for _, val := range []byte{idPrefixId, idPrefixVersion, idPrefixSeed, idPrefixRepresentation} {
		key := []byte{val}
		key = append(key, s.walletId...)
		err := batch.Delete(key)
		if err != nil {
			s.logger.Fatal(err)
		}
	}

	return nil
}

func (s *Session) EnterPassword(password string) error {
	s.setPassword(password)
	seed, err := s.GetSeed()
	if err != nil {
		return err
	}
	if len(seed) == 0 {
		return nil
	}
	return fmt.Errorf("already have encrypt seed")
}

func (s *Session) VerifyPassword(password string) (bool, error) {
	s.setPassword(password)
	seed, err := s.GetSeed()
	if err != nil {
		return false, err
	}
	if len(seed) == 0 {
		return false, fmt.Errorf("password is invalid")
	}
	return true, nil
}

func (s *Session) GetWalletId() ([]byte, error) {
	if len(s.walletId) == 0 {
		return nil, ErrEmptyId
	}
	return s.walletId, nil
}

func (s *Session) GetRepresentative() (types.Address, error) {
	var address types.Address
	key := s.getKey(idPrefixRepresentation)
	val, err := s.Get(key)
	if err != nil {
		return address, err
	}
	addr, err := types.BytesToAddress(val)
	address = addr
	return address, err
}

func (s *Session) SetRepresentative(address types.Address) error {
	key := s.getKey(idPrefixRepresentation)
	return s.Put(key, address[:])
}

func (s *Session) GetSeed() ([]byte, error) {
	defer monitor.Duration(time.Now(), "wallet.getseed")

	var seed []byte
	pw := s.getPassword()
	tmp := make([]byte, len(pw))
	copy(tmp, pw)
	key := s.getKey(idPrefixSeed)
	val, err := s.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return seed, nil
		} else {
			return nil, err
		}
	}
	bytes, err := util.DecryptBytes(val, tmp)
	if err != nil {
		return nil, err
	}
	seed = append(seed, bytes...)
	return seed, nil
}

func (s *Session) setSeed(seed []byte) error {
	return s.BatchWrite(true, func(batch storage.Batch) error {
		return s.setSeedByTxn(batch, seed)
	})
}

func (s *Session) setSeedByTxn(batch storage.Batch, seed []byte) error {
	encryptSeed, err := util.EncryptBytes(seed, s.getPassword())

	if err != nil {
		return err
	}

	key := s.getKey(idPrefixSeed)
	return batch.Put(key, encryptSeed)
}

func (s *Session) ResetDeterministicIndex() error {
	return s.SetDeterministicIndex(0)
}

func (s *Session) GetVersion() (int64, error) {
	var i int64
	key := s.getKey(idPrefixVersion)
	val, err := s.Get(key)
	if err != nil {
		return i, err
	}
	i, _ = binary.Varint(val)
	return i, nil
}

func (s *Session) SetVersion(version int64) error {
	return s.BatchWrite(true, func(batch storage.Batch) error {
		return s.setVersion(batch, version)
	})
}

func (s *Session) setVersion(batch storage.Batch, version int64) error {
	key := s.getKey(idPrefixVersion)
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, version)
	return batch.Put(key, buf[:n])
}

func (s *Session) GetDeterministicIndex() (int64, error) {
	var i int64
	key := s.getKey(idPrefixIndex)
	val, err := s.Get(key)
	if err != nil {
		return i, err
	}
	i, _ = binary.Varint(val)
	return i, nil
}

func (s *Session) SetDeterministicIndex(index int64) error {
	return s.BatchWrite(true, func(batch storage.Batch) error {
		return s.setDeterministicIndex(batch, index)
	})
}

func (s *Session) setDeterministicIndex(batch storage.Batch, index int64) error {
	key := s.getKey(idPrefixIndex)
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, index)
	return batch.Put(key, buf[:n])
}

func (s *Session) GetWork(account types.Address) (types.Work, error) {
	var work types.Work
	key := []byte{idPrefixWork}
	key = append(key, account[:]...)
	val, err := s.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return work, nil
		} else {
			return work, err
		}
	}
	if err := work.UnmarshalBinary(val); err != nil {
		return work, err
	}
	return work, nil
}

func (s *Session) generateWork(hash types.Hash) types.Work {
	var work types.Work
	worker, _ := types.NewWorker(work, hash)
	return worker.NewWork()
	//
	////cache to db
	//_ = s.setWork(hash, work)
}

func (s *Session) setWork(account types.Address, work types.Work) error {
	key := []byte{idPrefixWork}
	key = append(key, account[:]...)
	buf := make([]byte, work.Len())
	err := work.MarshalBinaryTo(buf)
	if err != nil {
		return err
	}
	return s.Put(key, buf)
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
	return s.setSeed(seed)
}

func (s *Session) GetRawKey(account types.Address) (*types.Account, error) {
	index, err := s.GetDeterministicIndex()
	if err != nil {
		index = 0
	}

	seedArray, err := s.GetSeed()
	if err != nil {
		return nil, err
	}

	max := util.UInt32Max(uint32(index), uint32(s.maxAccountCount))
	seed, _ := types.BytesToSeed(seedArray)

	for i := uint32(0); i < max; i++ {
		a, err := seed.Account(uint32(i))
		if err != nil {
			s.logger.Fatal(err)
		}
		address := a.Address()
		if address == account {
			return a, nil
		}
	}

	return nil, fmt.Errorf("can not fetch account[%s]'s raw key", account.String())
}

func (s *Session) getKey(t byte) []byte {
	var key []byte
	key = append(key, t)
	key = append(key, s.walletId...)
	return key[:]
}

func (s *Session) getPassword() []byte {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.password == nil {
		return []byte{}
	}
	return s.password.Bytes()
}

func (s *Session) setPassword(password string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.password != nil {
		s.password.Destroy()
	}
	if len(password) > 0 {
		var err error
		s.password, err = crypto.NewSecureString(password)
		if err != nil {
			s.logger.Error(err)
		}
	}
}
