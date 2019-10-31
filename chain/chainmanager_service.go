package chain

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/log"
)

type Manager struct {
	id string
	common.ServiceLifecycle
	cfgFile string
	logger  *zap.SugaredLogger
}

func NewChainManageService(cfgFile string) *Manager {
	return &Manager{
		cfgFile: cfgFile,
		logger:  log.NewLogger("chainManage_service"),
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
	if id, err := eb.Subscribe(common.EventRestartChain, restartChain); err != nil {
		c.logger.Error(err)
	} else {
		c.id = id
	}
	return nil
}

func (c *Manager) Stop() error {
	if !c.PreStop() {
		return errors.New("pre stop fail")
	}
	defer c.PostStop()
	cc := context.NewChainContext(c.cfgFile)
	eb := cc.EventBus()
	if err := eb.Unsubscribe(common.EventRestartChain, c.id); err != nil {
		c.logger.Error(err)
	}
	return nil
}

func (c *Manager) Status() int32 {
	return c.State()
}

func restartChain(cfgFile string, isSave bool) {
	logger := log.NewLogger("chainManage_service")
	go func() {
		cc := context.NewChainContext(cfgFile)
		cfgManager, err := cc.ConfigManager()
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
}
