package ledger

import (
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/util"
)

type DposStore interface {
	GetLastGapPovHeight() uint64
	SetLastGapPovHeight(height uint64) error
}

const (
	DPoSDataLastGapPovHeight byte = iota
)

func getDPoSDataKey(t byte) []byte {
	key := []byte{byte(storage.KeyPrefixDPoS)}
	key = append(key, t)
	return key
}

func (l *Ledger) GetLastGapPovHeight() uint64 {
	key := getDPoSDataKey(DPoSDataLastGapPovHeight)

	v, err := l.store.Get(key)
	if err != nil {
		return 0
	}

	return util.BE_BytesToUint64(v)
}

func (l *Ledger) SetLastGapPovHeight(height uint64) error {
	key := getDPoSDataKey(DPoSDataLastGapPovHeight)
	return l.store.Put(key, util.BE_Uint64ToBytes(height))
}
