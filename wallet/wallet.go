package wallet

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
)

type Wallet struct {
	db      WalletStore
	PoWChan chan types.Work
}

var log = common.NewLogger("wallet")

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
