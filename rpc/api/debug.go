package api

import (
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type DebugApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

func NewDebugApi(l *ledger.Ledger) *DebugApi {
	return &DebugApi{ledger: l, logger: log.NewLogger("api_debug")}
}

func (l *DebugApi) BlockCacheCount() (map[string]uint64, error) {
	unCount, err := l.ledger.CountBlockCache()
	if err != nil {
		return nil, err
	}
	c := make(map[string]uint64)
	c["blockCache"] = unCount
	return c, nil
}
