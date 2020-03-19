package ledger

import (
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

type VoteStore interface {
	AddVoteHistory(hash types.Hash, address types.Address) error
	HasVoteHistory(hash types.Hash, address types.Address) bool
	CleanBlockVoteHistory(hash types.Hash) error
	CleanAllVoteHistory() error
}

func (l *Ledger) AddVoteHistory(hash types.Hash, address types.Address) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixVoteHistory, hash, address)
	if err != nil {
		return err
	}

	return l.store.Put(k, nil)
}

func (l *Ledger) HasVoteHistory(hash types.Hash, address types.Address) bool {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixVoteHistory, hash, address)
	if err != nil {
		return false
	}

	if b, _ := l.store.Has(k); b {
		return true
	} else {
		return false
	}
}

func (l *Ledger) CleanBlockVoteHistory(hash types.Hash) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixVoteHistory, hash)
	if err != nil {
		return err
	}

	err = l.store.Iterator(k, nil, func(key []byte, val []byte) error {
		k := make([]byte, len(key))
		copy(k[:], key)
		if er := l.store.Delete(k); er != nil {
			l.logger.Error(er)
		}
		return nil
	})

	return nil
}

func (l *Ledger) CleanAllVoteHistory() error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixVoteHistory)
	if err != nil {
		return err
	}

	err = l.store.Iterator(k, nil, func(key []byte, val []byte) error {
		k := make([]byte, len(key))
		copy(k[:], key)
		if er := l.store.Delete(k); er != nil {
			l.logger.Error(er)
		}
		return nil
	})

	return nil
}
