package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type smsStore interface {
	// sender or receiver
	GetSenderBlocks(sender string, txns ...db.StoreTxn) ([]types.Hash, error)
	GetReceiverBlocks(receiver string, txns ...db.StoreTxn) ([]types.Hash, error)
	GetMessageBlock(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error)
}
