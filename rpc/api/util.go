package api

import (
	"errors"
	"math"
	"math/big"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type UtilApi struct {
	logger *zap.SugaredLogger
	ctx    *vmstore.VMContext
}

func NewUtilApi(l *ledger.Ledger) *UtilApi {
	return &UtilApi{ctx: vmstore.NewVMContext(l), logger: log.NewLogger("api_util")}
}

func (u *UtilApi) Decrypt(cryptograph string, passphrase string) (string, error) {
	return util.Decrypt(cryptograph, passphrase)
}

func (u *UtilApi) Encrypt(raw string, passphrase string) (string, error) {
	return util.Encrypt(raw, passphrase)
}

func decimal(d uint8) int64 {
	m := int64(math.Pow10(int(d)))
	if m <= 0 {
		return 0
	}
	return m
}

func (u *UtilApi) RawToBalance(balance types.Balance, unit string, tokenName *string) (*APIBalance, error) {
	if tokenName != nil {
		token, err := abi.GetTokenByName(u.ctx, *tokenName)
		if err != nil {
			return nil, err
		}
		if token.TokenId != common.ChainToken() {
			b := new(big.Float).SetInt(big.NewInt(balance.Int64()))
			d := new(big.Float).SetFloat64(math.Pow10(int(token.Decimals)))
			r := new(big.Float).Quo(b, d)
			return &APIBalance{r}, nil
		}
	}
	b, err := common.RawToBalanceFloat(balance, unit)
	if err != nil {
		return nil, err
	}
	return &APIBalance{b}, nil
}

func (u *UtilApi) BalanceToRaw(balance types.Balance, unit string, tokenName *string) (types.Balance, error) {
	if tokenName != nil {
		token, err := abi.GetTokenByName(u.ctx, *tokenName)
		if err != nil {
			return types.ZeroBalance, err
		}
		if token.TokenId != common.ChainToken() {
			d := decimal(token.Decimals)
			if d == 0 {
				return types.ZeroBalance, errors.New("error decimals")
			}
			return balance.Mul(d), nil
		}
	}
	b, err := common.BalanceToRaw(balance, unit)
	if err != nil {
		return types.ZeroBalance, err
	}
	return b, nil
}

type APIBalance struct {
	*big.Float
}

func (b *APIBalance) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *APIBalance) String() string {
	return b.Text('f', -1)
}
