package ledger

import (
	"errors"
	"strings"

	"github.com/qlcchain/go-qlc/common/util"

	"github.com/dgraph-io/badger"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) uncheckedKindToPrefix(kind types.UncheckedKind) byte {
	switch kind {
	case types.UncheckedKindPrevious:
		return idPrefixUncheckedBlockPrevious
	case types.UncheckedKindLink:
		return idPrefixUncheckedBlockLink
	case types.UncheckedKindTokenInfo:
		return idPrefixUncheckedTokenInfo
	default:
		panic("bad unchecked block kind")
	}
}

func (l *Ledger) AddUncheckedBlock(key types.Hash, value *types.StateBlock, kind types.UncheckedKind, sync types.SynchronizedKind, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(l.uncheckedKindToPrefix(kind), key)
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
		return ErrUncheckedBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.SetWithMeta(k, v, byte(sync))
}

func (l *Ledger) GetUncheckedBlock(key types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (*types.StateBlock, types.SynchronizedKind, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(l.uncheckedKindToPrefix(kind), key)
	if err != nil {
		return nil, 0, err
	}

	value := new(types.StateBlock)
	var sync types.SynchronizedKind
	err = txn.Get(k, func(val []byte, b byte) (err error) {
		if err = value.Deserialize(val); err != nil {
			return err
		}
		sync = types.SynchronizedKind(b)
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, 0, ErrUncheckedBlockNotFound
		}
		return nil, 0, err
	}
	return value, sync, nil
}

func (l *Ledger) DeleteUncheckedBlock(key types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(l.uncheckedKindToPrefix(kind), key)
	if err != nil {
		return err
	}
	return txn.Delete(k)
}

func (l *Ledger) HasUncheckedBlock(key types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(l.uncheckedKindToPrefix(kind), key)
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

func (l *Ledger) walkUncheckedBlocks(kind types.UncheckedKind, visit types.UncheckedBlockWalkFunc, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	prefix := l.uncheckedKindToPrefix(kind)
	err := txn.Iterator(prefix, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		h, err := types.BytesToHash(key[1:])
		if err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		if err := visit(blk, h, kind, types.SynchronizedKind(b)); err != nil {
			l.logger.Error("visit error %s", err)
			errStr = append(errStr, err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc, txns ...db.StoreTxn) error {
	if err := l.walkUncheckedBlocks(types.UncheckedKindPrevious, visit, txns...); err != nil {
		return err
	}

	if err := l.walkUncheckedBlocks(types.UncheckedKindLink, visit, txns...); err != nil {
		return err
	}

	return l.walkUncheckedBlocks(types.UncheckedKindTokenInfo, visit, txns...)
}

func (l *Ledger) CountUncheckedBlocks(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	var count uint64
	count, err := txn.Count([]byte{idPrefixUncheckedBlockLink})
	if err != nil {
		return 0, err
	}

	count2, err := txn.Count([]byte{idPrefixUncheckedBlockPrevious})
	if err != nil {
		return 0, err
	}

	count3 := l.CountGapPovBlocks()

	return count + count2 + count3, nil
}

func (l *Ledger) AddGapPovBlock(height uint64, block *types.StateBlock, sync types.SynchronizedKind, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixUncheckedPovHeight, height, []byte(sync.String()))
	if err != nil {
		return err
	}

	blocks := types.StateBlockList{}
	err = txn.Get(k, func(val []byte, b byte) (err error) {
		if err = blocks.Deserialize(val); err != nil {
			return err
		}
		return nil
	})

	for _, blk := range blocks {
		if blk.GetHash() == block.GetHash() {
			return nil
		}
	}

	blocks = append(blocks, block)
	v, err := blocks.Serialize()
	if err != nil {
		return err
	}

	return txn.Set(k, v)
}

func (l *Ledger) GetGapPovBlock(height uint64, txns ...db.StoreTxn) (types.StateBlockList, []types.SynchronizedKind, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	kind := make([]types.SynchronizedKind, 0)

	blocks1 := types.StateBlockList{}
	k, err := getKeyOfParts(idPrefixUncheckedPovHeight, height, []byte(types.Synchronized.String()))
	if err != nil {
		return nil, nil, err
	}

	err = txn.Get(k, func(val []byte, b byte) (err error) {
		if err = blocks1.Deserialize(val); err != nil {
			return err
		}

		for i := 0; i < len(blocks1); i++ {
			kind = append(kind, types.Synchronized)
		}

		return nil
	})

	blocks2 := types.StateBlockList{}
	k, err = getKeyOfParts(idPrefixUncheckedPovHeight, height, []byte(types.UnSynchronized.String()))
	if err != nil {
		return nil, nil, err
	}

	err = txn.Get(k, func(val []byte, b byte) (err error) {
		if err = blocks2.Deserialize(val); err != nil {
			return err
		}

		for i := 0; i < len(blocks2); i++ {
			kind = append(kind, types.UnSynchronized)
		}

		return nil
	})

	blocks1 = append(blocks1, blocks2...)
	return blocks1, kind, nil
}

func (l *Ledger) CountGapPovBlocks(txns ...db.StoreTxn) uint64 {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	var count uint64
	_ = txn.Iterator(idPrefixUncheckedPovHeight, func(key []byte, val []byte, b byte) error {
		blocks := types.StateBlockList{}
		if err := blocks.Deserialize(val); err != nil {
			return nil
		}
		count += uint64(len(blocks))
		return nil
	})

	return count
}

func (l *Ledger) DeleteGapPovBlock(height uint64, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixUncheckedPovHeight, height, []byte(types.Synchronized.String()))
	if err != nil {
		return err
	}

	err1 := txn.Delete(k)

	k, err = getKeyOfParts(idPrefixUncheckedPovHeight, height, []byte(types.UnSynchronized.String()))
	if err != nil {
		return err
	}

	err2 := txn.Delete(k)

	if err1 != nil {
		return err1
	}

	if err2 != nil {
		return err2
	}

	return nil
}

func (l *Ledger) WalkGapPovBlocks(visit types.GapPovBlockWalkFunc, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixUncheckedPovHeight, func(key []byte, val []byte, b byte) error {
		blocks := types.StateBlockList{}
		if err := blocks.Deserialize(val); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}

		height := util.BE_BytesToUint64(key[1:9])
		kind := types.StringToSyncKind(string(key[9:]))
		if err := visit(blocks, height, kind); err != nil {
			l.logger.Error("visit error %s", err)
			errStr = append(errStr, err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) AddGapPublishBlock(key types.Hash, blk *types.StateBlock, sync types.SynchronizedKind, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixGapPublish, key, blk.GetHash())
	if err != nil {
		return err
	}

	v, err := blk.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrUncheckedBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.SetWithMeta(k, v, byte(sync))
}

func (l *Ledger) DeleteGapPublishBlock(key types.Hash, blkHash types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixGapPublish, key, blkHash)
	if err != nil {
		return err
	}

	return txn.Delete(k)
}

func (l *Ledger) WalkGapPublishBlock(key types.Hash, visit types.GapPublishBlockWalkFunc, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixGapPublish, key)
	if err != nil {
		return err
	}

	errStr := make([]string, 0)
	err = txn.PrefixIterator(k, func(bytes []byte, bytes2 []byte, b byte) error {
		block := types.StateBlock{}
		err := block.Deserialize(bytes2)
		if err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}

		err = visit(&block, types.SynchronizedKind(b))
		if err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}

		return nil
	})
	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}
