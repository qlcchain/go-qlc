package miner

import (
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type Miner struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
	povEngine *consensus.PoVEngine

	povWorker *PovWorker
}

func NewMiner(cfg *config.Config, povEngine *consensus.PoVEngine) *Miner {
	miner := &Miner{
		logger: log.NewLogger("miner"),
		cfg:    cfg,
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