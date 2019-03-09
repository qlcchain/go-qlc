package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type NetApi struct {
	dpos   *consensus.DposService
	logger *zap.SugaredLogger
}

func NewNetApi(dpos *consensus.DposService) *NetApi {
	return &NetApi{dpos: dpos, logger: log.NewLogger("rpc/net")}
}

func (q *NetApi) OnlineRepresentatives() []types.Address {
	as := q.dpos.GetOnlineRepresentatives()
	if as == nil {
		return make([]types.Address, 0)
	}
	return as
}
