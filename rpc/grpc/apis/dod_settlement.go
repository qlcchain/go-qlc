// +build testnet

package apis

import (
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	"go.uber.org/zap"
)

type DoDSettlementAPI struct {
	contract *api.DoDSettlementAPI
	logger   *zap.SugaredLogger
}

func NewDoDSettlementAPI(cfgFile string, l ledger.Store) *DoDSettlementAPI {
	return &DoDSettlementAPI{
		contract: api.NewDoDSettlementAPI(cfgFile, l),
		logger:   log.NewLogger("grpc_dodSettlement"),
	}
}

//TODO: implement dod settlement apis
