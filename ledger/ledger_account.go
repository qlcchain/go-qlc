package ledger

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddAccountMeta(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrAccountExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetAccountMeta(key types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	am, err := l.GetAccountMetaCache(key)
	if err != nil {
		am = nil
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, key)
	if err != nil {
		return nil, err
	}
	meta := new(types.AccountMeta)
	err = txn.Get(k, func(v []byte, b byte) (err error) {
		if err = meta.Deserialize(v); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		meta = nil
	}
	if am != nil && meta == nil {
		return am, nil
	}
	if am == nil && meta != nil {
		return meta, nil
	}
	if am != nil && meta != nil {
		for _, v := range meta.Tokens {
			temp := am.Token(v.Type)
			if temp != nil {
				if temp.BlockCount < v.BlockCount {
					temp.BlockCount = v.BlockCount
					temp.Type = v.Type
					temp.BelongTo = v.BelongTo
					temp.Header = v.Header
					temp.Balance = v.Balance
					temp.Modified = v.Modified
					temp.OpenBlock = v.OpenBlock
					temp.Representative = v.Representative
				}
			} else {
				am.Tokens = append(am.Tokens, v)
			}
		}
		return am, nil
	}
	return nil, ErrAccountNotFound
}

func (l *Ledger) GetAccountMetaConfirmed(key types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, key)
	if err != nil {
		return nil, err
	}

	value := new(types.AccountMeta)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetAccountMetas(fn func(am *types.AccountMeta) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixAccount, func(key []byte, val []byte, b byte) error {
		am := new(types.AccountMeta)
		if err := am.Deserialize(val); err != nil {
			return err
		}
		if err := fn(am); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) CountAccountMetas(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Count([]byte{idPrefixAccount})
}

func (l *Ledger) UpdateAccountMeta(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrAccountNotFound
		}
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) AddOrUpdateAccountMeta(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) DeleteAccountMeta(key types.Address, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, key)
	if err != nil {
		return err
	}

	return txn.Delete(k)
}

func (l *Ledger) HasAccountMeta(key types.Address, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) AddTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		return err
	}

	if am.Token(meta.Type) != nil {
		return ErrTokenExists
	}

	am.Tokens = append(am.Tokens, meta)
	return l.UpdateAccountMeta(am, txns...)
}

func (l *Ledger) GetTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) (*types.TokenMeta, error) {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		return nil, err
	}

	tm := am.Token(tokenType)
	if tm == nil {
		return nil, ErrTokenNotFound
	}

	return tm, nil
}

func (l *Ledger) GetTokenMetaConfirmed(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) (*types.TokenMeta, error) {
	am, err := l.GetAccountMetaConfirmed(address, txns...)
	if err != nil {
		return nil, err
	}

	tm := am.Token(tokenType)
	if tm == nil {
		return nil, ErrTokenNotFound
	}

	return tm, nil
}

func (l *Ledger) UpdateTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		return err
	}

	//tm := am.Token(meta.Type)
	//
	//if tm != nil {
	//	tm = meta
	//	return l.UpdateAccountMeta(am, txns...)
	//}
	//tokens := am.Tokens
	for index, token := range am.Tokens {
		if token.Type == meta.Type {
			//am.Tokens = append(tokens[:index], tokens[index+1:]...)
			//am.Tokens = append(am.Tokens, meta)
			am.Tokens[index] = meta
			return l.UpdateAccountMeta(am, txns...)
		}
	}
	return ErrTokenNotFound
}

func (l *Ledger) AddOrUpdateTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range am.Tokens {
		if token.Type == meta.Type {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
			am.Tokens = append(am.Tokens, meta)
			return l.UpdateAccountMeta(am, txns...)
		}
	}

	am.Tokens = append(am.Tokens, meta)
	return l.UpdateAccountMeta(am, txns...)
}

func (l *Ledger) DeleteTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMetaConfirmed(address, txns...)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range tokens {
		if token.Type == tokenType {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}
	return l.UpdateAccountMeta(am, txns...)
}

func (l *Ledger) DeleteTokenMetaCache(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMetaCache(address, txns...)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range tokens {
		if token.Type == tokenType {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}
	return l.UpdateAccountMetaCache(am, txns...)
}

func (l *Ledger) HasTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) (bool, error) {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		if err == ErrAccountNotFound {
			return false, nil
		}
		return false, err
	}
	for _, t := range am.Tokens {
		if t.Type == tokenType {
			return true, nil
		}
	}
	return false, nil
}

func (l *Ledger) AddAccountMetaCache(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrAccountExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetAccountMetaCache(key types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, key)
	if err != nil {
		return nil, err
	}
	value := new(types.AccountMeta)
	err = txn.Get(k, func(v []byte, b byte) (err error) {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) AddOrUpdateAccountMetaCache(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) UpdateAccountMetaCache(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(vals []byte, b byte) error {
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrAccountNotFound
		}
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) DeleteAccountMetaCache(key types.Address, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, key)
	if err != nil {
		return err
	}

	return txn.Delete(k)
}

func (l *Ledger) HasAccountMetaCache(key types.Address, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(val []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) Latest(account types.Address, token types.Hash, txns ...db.StoreTxn) types.Hash {
	zero := types.Hash{}
	am, err := l.GetAccountMeta(account, txns...)
	if err != nil {
		return zero
	}
	tm := am.Token(token)
	if tm != nil {
		return tm.Header
	}
	return zero
}

func (l *Ledger) Account(hash types.Hash, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	block, err := l.GetStateBlock(hash, txns...)
	if err != nil {
		return nil, err
	}
	addr := block.GetAddress()
	am, err := l.GetAccountMeta(addr, txns...)
	if err != nil {
		return nil, err
	}

	return am, nil
}

func (l *Ledger) Token(hash types.Hash, txns ...db.StoreTxn) (*types.TokenMeta, error) {
	block, err := l.GetStateBlockConfirmed(hash, txns...)
	if err != nil {
		return nil, err
	}
	token := block.GetToken()
	addr := block.GetAddress()
	am, err := l.GetAccountMetaConfirmed(addr, txns...)
	if err != nil {
		return nil, err
	}

	tm := am.Token(token)
	if tm != nil {
		return tm, nil
	}

	//TODO: hash to token name
	return nil, fmt.Errorf("can not find token %s", token)
}

func (l *Ledger) Balance(account types.Address, txns ...db.StoreTxn) (map[types.Hash]types.Balance, error) {
	cache := make(map[types.Hash]types.Balance)
	am, err := l.GetAccountMeta(account, txns...)
	if err != nil {
		return cache, err
	}
	for _, tm := range am.Tokens {
		cache[tm.Type] = tm.Balance
	}

	if len(cache) == 0 {
		return nil, fmt.Errorf("can not find any token balance ")
	}

	return cache, nil
}

func (l *Ledger) TokenBalance(account types.Address, token types.Hash, txns ...db.StoreTxn) (types.Balance, error) {
	am, err := l.GetAccountMeta(account, txns...)
	if err != nil {
		return types.ZeroBalance, err
	}
	tm := am.Token(token)
	if tm != nil {
		return tm.Balance, nil
	}

	return types.ZeroBalance, fmt.Errorf("can not find %s balance", token)
}

func (l *Ledger) Weight(account types.Address, txns ...db.StoreTxn) types.Balance {
	benefit, err := l.GetRepresentation(account, txns...)
	if err != nil {
		return types.ZeroBalance
	}
	return benefit.Total
}

func (l *Ledger) CalculateAmount(block *types.StateBlock, txns ...db.StoreTxn) (types.Balance, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var prev *types.StateBlock
	var err error
	switch block.GetType() {
	case types.Open:
		return block.TotalBalance(), err
	case types.Send:
		if prev, err = l.GetStateBlock(block.GetPrevious(), txn); err != nil {
			return types.ZeroBalance, err
		}
		return prev.TotalBalance().Sub(block.TotalBalance()), nil
	case types.Receive:
		if prev, err = l.GetStateBlock(block.GetPrevious(), txn); err != nil {
			return types.ZeroBalance, err
		}
		return block.TotalBalance().Sub(prev.TotalBalance()), nil
	case types.Change:
		return types.ZeroBalance, nil
	case types.ContractReward:
		prevHash := block.GetPrevious()
		if prevHash.IsZero() {
			return block.TotalBalance(), nil
		} else {
			if prev, err = l.GetStateBlock(prevHash, txn); err != nil {
				return types.ZeroBalance, err
			}
			return block.TotalBalance().Sub(prev.TotalBalance()), nil
		}
	case types.ContractSend:
		if common.IsGenesisBlock(block) {
			return block.GetBalance(), nil
		} else {
			if prev, err = l.GetStateBlock(block.GetPrevious(), txn); err != nil {
				return types.ZeroBalance, err
			}
			return prev.TotalBalance().Sub(block.TotalBalance()), nil
		}
	default:
		return types.ZeroBalance, errors.New("invalid block type")
	}
}
