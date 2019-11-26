package miner

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus/pov"
	"github.com/qlcchain/go-qlc/log"
)

type Miner struct {
	logger     *zap.SugaredLogger
	cfg        *config.Config
	eb         event.EventBus
	subscriber *event.ActorSubscriber
	povEngine  *pov.PoVEngine

	povWorker *PovWorker
	syncState topic.SyncState
}

func NewMiner(cfgFile string, povEngine *pov.PoVEngine) *Miner {
	cc := context.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	miner := &Miner{
		logger:    log.NewLogger("miner"),
		cfg:       cfg,
		eb:        cc.EventBus(),
		povEngine: povEngine,
	}

	miner.povWorker = NewPovWorker(miner)

	return miner
}

func (miner *Miner) Init() error {
	err := miner.povWorker.Init()
	if err != nil {
		return err
	}

	return nil
}

func (miner *Miner) Start() error {
	miner.logger.Info("start miner service")

	miner.subscriber = event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case topic.SyncState:
			miner.onRecvPovSyncState(msg)
		}
	}), miner.eb)

	if err := miner.subscriber.Subscribe(topic.EventPovSyncState); err != nil {
		return err
	}

	err := miner.povWorker.Start()
	if err != nil {
		return err
	}

	return nil
}

func (miner *Miner) Stop() error {
	miner.logger.Info("stop miner service")

	if err := miner.subscriber.Unsubscribe(topic.EventPovSyncState); err != nil {
		return err
	}

	err := miner.povWorker.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (miner *Miner) GetConfig() *config.Config {
	return miner.cfg
}

func (miner *Miner) GetLogger() *zap.SugaredLogger {
	return miner.logger
}

func (miner *Miner) GetPovEngine() *pov.PoVEngine {
	return miner.povEngine
}

func (miner *Miner) GetSyncState() topic.SyncState {
	return miner.syncState
}

func (miner *Miner) onRecvPovSyncState(state topic.SyncState) {
	miner.logger.Infof("receive pov sync state [%s]", state)
	miner.syncState = state
}
