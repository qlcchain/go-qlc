package api

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
	"sync"
	"time"
)

const syncTimeout = 10 * time.Second

var (
	lastSyncTime int64
	mu           sync.Mutex
)

type NetApi struct {
	ledger *ledger.Ledger
	eb     event.EventBus
	logger *zap.SugaredLogger
}

func syncingTime(t time.Time) {
	mu.Lock()
	defer mu.Unlock()
	lastSyncTime = t.Add(syncTimeout).UTC().Unix()
}
func NewNetApi(l *ledger.Ledger, eb event.EventBus) *NetApi {
	_ = eb.Subscribe(string(common.EventSyncing), syncingTime)
	return &NetApi{ledger: l, eb: eb, logger: log.NewLogger("api_net")}
}

func (q *NetApi) OnlineRepresentatives() []types.Address {
	as, _ := q.ledger.GetOnlineRepresentations()
	if as == nil {
		return make([]types.Address, 0)
	}
	return as
}

type PeersInfo struct {
	Count int               `json:"count"`
	Infos map[string]string `json:"infos"`
}

func (q *NetApi) ConnectPeersInfo() *PeersInfo {
	p := make(map[string]string)
	q.eb.Publish(string(common.EventPeersInfo), p)
	i := &PeersInfo{
		Count: len(p),
		Infos: p,
	}
	return i
}
func (q *NetApi) Syncing() bool {
	now := time.Now().UTC().Unix()
	if lastSyncTime < now {
		return false
	}
	return true
}
