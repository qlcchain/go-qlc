/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package ledger

import (
	"fmt"
	"math/big"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/common/vmcontract/mintage"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
)

const (
	prefixLen = 1 + types.AddressSize //pre+len(contract_address)
)

type Iterator struct {
	address *types.Address
	store   storage.Store
	logger  *zap.SugaredLogger
}

func (i *Iterator) Next(prefix []byte, fn func(key []byte, value []byte) error) error {
	var storageKey []byte
	storageKey = append(storageKey, []byte{byte(storage.KeyPrefixVMStorage)}...)
	storageKey = append(storageKey, i.address[:]...)
	storageKey = append(storageKey, prefix...)
	err := i.store.Iterator(storageKey, nil, func(key []byte, val []byte) error {
		if err := fn(key[prefixLen:], val); err != nil {
			i.logger.Error(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

type vmStore interface {
	// NewVMIterator new Iterator by contract address
	NewVMIterator(address *types.Address) *Iterator
	SetStorage(val map[string]interface{}) error
	SaveStorage(val map[string]interface{}, c ...storage.Cache) error
	ListTokens() ([]*types.TokenInfo, error)
	GetTokenById(tokenId types.Hash) (*types.TokenInfo, error)
	GetTokenByName(tokenName string) (*types.TokenInfo, error)
}

func (l *Ledger) NewVMIterator(address *types.Address) *Iterator {
	return &Iterator{
		address: address,
		store:   l.DBStore(),
		logger:  l.logger,
	}
}

// set storage to badger, all value need to be slice
func (l *Ledger) SetStorage(val map[string]interface{}) error {
	batch := l.store.Batch(false)
	for k, v := range val {
		if err := batch.Put([]byte(k), v.([]byte)); err != nil {
			return err
		}
	}
	return l.store.PutBatch(batch)
}

// save storage to cache
func (l *Ledger) SaveStorage(val map[string]interface{}, c ...storage.Cache) error {
	if len(c) > 0 && c[0] != nil {
		for k, v := range val {
			if err := c[0].Put([]byte(k), v); err != nil {
				return err
			}
		}
	} else {
		if err := l.cache.BatchUpdate(func(c *Cache) error {
			for k, v := range val {
				if err := c.Put([]byte(k), v); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) ListTokens() ([]*types.TokenInfo, error) {
	logger := log.NewLogger("ListTokens")
	defer func() {
		logger.Sync()
	}()

	iterator := l.NewVMIterator(&contractaddress.MintageAddress)

	var infos []*types.TokenInfo
	if err := iterator.Next(nil, func(key []byte, value []byte) error {
		if len(value) > 0 {
			tokenId, _ := types.BytesToHash(key[(types.AddressSize + 1):])
			if config.IsGenesisToken(tokenId) {
				if info, err := mintage.ParseGenesisTokenInfo(value); err == nil {
					infos = append(infos, info)
				} else {
					logger.Error(err)
				}
			} else {
				if info, err := mintage.ParseTokenInfo(value); err == nil {
					exp := new(big.Int).Exp(util.Big10, new(big.Int).SetUint64(uint64(info.Decimals)), nil)
					info.TotalSupply = info.TotalSupply.Mul(info.TotalSupply, exp)
					infos = append(infos, info)
				} else {
					logger.Error(err)
				}
			}
		}
		return nil
	}); err == nil {
		return infos, nil
	} else {
		return nil, err
	}
}

func (l *Ledger) GetTokenById(tokenId types.Hash) (*types.TokenInfo, error) {
	if _, ok := l.tokenCache.Load(tokenId); !ok {
		if err := l.saveCache(); err != nil {
			return nil, err
		}
	}
	if ti, ok := l.tokenCache.Load(tokenId); ok {
		return ti.(*types.TokenInfo), nil
	}

	return nil, fmt.Errorf("can not find token %s", tokenId.String())
}

func (l *Ledger) GetTokenByName(tokenName string) (*types.TokenInfo, error) {
	if _, ok := l.tokenCache.Load(tokenName); !ok {
		if err := l.saveCache(); err != nil {
			return nil, err
		}
	}
	if ti, ok := l.tokenCache.Load(tokenName); ok {
		return ti.(*types.TokenInfo), nil
	}

	return nil, fmt.Errorf("can not find token %s", tokenName)
}

func (l *Ledger) saveCache() error {
	if infos, err := l.ListTokens(); err == nil {
		for _, v := range infos {
			l.tokenCache.Store(v.TokenId, v)
			l.tokenCache.Store(v.TokenName, v)
		}
		return nil
	} else {
		return err
	}
}
