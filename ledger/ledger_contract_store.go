package ledger

import (
	"github.com/qlcchain/go-qlc/ledger/db"
)

type contractStore interface {
	//GetStorage
	GetStorage(prefix, key []byte, txns ...db.StoreTxn) ([]byte, error)
	SetStorage(prefix, key []byte, value []byte, txns ...db.StoreTxn) error
}
