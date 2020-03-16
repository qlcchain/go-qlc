package api

import (
	"encoding/hex"
	"errors"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"github.com/qlcchain/go-qlc/wallet"
)

type WalletApi struct {
	wallet *wallet.WalletStore
	l      ledger.Store
	logger *zap.SugaredLogger
}

func NewWalletApi(l ledger.Store, wallet *wallet.WalletStore) *WalletApi {
	return &WalletApi{l: l, wallet: wallet, logger: log.NewLogger("api_wallet")}
}

// GetAccounts
func (w *WalletApi) GetAccounts(address types.Address, passphrase string) ([]types.Address, error) {
	s := w.wallet.NewSession(address)
	b, err := s.VerifyPassword(passphrase)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = s.Close()
	}()

	if !b {
		return nil, errors.New("password is invalid")
	}

	index, err := s.GetDeterministicIndex()
	if err != nil {
		index = 0
	}
	var accounts []types.Address
	if seedArray, err := s.GetSeed(); err == nil {
		max := util.UInt32Max(uint32(index), uint32(100))
		s, err := types.BytesToSeed(seedArray)
		if err != nil {
			return accounts, err
		}
		for i := uint32(0); i < max; i++ {
			if account, err := s.Account(i); err == nil {
				address := account.Address()
				if _, err := w.l.GetAccountMeta(address); err == nil {
					accounts = append(accounts, address)
				} else {
					break
				}
			}
		}
	} else {
		return nil, err
	}

	return accounts, nil
}

// GetBalance returns balance for each token of the wallet
func (w *WalletApi) GetBalances(address types.Address, passphrase string) (map[string]types.Balance, error) {
	accounts, err := w.GetAccounts(address, passphrase)
	if err != nil {
		return nil, err
	}

	cache := make(map[string]types.Balance)
	vmContext := vmstore.NewVMContext(w.l)

	for _, account := range accounts {
		if am, err := w.l.GetAccountMeta(account); err == nil {
			for _, tm := range am.Tokens {
				info, err := abi.GetTokenById(vmContext, tm.Type)
				if err != nil {
					return nil, err
				}
				if balance, ok := cache[info.TokenName]; ok {
					//b := cache[tm.Type]
					cache[info.TokenName] = balance.Add(tm.Balance)
				} else {
					cache[info.TokenName] = tm.Balance
				}
			}
		}
	}
	return cache, nil
}

func (w *WalletApi) GetRawKey(address types.Address, passphrase string) (map[string]string, error) {
	session := w.wallet.NewSession(address)
	b, err := session.VerifyPassword(passphrase)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = session.Close()
	}()

	if !b {
		return nil, errors.New("password is invalid")
	}
	acc, err := session.GetRawKey(address)
	if err != nil {
		return nil, err
	}
	r := make(map[string]string)
	r["pubKey"] = hex.EncodeToString(acc.Address().Bytes())
	r["privKey"] = hex.EncodeToString(acc.PrivateKey())
	return r, nil
}

// NewSeed generates new seed
func (w *WalletApi) NewSeed() (string, error) {
	seed, err := types.NewSeed()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(seed[:]), nil
}

// NewWallet creates wallet from hex seed string and passphrase ,
// seed string it is a optional parameter, if not set, will create seed randomly
func (w *WalletApi) NewWallet(passphrase string, seed *string) (types.Address, error) {
	var seedStr string
	if seed == nil {
		new, err := types.NewSeed()
		if err != nil {
			return types.ZeroAddress, err
		}
		seedStr = new.String()
	} else {
		seedStr = *seed
	}
	//w.logger.Debug(seedStr)
	return w.wallet.NewWalletBySeed(seedStr, passphrase)
}

func (w *WalletApi) List() ([]types.Address, error) {
	addrs, err := w.wallet.WalletIds()
	if err != nil {
		return nil, err
	}
	return addrs, nil
}

func (w *WalletApi) Remove(addr types.Address) error {
	return w.wallet.RemoveWallet(addr)
}

func (w *WalletApi) ChangePassword(addr types.Address, pwd string, newPwd string) error {
	session := w.wallet.NewSession(addr)
	b, err := session.VerifyPassword(pwd)
	if err != nil {
		return err
	}
	if !b {
		return errors.New("password is invalid")
	}
	err = session.ChangePassword(newPwd)
	if err != nil {
		return err
	}
	return nil
}
