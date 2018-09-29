package wallet

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
)

type Wallet struct {
	seed     *types.Seed
	accounts []*ledger.Account
	index    uint32
}

func NewWallet(seed *types.Seed, index uint32) (*Wallet, error) {
	var accounts []*ledger.Account
	s := seed.String()
	for i := uint32(0); i < index+1; i++ {
		_, priv, err := types.KeypairFromSeed(s, i)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, ledger.NewAccount(priv))
	}

	return &Wallet{
		seed:     seed,
		accounts: accounts,
		index:    index,
	}, nil
}

func Generate() (*Wallet, error) {
	seed, err := types.NewSeed()
	if err != nil {
		return nil, err
	}

	return NewWallet(seed, 0)
}

func (w *Wallet) Accounts() []*ledger.Account {
	accounts := make([]*ledger.Account, len(w.accounts))
	copy(accounts, w.accounts)
	return accounts
}
