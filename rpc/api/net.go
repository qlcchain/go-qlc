package api

import (
	"fmt"

	p2pmetrics "github.com/libp2p/go-libp2p-core/metrics"
	"go.uber.org/zap"

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

func NewNetApi(l *ledger.Ledger, eb event.EventBus) *NetApi {
	return &NetApi{ledger: l, eb: eb, logger: log.NewLogger("api_net")}
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

//type PeersInfo struct {
//	Count int               `json:"count"`
//	Infos []*types.PeerInfo `json:"infos"`
//}

func (q *NetApi) ConnectPeersInfo(count int, offset *int) ([]*types.PeerInfo, error) {
	c, o, err := checkOffset(count, offset)
	if err != nil {
		return nil, err
	}
	var p []*types.PeerInfo
	q.eb.Publish(common.EventPeersInfo, &p)
	r := p[o : c+o]
	return r, nil
}

func (q *NetApi) GetOnlinePeersInfo(count int, offset *int) ([]*types.PeerInfo, error) {
	c, o, err := checkOffset(count, offset)
	if err != nil {
		return nil, err
	}
	var p []*types.PeerInfo
	q.eb.Publish(common.EventOnlinePeersInfo, &p)
	r := p[o : c+o]
	return r, nil
}

func (q *NetApi) GetAllPeersInfo(count int, offset *int) ([]*types.PeerInfo, error) {
	c, o, err := checkOffset(count, offset)
	if err != nil {
		return nil, err
	}
	pis := make([]*types.PeerInfo, 0)
	err = q.ledger.GetPeersInfo(func(pi *types.PeerInfo) error {
		pis = append(pis, pi)
		return nil
	})
	if err != nil {
		return nil, err
	}
	pis2 := pis[o : c+o]
	return pis2, nil
}

func (q *NetApi) PeersCount() (map[string]uint64, error) {
	var p []*types.PeerInfo

	q.eb.Publish(common.EventPeersInfo, &p)
	connectCount := len(p)

	q.eb.Publish(common.EventOnlinePeersInfo, &p)
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

func (q *NetApi) GetBandwidthStats() *p2pmetrics.Stats {
	stats := new(p2pmetrics.Stats)
	q.eb.Publish(common.EventGetBandwidthStats, stats)
	return stats
}

func (q *NetApi) Syncing() bool {
	var ss common.SyncState
	q.eb.Publish(common.EventSyncStatus, &ss)
	if ss == common.Syncing || ss == common.SyncDone {
		return true
	}
	return false
}
