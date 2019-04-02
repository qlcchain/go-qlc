package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type smsStore interface {
	// sender or receiver
	GetSenderBlocks(sender []byte, txns ...db.StoreTxn) ([]types.Hash, error)
	GetReceiverBlocks(receiver []byte, txns ...db.StoreTxn) ([]types.Hash, error)
	//GetMessageBlock(hash types.Hash, txns ...Store.StoreTxn) (*types.StateBlock, error)
	AddMessageInfo(mHash types.Hash, message []byte, txns ...db.StoreTxn) error
	GetMessageInfo(mHash types.Hash, txns ...db.StoreTxn) ([]byte, error)
	GetMessageBlocks(mHash types.Hash, txns ...db.StoreTxn) ([]types.Hash, error)
}
