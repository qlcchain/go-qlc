package ledger

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

type PrivacyStore interface {
	AddBlockPrivatePayload(hash types.Hash, payload []byte, c ...storage.Cache) error
	DeleteBlockPrivatePayload(hash types.Hash, c ...storage.Cache) error
	GetBlockPrivatePayload(hash types.Hash, c ...storage.Cache) ([]byte, error)
}

func (l *Ledger) AddBlockPrivatePayload(hash types.Hash, payload []byte, c ...storage.Cache) error {
	if len(payload) == 0 {
		payload = []byte{1}
	}
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPrivatePayload, hash)
	if err != nil {
		return err
	}
	return l.Put(k, payload, c...)
}

func (l *Ledger) DeleteBlockPrivatePayload(hash types.Hash, c ...storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPrivatePayload, hash)
	if err != nil {
		return err
	}
	return l.Delete(k, c...)
}

func (l *Ledger) GetBlockPrivatePayload(hash types.Hash, c ...storage.Cache) ([]byte, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPrivatePayload, hash)
	if err != nil {
		return nil, err
	}

	pl, err := l.Get(k, c...)
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
