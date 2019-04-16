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
	txpool *PovTxPool
	chain  *PovBlockChain
}

func NewPovEngine(cfg *config.Config, netService p2p.Service) (*PoVEngine, error) {
	ledger := ledger.NewLedger(cfg.LedgerDir())

	pov := &PoVEngine{
		logger: log.NewLogger("pov_engine"),
		cfg:    cfg,
		ns:     netService,
		ledger: ledger,
	}

	pov.txpool = NewPovTxPool(pov)
	pov.chain = NewPovBlockChain(pov)

	return pov, nil
}

func (pov *PoVEngine) Init() error {
	pov.chain.Init()

	return nil
}

func (pov *PoVEngine) Start() error {
	pov.logger.Info("start pov engine service")

	pov.txpool.Start()

	pov.chain.Start()

	return nil
}

func (pov *PoVEngine) Stop() error {
	pov.txpool.Stop()

	pov.chain.Stop()

	return nil
}

func (pov *PoVEngine) GetConfig() *config.Config {
	return pov.cfg
}

func (pov *PoVEngine) GetNetService() p2p.Service {
	return pov.ns
}

func (pov *PoVEngine) GetLedger() ledger.Store {
	return pov.ledger
}