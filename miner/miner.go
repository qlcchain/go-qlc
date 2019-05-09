package miner

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type Miner struct {
	logger    *zap.SugaredLogger
	cfg       *config.Config
	eb        event.EventBus
	povEngine *consensus.PoVEngine

	povWorker *PovWorker
	syncState common.SyncState
}

func NewMiner(cfg *config.Config, povEngine *consensus.PoVEngine) *Miner {
	miner := &Miner{
		logger:    log.NewLogger("miner"),
		cfg:       cfg,
		eb:        event.GetEventBus(cfg.LedgerDir()),
		povEngine: povEngine,
	}

	miner.povWorker = NewPovWorker(miner)

	return miner
}

func (miner *Miner) Init() error {

	miner.povWorker.Init()

	return nil
}

func (miner *Miner) Start() error {
	miner.logger.Info("start miner service")

	err := miner.eb.Subscribe(string(common.EventPovSyncState), miner.onRecvPovSyncState)
	if err != nil {
		return err
	}

	miner.povWorker.Start()

	return nil
}

func (miner *Miner) Stop() error {

	miner.povWorker.Stop()

	return nil
}

func (miner *Miner) GetConfig() *config.Config {
	return miner.cfg
}

func (miner *Miner) GetLogger() *zap.SugaredLogger {
	return miner.logger
}

func (miner *Miner) GetPovEngine() *consensus.PoVEngine {
	return miner.povEngine
}

func (miner *Miner) GetSyncState() common.SyncState {
	return miner.syncState
}

func (miner *Miner) onRecvPovSyncState(state common.SyncState) {
	miner.logger.Infof("receive pov sync state [%s]", state)
	miner.syncState = state
}
