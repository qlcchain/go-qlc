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

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/test/mock"
	"go.uber.org/zap"
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
	db.Store
	dir    string
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

type Session struct {
	db.Store
	lock            sync.RWMutex
	ledger          *ledger.Ledger
	logger          *zap.SugaredLogger
	maxAccountCount uint64
	walletId        []byte
	password        *crypto.SecureString
}

var (
	EmptyIdErr = errors.New("empty wallet id")
)

func (ws *WalletStore) NewSession(walletId types.Address) *Session {
	s := &Session{
		Store:           ws.Store,
		ledger:          ws.ledger,
		logger:          log.NewLogger("wallet session: " + walletId.String()),
		maxAccountCount: searchAccountCount,
		walletId:        walletId.Bytes(),
	}
	//update database
	err := s.UpdateInTx(func(txn db.StoreTxn) error {
		var migrations []db.Migration
		return txn.Upgrade(migrations)
	})
	if err != nil {
		ws.logger.Error(err)
	}
	return s
}

func (s *Session) removeWallet(txn db.StoreTxn) error {
	for _, val := range []byte{idPrefixId, idPrefixVersion, idPrefixSeed, idPrefixRepresentation} {
		key := []byte{val}
		key = append(key, s.walletId...)
		err := txn.Delete(key)
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
			s, err := util.DecryptBytes(val, s.getPassword())
			seed = append(seed, s...)
			return err
		})
	})

	if err == badger.ErrKeyNotFound {
		err = nil
	}

	return seed, err
}

func (s *Session) setSeed(seed []byte) error {
	return s.UpdateInTx(func(txn db.StoreTxn) error {
		return s.setSeedByTxn(txn, seed)
	})
}

func (s *Session) setSeedByTxn(txn db.StoreTxn, seed []byte) error {
	encryptSeed, err := util.EncryptBytes(seed, s.getPassword())

	if err != nil {
		return err
	}

	key := s.getKey(idPrefixSeed)
	return txn.Set(key, encryptSeed)
}

func (s *Session) ResetDeterministicIndex() error {
	return s.SetDeterministicIndex(0)
}

func (s *Session) GetBalances() (map[types.Hash]types.Balance, error) {
	cache := map[types.Hash]types.Balance{}

	l := s.ledger
	accounts, err := s.GetAccounts()

	if err != nil {
		return cache, err
	}

	for _, account := range accounts {
		if am, err := l.GetAccountMeta(account); err == nil {
			for _, tm := range am.Tokens {
				if balance, ok := cache[tm.Type]; ok {
					//b := cache[tm.Type]
					cache[tm.Type] = balance.Add(tm.Balance)
				} else {
					cache[tm.Type] = tm.Balance
				}
			}
		}
	}

	return cache, nil
}

func (s *Session) SearchPending() error {
	l := s.ledger

	accounts, err := s.GetAccounts()

	if err != nil {
		return err
	}
	for _, account := range accounts {
		if keys, err := l.Pending(account); err == nil {
			for _, key := range keys {
				if block, err := l.GetStateBlock(key.Hash); err == nil {
					//TODO: implement
					s.logger.Debug(block)
					//_, _ = s.Receive(block)
				}
			}
		}
	}

	return nil
}

func (s *Session) GenerateSendBlock(source types.Address, token types.Hash, to types.Address, amount types.Balance) (types.Block, error) {
	acc, err := s.GetRawKey(source)
	if err != nil {
		return nil, err
	}

	l := s.ledger
	tm, err := l.GetTokenMeta(source, token)
	if err != nil {
		return nil, err
	}
	balance, err := l.TokenBalance(source, token)
	if err != nil {
		return nil, err
	}

	if balance.Compare(amount) == types.BalanceCompBigger {
		newBalance := balance.Sub(amount)
		sendBlock, _ := types.NewBlock(types.State)

		if sb, ok := sendBlock.(*types.StateBlock); ok {
			sb.Address = source
			sb.Token = token
			sb.Link = to.ToHash()
			sb.Balance = newBalance
			sb.Previous = tm.Header
			sb.Representative = tm.Representative
			sb.Work, _ = s.GetWork(source)
			sb.Signature = acc.Sign(sb.GetHash())
			if !sb.IsValid() {
				sb.Work = s.generateWork(sb.Root())
			}
		}
		return sendBlock, nil
	} else {
		return nil, fmt.Errorf("not enought balance(%s) of %s", balance, amount)
	}
}

func (s *Session) GenerateReceiveBlock(sendBlock types.Block) (types.Block, error) {
	hash := sendBlock.GetHash()
	var state *types.StateBlock
	ok := false
	if state, ok = sendBlock.(*types.StateBlock); !ok {
		return nil, fmt.Errorf("invalid state sendBlock(%s)", hash.String())
	}

	l := s.ledger

	// block not exist
	if exist, err := l.HasStateBlock(hash); !exist || err != nil {
		return nil, fmt.Errorf("sendBlock(%s) does not exist", hash.String())
	}
	sendTm, err := l.Token(hash)
	if err != nil {
		return nil, err
	}
	rxAccount := types.Address(state.Link)

	acc, err := s.GetRawKey(rxAccount)
	if err != nil {
		return nil, err
	}
	info, err := l.GetPending(types.PendingKey{Address: rxAccount, Hash: hash})
	if err != nil {
		return nil, err
	}
	receiveBlock, _ := types.NewBlock(types.State)
	has, err := l.HasAccountMeta(rxAccount)
	if err != nil {
		return nil, err
	}
	if has {
		rxAm, err := l.GetAccountMeta(rxAccount)
		if err != nil {
			return nil, err
		}
		rxTm := rxAm.Token(state.Token)
		if sb, ok := receiveBlock.(*types.StateBlock); ok {
			sb.Address = rxAccount
			sb.Balance = rxTm.Balance.Add(info.Amount)
			sb.Previous = rxTm.Header
			sb.Link = hash
			sb.Representative = rxTm.Representative
			sb.Token = rxTm.Type
			sb.Extra = types.Hash{}
			sb.Work, _ = s.GetWork(rxAccount)
			sb.Signature = acc.Sign(sb.GetHash())
			if !sb.IsValid() {
				sb.Work = s.generateWork(sb.Root())
			}
		}
	} else {
		if sb, ok := receiveBlock.(*types.StateBlock); ok {
			sb.Address = rxAccount
			sb.Balance = info.Amount
			sb.Previous = types.Hash{}
			sb.Link = hash
			sb.Representative = sendTm.Representative
			sb.Token = sendTm.Type
			sb.Extra = types.Hash{}
			sb.Work, _ = s.GetWork(rxAccount)
			sb.Signature = acc.Sign(sb.GetHash())
			if !sb.IsValid() {
				sb.Work = s.generateWork(sb.Root())
			}
		}
	}
	return receiveBlock, nil
}

func (s *Session) GenerateChangeBlock(account types.Address, representative types.Address) (types.Block, error) {
	if exist := s.IsAccountExist(account); !exist {
		return nil, fmt.Errorf("account[%s] is not exist", account.String())
	}

	l := s.ledger
	if _, err := l.GetAccountMeta(representative); err != nil {
		return nil, fmt.Errorf("invalid representative[%s]", representative.String())
	}

	//get latest chain token block
	hash := l.Latest(account, mock.GetChainTokenType())

	if hash.IsZero() {
		return nil, fmt.Errorf("account [%s] does not have the main chain account", account.String())
	}

	block, err := l.GetStateBlock(hash)
	if err != nil {
		return nil, err
	}
	changeBlock, err := types.NewBlock(types.State)
	if err != nil {
		return nil, err
	}
	tm, err := l.GetTokenMeta(account, mock.GetChainTokenType())
	if newSb, ok := changeBlock.(*types.StateBlock); ok {
		acc, err := s.GetRawKey(account)
		if err != nil {
			return nil, err
		}
		newSb.Address = account
		newSb.Balance = tm.Balance
		newSb.Previous = tm.Header
		newSb.Link = account.ToHash()
		newSb.Representative = representative
		newSb.Token = block.Token
		newSb.Extra = types.Hash{}
		newSb.Work, _ = s.GetWork(account)
		newSb.Signature = acc.Sign(newSb.GetHash())
		if !newSb.IsValid() {
			newSb.Work = s.generateWork(newSb.Root())
			_ = s.setWork(account, newSb.Work)
		}
	}
	return changeBlock, nil

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
		return s.setVersion(txn, version)
	})
}

func (s *Session) setVersion(txn db.StoreTxn, version int64) error {
	key := s.getKey(idPrefixVersion)
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, version)
	return txn.Set(key, buf[:n])
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
		return s.setDeterministicIndex(txn, index)
	})
}

func (s *Session) setDeterministicIndex(txn db.StoreTxn, index int64) error {
	key := s.getKey(idPrefixIndex)
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, index)
	return txn.Set(key, buf[:n])
}

func (s *Session) GetWork(account types.Address) (types.Work, error) {
	var work types.Work
	err := s.ViewInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixWork}
		key = append(key, account[:]...)
		return txn.Get(key, func(val []byte, b byte) error {
			return work.UnmarshalBinary(val)
		})
	})

	if err != nil && err == badger.ErrKeyNotFound {
		err = nil
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
	return s.UpdateInTx(func(txn db.StoreTxn) error {
		key := []byte{idPrefixWork}
		key = append(key, account[:]...)
		buf := make([]byte, work.Len())
		err := work.MarshalBinaryTo(buf)
		if err != nil {
			return err
		}
		return txn.Set(key, buf)
	})
}

func (s *Session) IsAccountExist(addr types.Address) bool {
	if am, err := s.ledger.GetAccountMeta(addr); err == nil && am.Address == addr {
		_, err := s.GetRawKey(addr)
		return err == nil
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

	max := max(uint32(index), uint32(s.maxAccountCount))
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

func (s *Session) GetAccounts() ([]types.Address, error) {
	l := s.ledger
	var accounts []types.Address
	index, err := s.GetDeterministicIndex()
	if err != nil {
		index = 0
	}

	if seedArray, err := s.GetSeed(); err == nil {
		max := max(uint32(index), uint32(s.maxAccountCount))
		s, err := types.BytesToSeed(seedArray)
		if err != nil {
			return accounts, err
		}
		for i := uint32(0); i < max; i++ {
			if account, err := s.Account(uint32(i)); err == nil {
				address := account.Address()
				if _, err := l.GetAccountMeta(address); err == nil {
					accounts = append(accounts, address)
				} else {
					//s.logger.Error(err)
				}
			}
		}
	} else {
		return nil, err
	}

	return accounts, nil
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
	s.password, _ = crypto.NewSecureString(password)
}

// max returns the larger of x or y.
func max(x, y uint32) uint32 {
	if x < y {
		return y
	}
	return x
}

func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
