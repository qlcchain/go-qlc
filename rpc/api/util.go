package api

import (
	"math"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type UtilApi struct {
	logger *zap.SugaredLogger
	ledger *ledger.Ledger
}

func NewUtilApi(l *ledger.Ledger) *UtilApi {
	return &UtilApi{ledger: l, logger: log.NewLogger("api_util")}
}

func (u *UtilApi) Decrypt(cryptograph string, passphrase string) (string, error) {
	return util.Decrypt(cryptograph, passphrase)
}

func (u *UtilApi) Encrypt(raw string, passphrase string) (string, error) {
	return util.Encrypt(raw, passphrase)
}

func (u *UtilApi) RawToBalance(balance types.Balance, unit string, tokenName *string) (types.Balance, error) {
	if tokenName != nil {
		token, err := u.ledger.GetTokenByName(*tokenName)
		if err != nil {
			return types.ZeroBalance, err
		}
		if token.TokenId != common.ChainToken() {
			b, err := balance.Div(int64(math.Pow10(int(token.Decimals))))
			if err != nil {
				return types.ZeroBalance, nil
			}
			return b, nil
		}
	}
	b, err := common.RawToBalance(balance, unit)
	if err != nil {
		return types.ZeroBalance, err
	}
	return b, nil
}

func (u *UtilApi) BalanceToRaw(balance types.Balance, unit string, tokenName *string) (types.Balance, error) {
	if tokenName != nil {
		token, err := u.ledger.GetTokenByName(*tokenName)
		if err != nil {
			return types.ZeroBalance, err
		}
		if token.TokenId != common.ChainToken() {
			return balance.Mul(int64(math.Pow10(int(token.Decimals)))), nil
		}
	}
	b, err := common.BalanceToRaw(balance, unit)
	if err != nil {
		return types.ZeroBalance, err
	}
	return b, nil
}
