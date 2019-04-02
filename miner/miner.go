package miner

import (
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type Miner struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
}

func NewMiner(cfg *config.Config) *Miner {
	miner := &Miner{
		logger: log.NewLogger("miner"),
		cfg:    cfg,
	}

	return miner
}

func (miner *Miner) Init() error {
	return nil
}

func (miner *Miner) Start() error {
	miner.logger.Info("start miner service")

	return nil
}

func (miner *Miner) Stop() error {
	return nil
}
