// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

// Note: This file generate CURD code by genny automatically (https://github.com/cheekybits/genny)
// Usage:
// - install gen, command is  ` go get github.com/cheekybits/genny
// - create type in /common/types package, example: StateBlock, Hash
// - generate code, command is ` cat ledger_generic.go | genny gen "ContractValue=StateBlock ContractKey=Hash" > ledger_type.go `
// - generate testcase code, command is ` cat ledger_generic_test.go | genny gen "ContractValue=StateBlock ContractKey=Hash" > ledger_type_test.go `

package ledger

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

var (
	ErrContractValueNotFound = errors.New("ContractValue not found")
	LatestSuffix             = []byte("LATEST")
	//GlobalSuffix             = []byte("GLOBAL")
)

type contractValueStore interface {
	GetContractValue(key *types.ContractKey, c ...storage.Cache) (*types.ContractValue, error)
	DeleteContractValue(key *types.ContractKey, c ...storage.Cache) error
	AddOrUpdateContractValue(key *types.ContractKey, value *types.ContractValue, c ...storage.Cache) error
	CountContractValues() (uint64, error)
	IteratorContractStorage(prefix []byte, callback func(key *types.ContractKey, value *types.ContractValue) error) error
	UpdateContractValueByBlock(block *types.StateBlock, trieRoot *types.Hash, c ...storage.Cache) error
	DeleteContractValueByBlock(block *types.StateBlock, c ...storage.Cache) error
}

func (l *Ledger) GetContractValue(key *types.ContractKey, c ...storage.Cache) (*types.ContractValue, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixContractValue, key)
	if err != nil {
		return nil, err
	}

	if v, err := l.getFromCache(k, c...); err == nil {
		return v.(*types.ContractValue), nil
	}

	value := new(types.ContractValue)
	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrContractValueNotFound
		}
		return nil, err
	}
	if err := value.Deserialize(v); err != nil {
		return nil, err
	}
	return value, nil
}

func (l *Ledger) DeleteContractValue(key *types.ContractKey, c ...storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixContractValue, key)
	if err != nil {
		return err
	}

	if len(c) > 0 {
		return c[0].Delete(k)
	} else {
		return l.store.Delete(k)
	}
}

func (l *Ledger) AddOrUpdateContractValue(key *types.ContractKey, value *types.ContractValue, c ...storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixContractValue, key)
	if err != nil {
		return err
	}

	if len(c) > 0 {
		return c[0].Put(k, value)
	}

	v, err := value.Serialize()
	if err != nil {
		return err
	}

	return l.store.Put(k, v)
}

func (l *Ledger) GetContractValues(fn func(key *types.ContractKey, value *types.ContractValue) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixContractValue)

	return l.store.Iterator(prefix, nil, func(k []byte, v []byte) error {
		key := new(types.ContractKey)
		if err := key.Deserialize(k[1:]); err != nil {
			return err
		}
		value := new(types.ContractValue)
		if err := value.Deserialize(v); err != nil {
			return err
		}
		if err := fn(key, value); err != nil {
			return err
		}
		return nil
	})
}

func (l *Ledger) HasContractValue(key *types.ContractKey) (bool, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixContractValue, key)
	if err != nil {
		return false, err
	}
	_, _, err = l.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return false, ErrContractValueNotFound
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) CountContractValues() (uint64, error) {
	return l.store.Count([]byte{storage.KeyPrefixContractValue})
}

func (l *Ledger) IteratorContractStorage(prefix []byte, callback func(key *types.ContractKey, value *types.ContractValue) error) error {
	var p []byte
	p = append(p, storage.KeyPrefixContractValue)
	p = append(p, prefix...)

	if err := l.Iterator(p, nil, func(key []byte, val []byte) error {
		k := &types.ContractKey{}
		v := &types.ContractValue{}
		if err := k.Deserialize(key[1:]); err != nil {
			return err
		}
		if err := v.Deserialize(val); err != nil {
			return err
		}
		return callback(k, v)
	}); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) UpdateContractValueByBlock(block *types.StateBlock, trieRoot *types.Hash, c ...storage.Cache) error {
	a := block.Address
	ca := block.ContractAddress()
	hash := block.GetHash()
	if ca == nil {
		return fmt.Errorf("can not extract contract address from block[%s]", hash)
	}

	previous := types.ZeroHash
	latestKey := &types.ContractKey{
		ContractAddress: *ca,
		AccountAddress:  a,
		Hash:            block.Token,
		Suffix:          LatestSuffix,
	}

	// get latest to fetch previous block hash
	if value, err := l.GetContractValue(latestKey, c...); err == nil {
		previous = value.BlockHash
	} else {
		return err
	}

	// update latest key/value
	latestV := &types.ContractValue{
		BlockHash: hash,
		Root:      nil,
	}
	if err := l.AddOrUpdateContractValue(latestKey, latestV, c...); err != nil {
		return err
	}

	// add the block as contract value
	ck := &types.ContractKey{
		ContractAddress: *ca,
		AccountAddress:  a,
		Hash:            hash,
		Suffix:          nil,
	}

	cv := &types.ContractValue{BlockHash: previous, Root: trieRoot}
	if err := l.AddOrUpdateContractValue(ck, cv, c...); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) DeleteContractValueByBlock(block *types.StateBlock, c ...storage.Cache) error {
	a := block.Address
	ca := block.ContractAddress()
	hash := block.GetHash()
	if ca == nil {
		return fmt.Errorf("can not extract contract address from block[%s]", hash)
	}
	// find previous block hash
	previous := types.ZeroHash
	key := &types.ContractKey{
		ContractAddress: *ca,
		AccountAddress:  a,
		Hash:            hash,
	}
	if value, err := l.GetContractValue(key, c...); err == nil {
		previous = value.BlockHash
	} else {
		if err != ErrContractValueNotFound {
			return err
		}
	}

	// remove current block hash
	if err := l.DeleteContractValue(key, c...); err != nil {
		return err
	}

	// set latest block hash to previous
	ck := &types.ContractKey{
		ContractAddress: *ca,
		AccountAddress:  a,
		Hash:            block.Token,
		Suffix:          LatestSuffix,
	}

	if previous.IsZero() {
		if err := l.DeleteContractValue(ck, c...); err != nil {
			return err
		}
	} else {
		if err := l.AddOrUpdateContractValue(ck, &types.ContractValue{BlockHash: previous}); err != nil {
			return err
		}
	}

	return nil
}