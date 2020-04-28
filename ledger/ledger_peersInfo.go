package ledger

import (
	"errors"
	"strings"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/types"
)

type PeerInfoStore interface {
	AddPeerInfo(info *types.PeerInfo) error
	GetPeerInfo(peerID string) (*types.PeerInfo, error)
	GetPeersInfo(fn func(info *types.PeerInfo) error) error
	CountPeersInfo() (uint64, error)
	UpdatePeerInfo(value *types.PeerInfo) error
	AddOrUpdatePeerInfo(value *types.PeerInfo) error
}

func (l *Ledger) AddPeerInfo(info *types.PeerInfo) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPeerInfo, []byte(info.PeerID))
	if err != nil {
		return err
	}
	v, err := info.Serialize()
	if err != nil {
		return err
	}

	_, err = l.store.Get(k)
	if err == nil {
		return ErrPeerExists
	} else if err != storage.KeyNotFound {
		return err
	}
	return l.store.Put(k, v)
}

func (l *Ledger) GetPeerInfo(peerID string) (*types.PeerInfo, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixPeerInfo, []byte(peerID))
	if err != nil {
		return nil, err
	}

	pi := new(types.PeerInfo)
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrPovHeaderNotFound
		}
		return nil, err
	}
	if err := pi.Deserialize(val); err != nil {
		return nil, err
	}
	return pi, nil
}

func (l *Ledger) GetPeersInfo(fn func(info *types.PeerInfo) error) error {
	errStr := make([]string, 0)
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPeerInfo)

	err := l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
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

func (l *Ledger) CountPeersInfo() (uint64, error) {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixPeerInfo)
	return l.store.Count(prefix)
}

func (l *Ledger) UpdatePeerInfo(value *types.PeerInfo) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPeerInfo, []byte(value.PeerID))
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	_, err = l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return ErrPeerNotFound
		}
		return err
	}
	return l.store.Put(k, v)
}

func (l *Ledger) AddOrUpdatePeerInfo(value *types.PeerInfo) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixPeerInfo, []byte(value.PeerID))
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return l.store.Put(k, v)
}
