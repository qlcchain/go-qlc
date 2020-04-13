package ledger

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

type PrivacyStore interface {
	AddBlockPrivatePayload(hash types.Hash, payload []byte) error
	DeleteBlockPrivatePayload(hash types.Hash) error
	GetBlockPrivatePayload(hash types.Hash) ([]byte, error)
}

func (l *Ledger) AddBlockPrivatePayload(hash types.Hash, payload []byte) error {
	if len(payload) == 0 {
		payload = []byte{1}
	}
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPrivatePayload, hash)
	if err != nil {
		return err
	}
	return l.store.Put(k, payload)
}

func (l *Ledger) DeleteBlockPrivatePayload(hash types.Hash) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPrivatePayload, hash)
	if err != nil {
		return err
	}
	if err := l.store.Delete(k); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetBlockPrivatePayload(hash types.Hash) ([]byte, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPrivatePayload, hash)
	if err != nil {
		return nil, err
	}

	pl, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, errors.New("block private payload not found")
		}
		return nil, fmt.Errorf("get private payload error: %s", err)
	}
	if len(pl) == 1 {
		return nil, nil
	}
	return pl, nil
}
