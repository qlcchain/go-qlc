package api

import (
	"encoding/hex"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type AccountApi struct {
	logger *zap.SugaredLogger
}

func NewAccountApi() *AccountApi {
	return &AccountApi{logger: log.NewLogger("api_account")}
}

func (a *AccountApi) Create(seed string, index *uint32) (*types.Account, error) {
	return nil, nil
}

func (a *AccountApi) ForPublicKey(pubStr string) (types.Address, error) {
	pub, err := hex.DecodeString(pubStr)
	if err != nil {
		return types.ZeroAddress, err
	}
	addr := types.PubToAddress(pub)
	a.logger.Debug(addr)
	return addr, nil
}

func (a *AccountApi) PublicKey(addr types.Address) string {
	pub := hex.EncodeToString(addr.Bytes())
	return pub
}

func (a *AccountApi) Validate(addr string) bool {
	return types.IsValidHexAddress(addr)
}
