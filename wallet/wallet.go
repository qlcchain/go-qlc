package wallet

import (
	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"io"
)

type Wallet struct {
	io.Closer
	db      db.Store
	PoWChan chan types.Work
}

func NewWallet(seed *types.Seed, index uint32) (*Wallet, error) {
	var accounts []*types.Account
	s := seed.String()
	for i := uint32(0); i < index+1; i++ {
		_, priv, err := types.KeypairFromSeed(s, i)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, types.NewAccount(priv))
	}

	return &Wallet{
		//seed:     seed,
		//accounts: accounts,
		//index:    index,
	}, nil
}

func Generate() (*Wallet, error) {
	seed, err := types.NewSeed()
	if err != nil {
		return nil, err
	}

	return NewWallet(seed, 0)
}

func NewWalletStore(ctx *chain.QlcContext) *WalletStore {
	return &WalletStore{}
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

	Init() error
	Remove() error
}

type walletAction interface {
	Import(content string, password string) error
	Export(path string) error

	GetRawKey(account types.Address) (*types.Account, error)
	//GetBalances get all account balance ordered by token type
	GetBalances() (map[types.Hash]types.Balance, error)
	SearchPending() error
	Send(source types.Address, token types.Hash, to types.Address, amount types.Balance) (*types.Block, error)
	Receive(sendBlock types.Block) (*types.Block, error)
	Change(account types.Address, representative types.Address) (*types.Block, error)
}

type Walleter interface {
	walletAction
	walletStoreApi
}
