package ledger

import (
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
)

func (l *Ledger) GetSenderBlocks(sender string, txns ...db.StoreTxn) ([]types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	h, err := getSenderOrReceiver(sender, idPrefixSender, txn)
	if err != nil {
		return []types.Hash{}, err
	}
	return h, nil
}

func (l *Ledger) GetReceiverBlocks(receiver string, txns ...db.StoreTxn) ([]types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	h, err := getSenderOrReceiver(receiver, idPrefixReceiver, txn)
	if err != nil {
		return []types.Hash{}, err
	}
	return h, nil
}

func getSenderOrReceiver(number string, t byte, txn db.StoreTxn) ([]types.Hash, error) {
	key := getKeyOfBytes(util.String2Bytes(number), t)
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

func (l *Ledger) GetMessageBlock(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	key := getKeyOfHash(hash, idPrefixMessage)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	h := new(types.Hash)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := h.UnmarshalBinary(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, errors.New("message block can not found")
		}
		return nil, err
	}
	return l.GetStateBlock(*h)
}

func (l *Ledger) addMessageInfo(message string, txns ...db.StoreTxn) (types.Hash, error) {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)
	if message == "" {
		return types.ZeroHash, nil
	}
	m := util.String2Bytes(message)
	hash, err := types.HashBytes(m)
	if err != nil {
		return types.ZeroHash, err
	}
	key := getKeyOfHash(hash, idPrefixMessageInfo)
	if err := txn.Set(key, m); err != nil {
		return types.ZeroHash, err
	}
	return hash, nil
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

func addSenderOrReceiver(number string, t byte, hash types.Hash, txn db.StoreTxn) error {
	if number != "" {
		key := getKeyOfBytes(util.String2Bytes(number), t)
		err := txn.Get(key, func(val []byte, b byte) error {
			hs := new([]types.Hash)
			if err := json.Unmarshal(val, hs); err != nil {
				return err
			}
			//for _, h := range *hs {
			//	if h == hash {
			//		return nil
			//	}
			//}
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

func addMessage(blk *types.StateBlock, txn db.StoreTxn) error {
	message := blk.GetMessage()
	if !message.IsZero() {
		hash := blk.GetHash()
		key := getKeyOfHash(blk.GetMessage(), idPrefixMessage)
		val := make([]byte, types.HashSize)
		err := hash.MarshalBinaryTo(val)
		if err != nil {
			return err
		}
		return txn.Set(key, val)
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

func deleteMessage(blk *types.StateBlock, txn db.StoreTxn) error {
	key := getKeyOfHash(blk.GetMessage(), idPrefixMessage)
	if err := txn.Delete(key); err != nil {
		return err
	}
	key = getKeyOfHash(blk.GetMessage(), idPrefixMessageInfo)
	if err := txn.Delete(key); err != nil {
		return err
	}
	return nil
}
