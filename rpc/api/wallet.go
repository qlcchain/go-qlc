package api

import (
	"encoding/hex"

	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/wallet"
	"go.uber.org/zap"
)

type WalletApi struct {
	wallet *wallet.WalletStore
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

func NewWalletApi(l *ledger.Ledger, wallet *wallet.WalletStore) *WalletApi {
	return &WalletApi{ledger: l, wallet: wallet, logger: log.NewLogger("api_wallet")}
}

// GetBalance returns balance for each token of the wallet
func (w *WalletApi) GetBalances(address types.Address, passphrase string) (map[string]types.Balance, error) {
	session := w.wallet.NewSession(address)
	b, err := session.VerifyPassword(passphrase)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, errors.New("password is invalid")
	}
	balances, err := session.GetBalances()
	if err != nil {
		return nil, err
	}
	cache := make(map[string]types.Balance)
	vmContext := vmstore.NewVMContext(w.ledger)
	for token, balance := range balances {
		info, err := abi.GetTokenById(vmContext, token)
		if err != nil {
			return nil, err
		}
		cache[info.TokenName] = balance
	}
	return cache, nil
}

func (w *WalletApi) GetRawKey(address types.Address, passphrase string) (map[string]string, error) {
	session := w.wallet.NewSession(address)
	b, err := session.VerifyPassword(passphrase)
	if err != nil {
		return nil, err
	}
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
