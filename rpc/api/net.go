package api

import (
	"fmt"

	"github.com/qlcchain/go-qlc/common/topic"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
)

type NetApi struct {
	ledger *ledger.Ledger
	eb     event.EventBus
	logger *zap.SugaredLogger
	cc     *chainctx.ChainContext
}

type OnlineRepTotal struct {
	Reps              []*OnlineRepInfo
	ValidVotes        types.Balance
	ValidVotesPercent string
}

type OnlineRepInfo struct {
	Account types.Address
	Vote    types.Balance
}

func NewNetApi(l *ledger.Ledger, eb event.EventBus, cc *chainctx.ChainContext) *NetApi {
	return &NetApi{ledger: l, eb: eb, logger: log.NewLogger("api_net"), cc: cc}
}

func (q *NetApi) OnlineRepresentatives() []types.Address {
	as, _ := q.ledger.GetOnlineRepresentations()
	if as == nil {
		return make([]types.Address, 0)
	}
	return as
}

func (q *NetApi) OnlineRepsInfo() *OnlineRepTotal {
	as, _ := q.ledger.GetOnlineRepresentations()
	if as == nil {
		return &OnlineRepTotal{}
	}

	ot := &OnlineRepTotal{
		Reps:       make([]*OnlineRepInfo, 0),
		ValidVotes: types.ZeroBalance,
	}

	supply := common.GenesisBlock().Balance
	minWeight, _ := supply.Div(common.DposVoteDivisor)

	for _, account := range as {
		weight := q.ledger.Weight(account)
		oi := &OnlineRepInfo{
			Account: account,
			Vote:    weight,
		}
		ot.Reps = append(ot.Reps, oi)

		if weight.Compare(minWeight) == types.BalanceCompBigger {
			ot.ValidVotes = ot.ValidVotes.Add(weight)
		}
	}

	ot.ValidVotesPercent = fmt.Sprintf("%.2f%%", float64(ot.ValidVotes.Uint64()*100)/float64(supply.Uint64()))

	return ot
}

type PeersInfo struct {
	Count int               `json:"count"`
	Infos map[string]string `json:"infos"`
}

func (q *NetApi) ConnectPeersInfo() *PeersInfo {
	p := q.cc.GetPeersPool()
	i := &PeersInfo{
		Count: len(p),
		Infos: p,
	}
	return i
}

func (q *NetApi) GetBandwidthStats() *topic.EventBandwidthStats {
	return q.cc.GetBandwidthStats()
}

func (q *NetApi) Syncing() bool {
	ss := q.cc.P2PSyncState()
	if ss == topic.Syncing || ss == topic.SyncDone {
		return true
	}
	return false
}
