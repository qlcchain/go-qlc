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

func (a *AccountApi) Create(seedStr string, i *uint32) (map[string]string, error) {
	var index uint32
	b, err := hex.DecodeString(seedStr)
	if err != nil {
		return nil, err
	}
	if i == nil {
		index = 0
	} else {
		index = *i
	}
	seed, err := types.BytesToSeed(b)
	if err != nil {
		return nil, err
	}
	acc, err := seed.Account(index)
	if err != nil {
		return nil, err
	}
	r := make(map[string]string)
	r["pubKey"] = hex.EncodeToString(acc.Address().Bytes())
	r["privKey"] = hex.EncodeToString(acc.PrivateKey())
	return r, nil
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
