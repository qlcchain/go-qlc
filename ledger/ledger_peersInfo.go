package ledger

import (
	"errors"
	"strings"

	"github.com/dgraph-io/badger"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) AddPeerInfo(info *types.PeerInfo, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPeerInfo, []byte(info.PeerID))
	if err != nil {
		return err
	}
	v, err := info.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrPeerExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	if err := txn.Set(k, v); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetPeerInfo(peerID string, txns ...db.StoreTxn) (*types.PeerInfo, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixPeerInfo, []byte(peerID))
	if err != nil {
		return nil, err
	}

	pi := new(types.PeerInfo)
	err = txn.Get(key, func(val []byte, b byte) error {
		if err := pi.Deserialize(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPovHeaderNotFound
		}
		return nil, err
	}
	return pi, nil
}

func (l *Ledger) GetPeersInfo(fn func(info *types.PeerInfo) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixPeerInfo, func(key []byte, val []byte, b byte) error {
		pi := new(types.PeerInfo)
		if err := pi.Deserialize(val); err != nil {
			l.logger.Errorf("deserialize peerInfo error: %s", err)
			errStr = append(errStr, err.Error())
			return nil
		}
		if err := fn(pi); err != nil {
			l.logger.Errorf("process peerInfo error: %s", err)
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

func (l *Ledger) CountPeersInfo(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Count([]byte{idPrefixPeerInfo})
}

func (l *Ledger) UpdatePeerInfo(value *types.PeerInfo, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPeerInfo, []byte(value.PeerID))
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
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrPeerNotFound
		}
		return err
	}
	if err := txn.Set(k, v); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) AddOrUpdatePeerInfo(value *types.PeerInfo, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPeerInfo, []byte(value.PeerID))
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}
	if err := txn.Set(k, v); err != nil {
		return err
	}
	return nil
}
