package api

import (
	"fmt"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
)

type NetApi struct {
	ledger ledger.Store
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

func NewNetApi(l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *NetApi {
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
	return onlineRepsInfo(q.ledger)
}

func onlineRepsInfo(ledger ledger.Store) *OnlineRepTotal {
	as, _ := ledger.GetOnlineRepresentations()
	if as == nil {
		return &OnlineRepTotal{}
	}

	ot := &OnlineRepTotal{
		Reps:       make([]*OnlineRepInfo, 0),
		ValidVotes: types.ZeroBalance,
	}

	supply := config.GenesisBlock().Balance
	minWeight, _ := supply.Div(common.DposVoteDivisor)

	for _, account := range as {
		weight := ledger.Weight(account)
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

func (q *NetApi) ConnectPeersInfo(count int, offset *int) ([]*types.PeerInfo, error) {
	p := q.cc.GetConnectPeersInfo()
	start, end, err := calculateRange(len(p), count, offset)
	if err != nil {
		return nil, err
	}
	return p[start:end], nil
}

func (q *NetApi) GetOnlinePeersInfo(count int, offset *int) ([]*types.PeerInfo, error) {
	p := q.cc.GetOnlinePeersInfo()
	start, end, err := calculateRange(len(p), count, offset)
	if err != nil {
		return nil, err
	}
	return p[start:end], nil
}

func (q *NetApi) GetAllPeersInfo(count int, offset *int) ([]*types.PeerInfo, error) {
	pis := make([]*types.PeerInfo, 0)
	err := q.ledger.GetPeersInfo(func(pi *types.PeerInfo) error {
		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		return nil, err
	}
	start, end, err := calculateRange(len(pis), count, offset)
	if err != nil {
		return nil, err
	}
	return pis[start:end], nil
}

func (q *NetApi) PeersCount() (map[string]uint64, error) {
	p := q.cc.GetConnectPeersInfo()
	connectCount := len(p)

	p = q.cc.GetOnlinePeersInfo()
	onlineCount := len(p)

	var pa []*types.PeerInfo
	err := q.ledger.GetPeersInfo(func(pi *types.PeerInfo) error {
		pa = append(pa, pi)
		return nil
	})
	if err != nil {
		return nil, err
	}
	allCount := len(pa)

	c := make(map[string]uint64)
	c["connect"] = uint64(connectCount)
	c["online"] = uint64(onlineCount)
	c["all"] = uint64(allCount)

	return c, nil
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

func (q *NetApi) GetPeerId() string {
	cfg, _ := q.cc.Config()
	return cfg.P2P.ID.PeerID
}
