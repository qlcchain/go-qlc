/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	cabi "github.com/qlcchain/go-qlc/vm/abi/contract"
)

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func getStorageKey(addr *types.Address, k []byte) []byte {
	var storageKey []byte
	storageKey = append(storageKey, []byte{idPrefixStorage}...)
	if addr != nil {
		storageKey = append(storageKey, addr[:]...)
	}
	storageKey = append(storageKey, k...)
	return storageKey[:]
}

func (l *Ledger) GetStorage(addr *types.Address, key []byte, txns ...db.StoreTxn) ([]byte, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	storageKey := getStorageKey(addr, key)
	var storage []byte
	err := txn.Get(storageKey, func(val []byte, b byte) (err error) {
		storage = val
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}
	return storage, nil
}

func (l *Ledger) SetStorage(key []byte, value []byte, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)
	storageKey := getStorageKey(nil, key)
	err := txn.Get(storageKey, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrStorageExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(storageKey, value)
}

func (l *Ledger) ListTokens(txns ...db.StoreTxn) ([]*types.TokenInfo, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var tokens []*types.TokenInfo
	err := txn.Iterator(idPrefixToken, func(key []byte, val []byte, b byte) error {
		token := new(types.TokenInfo)
		if err := json.Unmarshal(val, token); err != nil {
			fmt.Println(err)
			return err
		}
		tokens = append(tokens, token)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (l *Ledger) GetTokenById(tokenId types.Hash, txns ...db.StoreTxn) (*types.TokenInfo, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key := getKey(tokenId, idPrefixToken)
	token := new(types.TokenInfo)
	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if err := json.Unmarshal(val, token); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	return token, nil
}

func (l *Ledger) GetTokenByName(tokenName string, txns ...db.StoreTxn) (*types.TokenInfo, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var token *types.TokenInfo
	err := txn.Iterator(idPrefixToken, func(key []byte, val []byte, b byte) error {
		t := new(types.TokenInfo)
		if err := json.Unmarshal(val, t); err != nil {
			return err
		}
		if t.TokenName == tokenName {
			token = t
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (l *Ledger) GetGenesis(txns ...db.StoreTxn) ([]*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var blocks []*types.StateBlock
	err := txn.Iterator(idPrefixToken, func(key []byte, val []byte, b byte) error {
		t := new(types.TokenInfo)
		if err := json.Unmarshal(val, t); err != nil {
			return err
		}
		tm, err := l.GetTokenMeta(t.Owner, t.TokenId, txn)
		if err != nil {
			return err
		}
		block, err := l.GetStateBlock(tm.OpenBlock, txn)
		if err != nil {
			return err
		}
		blocks = append(blocks, block)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (l *Ledger) testUnpack() {
	block := types.StateBlock{}
	_, err := cabi.ParseTokenInfo(block.Data)
	if err != nil {
		l.logger.Error(err)
	}
}
