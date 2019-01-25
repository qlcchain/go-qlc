package api

import (
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

func (a *AccountApi) Create(seed string, index *int) (*types.Account, error) {
	return nil, nil
}
