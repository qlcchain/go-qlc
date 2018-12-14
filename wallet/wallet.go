package wallet

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"io"
)

type Wallet struct {
	io.Closer
	db      db.Store
	PoWChan chan types.Work
}

type walletStoreApi interface {
	GetWalletId() ([]byte, error)
	GetVersion() (int64, error)
	SetVersion(version int64) error
	GetSeed() ([]byte, error)
	SetSeed(seed []byte) error
	GetDeterministicIndex() (int64, error)
	ResetDeterministicIndex() error
	SetDeterministicIndex(index int64) error
	GetWork(hash types.Address) (types.Work, error)
	GetRepresentative() (types.Address, error)
	SetRepresentative(address types.Address) error
	IsAccountExist(addr types.Address) bool
	GetAccounts() ([]types.Address, error)
	ChangePassword(password string) error
	EnterPassword(password string) error
}

type walletAction interface {
	Import(content string, password string) error
	Export(path string) error

	GetRawKey(account types.Address) (*types.Account, error)
	//GetBalances get all account balance ordered by token type
	GetBalances() (map[types.Hash]types.Balance, error)
	SearchPending() error
	GenerateSendBlock(source types.Address, token types.Hash, to types.Address, amount types.Balance) (*types.Block, error)
	GenerateReceiveBlock(sendBlock types.Block) (*types.Block, error)
	GenerateChangeBlock(account types.Address, representative types.Address) (*types.Block, error)
}

type IWallet interface {
	walletAction
	walletStoreApi
}
