package ledger

import (
	"errors"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
)

type AccountStore interface {
	GetAccountMeta(address types.Address, c ...storage.Cache) (*types.AccountMeta, error)
	GetAccountMetas(fn func(am *types.AccountMeta) error) error
	CountAccountMetas() (uint64, error)
	HasAccountMetaConfirmed(address types.Address) (bool, error)
	AddAccountMeta(value *types.AccountMeta, c storage.Cache) error

	GetAccountMetaConfirmed(address types.Address, c ...storage.Cache) (*types.AccountMeta, error)

	GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error)
	HasTokenMeta(address types.Address, tokenType types.Hash) (bool, error)

	AddTokenMetaConfirmed(address types.Address, meta *types.TokenMeta, cache *Cache) error
	DeleteTokenMetaConfirmed(address types.Address, tokenType types.Hash, c *Cache) error
	GetTokenMetaConfirmed(address types.Address, tokenType types.Hash) (*types.TokenMeta, error)

	AddOrUpdateAccountMetaCache(value *types.AccountMeta, batch ...storage.Batch) error
	UpdateAccountMeteCache(value *types.AccountMeta, batch ...storage.Batch) error
	DeleteAccountMetaCache(key types.Address, batch ...storage.Batch) error
	GetAccountMeteCache(key types.Address, batch ...storage.Batch) (*types.AccountMeta, error)
	GetAccountMetaCaches(fn func(am *types.AccountMeta) error) error
	HasAccountMetaCache(key types.Address) (bool, error)

	Weight(account types.Address) types.Balance
	CalculateAmount(block *types.StateBlock) (types.Balance, error)
}

func (l *Ledger) GetAccountMeta(address types.Address, c ...storage.Cache) (*types.AccountMeta, error) {
	am, err := l.GetAccountMeteCache(address)
	if err != nil {
		am = nil
	}

	meta, er := l.GetAccountMetaConfirmed(address, c...)
	if er != nil {
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
					if temp.Type == config.ChainToken() {
						am.CoinBalance = meta.GetBalance()
						am.CoinOracle = meta.GetOracle()
						am.CoinNetwork = meta.GetNetwork()
						am.CoinVote = meta.GetVote()
						am.CoinStorage = meta.GetStorage()
					}
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

// AccountMeta Confirmed
func (l *Ledger) AddAccountMeta(value *types.AccountMeta, c storage.Cache) error {
	if _, err := l.GetAccountMetaConfirmed(value.Address); err == nil {
		return ErrAccountExists
	}

	k, err := storage.GetKeyOfParts(storage.KeyPrefixAccount, value.Address)
	if err != nil {
		return err
	}
	return c.Put(k, value.Clone())
}

func (l *Ledger) UpdateAccountMeta(value *types.AccountMeta, c storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixAccount, value.Address)
	if err != nil {
		return err
	}
	return c.Put(k, value.Clone())
}

func (l *Ledger) DeleteAccountMeta(key types.Address, c storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixAccount, key)
	if err != nil {
		return err
	}
	return c.Delete(k)
}

func (l *Ledger) HasAccountMetaConfirmed(address types.Address) (bool, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixAccount, address)
	if err != nil {
		return false, err
	}

	if r, err := l.getFromCache(k); r != nil {
		return true, nil
	} else {
		if err == ErrKeyDeleted {
			return false, nil
		}
	}

	return l.store.Has(k)
}

func (l *Ledger) GetAccountMetaConfirmed(address types.Address, c ...storage.Cache) (*types.AccountMeta, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixAccount, address)
	if err != nil {
		return nil, err
	}

	if r, err := l.getFromCache(k, c...); r != nil {
		return r.(*types.AccountMeta).Clone(), nil
	} else {
		if err == ErrKeyDeleted {
			return nil, ErrAccountNotFound
		}
	}

	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	meta := new(types.AccountMeta)
	if err := meta.Deserialize(v); err != nil {
		return nil, err
	}
	return meta, nil
}

func (l *Ledger) GetAccountMetas(fn func(am *types.AccountMeta) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixAccount)

	return l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		am := new(types.AccountMeta)
		if err := am.Deserialize(val); err != nil {
			return err
		}
		if err := fn(am); err != nil {
			return err
		}
		return nil
	})
}

func (l *Ledger) CountAccountMetas() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixAccount)})
}

// AccountMeta UnConfirmed
func (l *Ledger) AddAccountMetaCache(value *types.AccountMeta, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCacheAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	if _, err := b.Get(k); err == nil {
		return ErrAccountExists
	}

	return b.Put(k, v)
}

func (l *Ledger) AddOrUpdateAccountMetaCache(value *types.AccountMeta, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCacheAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return b.Put(k, v)
}

func (l *Ledger) UpdateAccountMeteCache(value *types.AccountMeta, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCacheAccount, value.Address)
	if err != nil {
		return err
	}
	if _, err := b.Get(k); err != nil {
		l.logger.Error(err)
		return ErrAccountNotFound
	}

	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return b.Put(k, v)
}

func (l *Ledger) GetAccountMeteCache(address types.Address, batch ...storage.Batch) (*types.AccountMeta, error) {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	key, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCacheAccount, address)
	if err != nil {
		return nil, err
	}
	v, err := b.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	value := new(types.AccountMeta)
	if err := value.Deserialize(v.([]byte)); err != nil {
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetAccountMetaCaches(fn func(am *types.AccountMeta) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixBlockCacheAccount)

	return l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		am := new(types.AccountMeta)
		if err := am.Deserialize(val); err != nil {
			return err
		}
		if err := fn(am); err != nil {
			return err
		}
		return nil
	})
}

func (l *Ledger) DeleteAccountMetaCache(address types.Address, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	key, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCacheAccount, address)
	if err != nil {
		return err
	}

	return b.Delete(key)
}

func (l *Ledger) HasAccountMetaCache(address types.Address) (bool, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCacheAccount, address)
	if err != nil {
		return false, err
	}
	return l.store.Has(k)
}

// Token
func (l *Ledger) GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error) {
	am, err := l.GetAccountMeta(address)
	if err != nil {
		return nil, err
	}

	tm := am.Token(tokenType)
	if tm == nil {
		return nil, ErrTokenNotFound
	}
	return tm, nil
}

func (l *Ledger) HasTokenMeta(address types.Address, tokenType types.Hash) (bool, error) {
	am, err := l.GetAccountMeta(address)
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

// Token Confirmed
func (l *Ledger) GetTokenMetaConfirmed(address types.Address, tokenType types.Hash) (*types.TokenMeta, error) {
	am, err := l.GetAccountMetaConfirmed(address)
	if err != nil {
		return nil, err
	}

	tm := am.Token(tokenType)
	if tm == nil {
		return nil, ErrTokenNotFound
	}

	return tm, nil
}

func (l *Ledger) AddTokenMetaConfirmed(address types.Address, meta *types.TokenMeta, cache *Cache) error {
	am, err := l.GetAccountMeta(address)
	if err != nil {
		return err
	}

	if am.Token(meta.Type) != nil {
		return ErrTokenExists
	}

	am.Tokens = append(am.Tokens, meta)
	return l.UpdateAccountMeta(am, cache)
}

func (l *Ledger) DeleteTokenMetaConfirmed(address types.Address, tokenType types.Hash, c *Cache) error {
	am, err := l.GetAccountMetaConfirmed(address)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range tokens {
		if token.Type == tokenType {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}
	return l.UpdateAccountMeta(am, c)
}

// Token UnConfirmed
func (l *Ledger) DeleteTokenMetaCache(address types.Address, tokenType types.Hash, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	am, err := l.GetAccountMeteCache(address, batch...)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range tokens {
		if token.Type == tokenType {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}
	return l.UpdateAccountMeteCache(am, batch...)
}

func (l *Ledger) Weight(account types.Address) types.Balance {
	benefit, err := l.GetRepresentation(account)
	if err != nil {
		return types.ZeroBalance
	}
	return benefit.Total
}

func (l *Ledger) CalculateAmount(block *types.StateBlock) (types.Balance, error) {
	var prev *types.StateBlock
	var err error
	switch block.GetType() {
	case types.Open:
		return block.TotalBalance(), err
	case types.Send:
		if prev, err = l.GetStateBlock(block.GetPrevious()); err != nil {
			return types.ZeroBalance, err
		}
		return prev.TotalBalance().Sub(block.TotalBalance()), nil
	case types.Receive:
		if prev, err = l.GetStateBlock(block.GetPrevious()); err != nil {
			return types.ZeroBalance, err
		}
		return block.TotalBalance().Sub(prev.TotalBalance()), nil
	case types.Change, types.Online:
		return types.ZeroBalance, nil
	case types.ContractReward:
		prevHash := block.GetPrevious()
		if prevHash.IsZero() {
			return block.TotalBalance(), nil
		} else {
			if prev, err = l.GetStateBlock(prevHash); err != nil {
				return types.ZeroBalance, err
			}
			return block.TotalBalance().Sub(prev.TotalBalance()), nil
		}
	case types.ContractSend:
		if config.IsGenesisBlock(block) {
			return block.GetBalance(), nil
		} else {
			if prev, err = l.GetStateBlock(block.GetPrevious()); err != nil {
				return types.ZeroBalance, err
			}
			return prev.TotalBalance().Sub(block.TotalBalance()), nil
		}
	default:
		return types.ZeroBalance, errors.New("invalid block type")
	}
}
