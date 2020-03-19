package ledger

import (
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

type SmartBlockStore interface {
	AddSmartContractBlock(value *types.SmartContractBlock) error
	HasSmartContractBlock(key types.Hash) (bool, error)
	GetSmartContractBlock(key types.Hash) (*types.SmartContractBlock, error)
	GetSmartContractBlocks(fn func(block *types.SmartContractBlock) error) error
	CountSmartContractBlocks() (uint64, error)
}

func (l *Ledger) AddSmartContractBlock(value *types.SmartContractBlock) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixSmartContractBlock, value.GetHash())
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	_, err = l.store.Get(k)
	if err == nil {
		return ErrBlockExists
	} else if err != storage.KeyNotFound {
		return err
	}
	return l.store.Put(k, v)
}

func (l *Ledger) GetSmartContractBlock(key types.Hash) (*types.SmartContractBlock, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixSmartContractBlock, key)
	if err != nil {
		return nil, err
	}

	value := new(types.SmartContractBlock)
	val, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	if err := value.Deserialize(val); err != nil {
		return nil, err
	}

	return value, nil
}

func (l *Ledger) GetSmartContractBlocks(fn func(block *types.SmartContractBlock) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixSmartContractBlock)
	return l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		blk := new(types.SmartContractBlock)
		if err := blk.Deserialize(val); err != nil {
			return nil
		}
		if err := fn(blk); err != nil {
			return err
		}
		return nil
	})
}

func (l *Ledger) HasSmartContractBlock(key types.Hash) (bool, error) {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixSmartContractBlock, key)
	return l.store.Has(prefix)
}

func (l *Ledger) CountSmartContractBlocks() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixSmartContractBlock)})
}
