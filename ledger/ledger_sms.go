package ledger

import (
	"encoding/json"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func addSMSInfo(key []byte, hash types.Hash, txn db.StoreTxn) error {
	err := txn.Get(key, func(val []byte, b byte) error {
		hs := new([]types.Hash)
		if err := json.Unmarshal(val, hs); err != nil {
			return err
		}
		*hs = append(*hs, hash)
		val2, err := json.Marshal(*hs)
		if err != nil {
			return err
		}
		return txn.Set(key, val2)
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			v := []types.Hash{hash}
			val, err := json.Marshal(v)
			if err != nil {
				return err
			}
			return txn.Set(key, val)
		}
		return err
	}
	return nil
}

func addSenderOrReceiver(number []byte, t byte, hash types.Hash, txn db.StoreTxn) error {
	if number != nil && len(number) != 0 {
		key := getKeyOfBytes(number, t)
		if err := addSMSInfo(key, hash, txn); err != nil {
			return err
		}
	}
	return nil
}

func addSenderAndReceiver(blk *types.StateBlock, txn db.StoreTxn) error {
	sender := blk.GetSender()
	receiver := blk.GetReceiver()
	hash := blk.GetHash()
	if err := addSenderOrReceiver(sender, idPrefixSender, hash, txn); err != nil {
		return err
	}
	if err := addSenderOrReceiver(receiver, idPrefixReceiver, hash, txn); err != nil {
		return err
	}
	return nil
}

func getSenderOrReceiver(number []byte, t byte, txn db.StoreTxn) ([]types.Hash, error) {
	key := getKeyOfBytes(number, t)
	h := new([]types.Hash)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := json.Unmarshal(val, h); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return []types.Hash{}, err
	}
	return *h, nil
}

func (l *Ledger) GetSenderBlocks(sender []byte, txns ...db.StoreTxn) ([]types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	h, err := getSenderOrReceiver(sender, idPrefixSender, txn)
	if err != nil && err != badger.ErrKeyNotFound {
		return []types.Hash{}, err
	}
	return h, nil
}

func (l *Ledger) GetReceiverBlocks(receiver []byte, txns ...db.StoreTxn) ([]types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	h, err := getSenderOrReceiver(receiver, idPrefixReceiver, txn)
	if err != nil && err != badger.ErrKeyNotFound {
		return []types.Hash{}, err
	}
	return h, nil
}

func deleteSenderAndReceiver(blk *types.StateBlock, txn db.StoreTxn) error {
	sender := blk.GetSender()
	receiver := blk.GetReceiver()
	hash := blk.GetHash()
	if err := deleteSenderOrReceiver(sender, idPrefixSender, hash, txn); err != nil {
		return err
	}
	if err := deleteSenderOrReceiver(receiver, idPrefixReceiver, hash, txn); err != nil {
		return err
	}
	return nil
}

func deleteSenderOrReceiver(number []byte, t byte, hash types.Hash, txn db.StoreTxn) error {
	if number != nil && len(number) != 0 {
		key := getKeyOfBytes(number, t)
		err := txn.Get(key, func(val []byte, b byte) error {
			hs := new([]types.Hash)
			if err := json.Unmarshal(val, hs); err != nil {
				return err
			}
			hashes := *hs
			if len(hashes) == 1 {
				return txn.Delete(key)
			}
			var hashes2 []types.Hash
			for index, h := range hashes {
				if h == hash {
					hashes2 = append(hashes[:index], hashes[index+1:]...)
					break
				}
			}
			val2, err := json.Marshal(hashes2)
			if err != nil {
				return err
			}
			return txn.Set(key, val2)
		})
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
	}
	return nil
}

func (l *Ledger) AddMessageInfo(mHash types.Hash, message []byte, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key := getKeyOfHash(mHash, idPrefixMessageInfo)
	if err := txn.Set(key, message); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetMessageInfo(mHash types.Hash, txns ...db.StoreTxn) ([]byte, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key := getKeyOfHash(mHash, idPrefixMessageInfo)
	var m []byte
	err := txn.Get(key, func(val []byte, b byte) error {
		m = val
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func addMessage(blk *types.StateBlock, txn db.StoreTxn) error {
	message := blk.GetMessage()
	if !message.IsZero() {
		hash := blk.GetHash()
		key := getKeyOfHash(message, idPrefixMessage)
		if err := addSMSInfo(key, hash, txn); err != nil {
			return err
		}
	}
	return nil
}

func deleteMessage(blk *types.StateBlock, txn db.StoreTxn) error {
	key := getKeyOfHash(blk.GetMessage(), idPrefixMessage)
	hash := blk.GetHash()
	err := txn.Get(key, func(val []byte, b byte) error {
		hs := new([]types.Hash)
		if err := json.Unmarshal(val, hs); err != nil {
			return err
		}
		hashes := *hs
		if len(hashes) == 1 {
			mKey := getKeyOfHash(blk.GetMessage(), idPrefixMessageInfo)
			if err := txn.Delete(mKey); err != nil {
				return err
			}
			return txn.Delete(key)
		}
		var hashes2 []types.Hash
		for index, h := range hashes {
			if h == hash {
				hashes2 = append(hashes[:index], hashes[index+1:]...)
				break
			}
		}
		val2, err := json.Marshal(hashes2)
		if err != nil {
			return err
		}
		return txn.Set(key, val2)
	})
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return nil
}

func (l *Ledger) GetMessageBlocks(mHash types.Hash, txns ...db.StoreTxn) ([]types.Hash, error) {
	key := getKeyOfHash(mHash, idPrefixMessage)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	h := new([]types.Hash)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := json.Unmarshal(val, h); err != nil {
			return err
		}
		return nil
	})
	if err != nil && err != badger.ErrKeyNotFound {
		return []types.Hash{}, err
	}
	return *h, nil
}

func addSMSDataForBlock(blk *types.StateBlock, txn db.StoreTxn) error {
	if err := addSenderAndReceiver(blk, txn); err != nil {
		return err
	}
	if err := addMessage(blk, txn); err != nil {
		return err
	}
	return nil
}

func deleteSmsDataForBlock(blk *types.StateBlock, txn db.StoreTxn) error {
	if err := deleteSenderAndReceiver(blk, txn); err != nil {
		return err
	}
	if err := deleteMessage(blk, txn); err != nil {
		return err
	}
	return nil
}
