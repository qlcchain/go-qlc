package consensus

import (
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"go.uber.org/zap"
)

type PoVEngine struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
	ledger *ledger.Ledger
	ns     p2p.Service
}

func NewPovEngine(cfg *config.Config, netService p2p.Service) (*PoVEngine, error) {
	ledger := ledger.NewLedger(cfg.LedgerDir())

	pov := &PoVEngine{
		logger: log.NewLogger("pov_engine"),
		cfg:    cfg,
		ns:     netService,
		ledger: ledger,
	}
	return pov, nil
}

func (pov *PoVEngine) Init() error {
	return nil
}

func (pov *PoVEngine) Start() error {
	pov.logger.Info("start pov engine service")

	return nil
}

func (pov *PoVEngine) Stop() error {
	return nil
}
