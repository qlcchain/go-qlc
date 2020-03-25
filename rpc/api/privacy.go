package api

import (
	"errors"
	"github.com/qlcchain/go-qlc/common/types"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
)

type PrivacyApi struct {
	cfg    *config.Config
	l      ledger.Store
	logger *zap.SugaredLogger
	eb     event.EventBus
	feb    *event.FeedEventBus
	cc     *chainctx.ChainContext
}

func NewPrivacyApi(cfg *config.Config, l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *PrivacyApi {
	api := &PrivacyApi{
		cfg:    cfg,
		l:      l,
		eb:     eb,
		feb:    cc.FeedEventBus(),
		logger: log.NewLogger("rpc/privacy"),
		cc:     cc,
	}
	return api
}

func (api *PrivacyApi) DistributeRawPayload() (*types.Hash, error) {
	return nil, errors.New("api not support")
}
