package consensus

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type PoVEngine struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
	ledger *ledger.Ledger
	eb     event.EventBus
	accounts []*types.Account

	bp     *PovBlockProcessor
	txpool *PovTxPool
	chain  *PovBlockChain
	verifier process.BlockVerifier
}

func NewPovEngine(cfg *config.Config, accounts []*types.Account, eb event.EventBus) (*PoVEngine, error) {
	ledger := ledger.NewLedger(cfg.LedgerDir(), eb)

	pov := &PoVEngine{
		logger: log.NewLogger("pov_engine"),
		cfg:    cfg,
		eb:     eb,
		accounts: accounts,
		ledger: ledger,
	}

	pov.bp = NewPovBlockProcessor(pov)
	pov.txpool = NewPovTxPool(pov)
	pov.chain = NewPovBlockChain(pov)
	pov.verifier = process.NewPovVerifier(ledger, pov.chain)

	return pov, nil
}

func (pov *PoVEngine) Init() error {
	pov.bp.Init()
	pov.chain.Init()
	pov.txpool.Init()

	return nil
}

func (pov *PoVEngine) Start() error {
	pov.logger.Info("start pov engine service")

	pov.txpool.Start()

	pov.chain.Start()

	pov.bp.Start()

	pov.setEvent()

	return nil
}

func (pov *PoVEngine) Stop() error {
	pov.unsetEvent()

	pov.txpool.Stop()

	pov.chain.Stop()

	pov.bp.Stop()

	return nil
}

func (pov *PoVEngine) GetConfig() *config.Config {
	return pov.cfg
}

func (pov *PoVEngine) GetLogger() *zap.SugaredLogger {
	return pov.logger
}

func (pov *PoVEngine) GetEventBus() event.EventBus {
	return pov.eb
}

func (pov *PoVEngine) GetLedger() ledger.Store {
	return pov.ledger
}

func (pov *PoVEngine) GetChain() *PovBlockChain {
	return pov.chain
}

func (pov *PoVEngine) GetTxPool() *PovTxPool {
	return pov.txpool
}

func (pov *PoVEngine) GetVerifier() process.BlockVerifier {
	return pov.verifier
}

func (pov *PoVEngine) GetAccounts() []*types.Account {
	return pov.accounts
}

func (pov *PoVEngine) AddMinedBlock(block *types.PovBlock) error {
	return pov.bp.AddBlock(block, types.PovBlockFromLocal)
}

func (pov *PoVEngine) setEvent() error {
	err := pov.eb.SubscribeAsync(string(common.EventRecvPovBlock), pov.onRecvPovBlock, false)
	if err != nil {
		return err
	}
	return nil
}

func (pov *PoVEngine) unsetEvent() error {
	err := pov.eb.Unsubscribe(string(common.EventRecvPovBlock), pov.onRecvPovBlock)
	if err != nil {
		return err
	}
	return nil
}

func (pov *PoVEngine) onRecvPovBlock(block *types.PovBlock, from types.PovBlockFrom) error {
	pov.logger.Infof("receive block [%s] from [%d]", block.GetHash(), from)
	pov.bp.AddBlock(block, from)
	return nil
}
