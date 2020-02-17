package ledger

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

func (l *Ledger) AddOrUpdatePerformance(value *types.PerformanceTime) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPerformance, value.Hash)
	if err != nil {
		return err
	}
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return l.store.Put(k, v)
}

func (l *Ledger) PerformanceTimes(fn func(*types.PerformanceTime)) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPerformance)
	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		pt := new(types.PerformanceTime)
		err := json.Unmarshal(val, pt)
		if err != nil {
			return err
		}
		fn(pt)
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetPerformanceTime(key types.Hash) (*types.PerformanceTime, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPerformance, key)
	if err != nil {
		return nil, err
	}

	value := types.NewPerformanceTime()
	val, err := l.store.Get(k)

	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrPerformanceNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(val, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func (l *Ledger) IsPerformanceTimeExist(key types.Hash) (bool, error) {
	if _, err := l.GetPerformanceTime(key); err == nil {
		return true, nil
	} else {
		if err == ErrPerformanceNotFound {
			return false, nil
		} else {
			return false, err
		}
	}
}
