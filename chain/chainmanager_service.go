package chain

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

type Manager struct {
	common.ServiceLifecycle
	subscriber *event.ActorSubscriber
	cfgFile    string
	logger     *zap.SugaredLogger
}

func NewChainManageService(cfgFile string) *Manager {
	return &Manager{
		cfgFile: cfgFile,
		logger:  log.NewLogger("chain_manager_service"),
	}
}

func (c *Manager) Init() error {
	if !c.PreInit() {
		return errors.New("pre init fail")
	}
	defer c.PostInit()

	return nil
}

func (c *Manager) Start() error {
	if !c.PreStart() {
		return errors.New("pre start fail ")
	}
	defer c.PostStart()
	cc := context.NewChainContext(c.cfgFile)
	eb := cc.EventBus()
	c.subscriber = event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *types.Tuple:
			restartChain(msg.First.(string), msg.Second.(bool))
		}
	}), eb)

	return c.subscriber.Subscribe(topic.EventRestartChain)
}

func (c *Manager) Stop() error {
	if !c.PreStop() {
		return errors.New("pre stop fail")
	}
	defer c.PostStop()

	return c.subscriber.Unsubscribe(topic.EventRestartChain)
}

func (c *Manager) Status() int32 {
	return c.State()
}

func restartChain(cfgFile string, isSave bool) <-chan interface{} {
	logger := log.NewLogger("chain_manager_service")
	quit := make(chan interface{})

	go func() {
		cc := context.NewChainContext(cfgFile)
		cfgManager, err := cc.ConfigManager()
		defer func() {
			quit <- struct{}{}
			close(quit)
		}()
		if err != nil {
			logger.Error(err)
		}
		if isSave {
			if err := cfgManager.CommitAndSave(); err != nil {
				logger.Error(err)
			}
			if err := cc.Destroy(); err != nil {
				logger.Error(err)
			}
			ccNew := context.NewChainContext(cfgFile)
			err = ccNew.Init(func() error {
				return RegisterServices(ccNew)
			})
			if err != nil {
				logger.Error(err)
			}
			if err := ccNew.Start(); err != nil {
				logger.Error(err)
			}
		} else {
			if err := cfgManager.Commit(); err != nil {
				logger.Error(err)
			}
			if err := cc.Destroy(); err != nil {
				logger.Error(err)
			}
			ccNew := context.NewChainContextFromOriginal(cc)
			err = ccNew.Init(func() error {
				return RegisterServices(ccNew)
			})
			if err != nil {
				logger.Error(err)
			}
			if err := ccNew.Start(); err != nil {
				logger.Error(err)
			}
		}
	}()

	return quit
}
