package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type NetApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

func NewNetApi(l *ledger.Ledger) *NetApi {
	return &NetApi{ledger: l, logger: log.NewLogger("rpc/net")}
}

func (q *NetApi) OnlineRepresentatives() []types.Address {
	as, _ := q.ledger.GetOnlineRepresentations()
	if as == nil {
		return make([]types.Address, 0)
	}
	return as
}
