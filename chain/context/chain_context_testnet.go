// +build testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package context

import (
	"errors"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"go.uber.org/atomic"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/hashmap"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
)

var cache = hashmap.New(10)

var ErrPoVNotFinish = errors.New("pov sync is not finished, please check it")

const (
	LedgerService       = "ledgerService"
	WalletService       = "walletService"
	BootNodeHttpService = "bootNodeHttpService"
	P2PService          = "P2PService"
	ConsensusService    = "consensusService"
	RPCService          = "rpcService"
	PovService          = "povService"
	MinerService        = "minerService"
	AutoReceiveService  = "autoReceiveService"
	MetricsService      = "metricsService"
	ChainManageService  = "chainManageService"
	LogService          = "logService"
	ResendBlockService  = "resendBlockService"
	PrivacyService      = "privacyService"
	PermissionService   = "permissionService"
)

type serviceManager interface {
	common.Service
	Register(name string, service common.Service) error
	UnRegister(name string) error
	AllServices() ([]common.Service, error)
	Service(name string) (common.Service, error)
	HasService(name string) bool
	//Control
	ReloadService(name string) error
	RestartAll() error
	// config
	ConfigManager() (*config.CfgManager, error)
	Config() (*config.Config, error)
	EventBus() event.EventBus
}

type Option func(cm *config.CfgManager) error

func NewChainContext(cfgFile string) *ChainContext {
	var dataDir string
	if len(cfgFile) == 0 {
		dataDir = config.DefaultDataDir()
		cfgFile = path.Join(dataDir, config.QlcConfigFile)
	} else {
		cm := config.NewCfgManagerWithFile(cfgFile)
		dataDir, _ = cm.ParseDataDir()
	}
	id := types.HashData([]byte(dataDir)).String()
	if v, ok := cache.GetStringKey(id); ok {
		return v.(*ChainContext)
	} else {
		sr := &ChainContext{
			services:         newServiceContainer(),
			cfgFile:          cfgFile,
			chainID:          id,
			connectPeersPool: new(sync.Map),
			bandwidthStats:   new(topic.EventBandwidthStats),
		}
		sr.povSyncState.Store(topic.SyncNotStart)
		sr.p2pSyncState.Store(topic.SyncNotStart)
		cache.Set(id, sr)
		return sr
	}
}

func NewChainContextFromOriginal(cc *ChainContext) *ChainContext {
	cfgFile := cc.cfgFile
	ctx := NewChainContext(cfgFile)
	cfg, err := cc.Config()
	if err != nil {
		log.Root.Error(err)
	}
	ctx.accounts = cc.accounts
	ctx.cm = config.NewCfgManagerWithConfig(cfgFile, cfg)
	return ctx
}

type ChainContext struct {
	common.ServiceLifecycle
	services         *serviceContainer
	cm               *config.CfgManager
	cfgFile          string
	chainID          string
	locker           sync.RWMutex
	accounts         []*types.Account
	povSyncState     atomic.Value
	p2pSyncState     atomic.Value
	subscriber       *event.ActorSubscriber
	connectPeersPool *sync.Map
	connectPeersInfo []*types.PeerInfo
	onlinePeersInfo  []*types.PeerInfo
	bandwidthStats   *topic.EventBandwidthStats
}

func (cc *ChainContext) EventBus() event.EventBus {
	return event.GetEventBus(cc.Id())
}

func (cc *ChainContext) FeedEventBus() *event.FeedEventBus {
	return event.GetFeedEventBus(cc.Id())
}

func (cc *ChainContext) PoVState() topic.SyncState {
	return cc.povSyncState.Load().(topic.SyncState)
}

func (cc *ChainContext) IsPoVDone() bool {
	return cc.povSyncState.Load().(topic.SyncState) == topic.SyncDone
}

func (cc *ChainContext) P2PSyncState() topic.SyncState {
	return cc.p2pSyncState.Load().(topic.SyncState)
}

func (cc *ChainContext) GetPeersPool() map[string]string {
	p := make(map[string]string)
	cc.connectPeersPool.Range(func(key, value interface{}) bool {
		peerId := key.(string)
		addr := value.(string)
		p[peerId] = addr
		return true
	})
	return p
}

func (cc *ChainContext) GetConnectPeersInfo() []*types.PeerInfo {
	return cc.connectPeersInfo
}

func (cc *ChainContext) GetOnlinePeersInfo() []*types.PeerInfo {
	return cc.onlinePeersInfo
}

func (cc *ChainContext) GetBandwidthStats() *topic.EventBandwidthStats {
	return cc.bandwidthStats
}

func (cc *ChainContext) ConfigFile() string {
	return cc.cfgFile
}

func (cc *ChainContext) Init(fn func() error) error {
	if !cc.PreInit() {
		return errors.New("pre init fail")
	}
	defer cc.PostInit()

	if fn != nil {
		err := fn()
		if err != nil {
			return err
		}
	}

	err := cc.services.IterWithPredicate(func(name string, service common.Service) error {
		err := service.Init()
		if err != nil {
			return err
		}
		log.Root.Infof("%s init successfully", name)
		return nil
	}, func(name string) bool {
		return name != LogService
	})
	if err != nil {
		return err
	}

	cc.subscriber = event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case topic.SyncState:
			cc.povSyncState.Store(msg)
		case *topic.EventP2PSyncStateMsg:
			cc.p2pSyncState.Store(msg.P2pSyncState)
		case *topic.EventAddP2PStreamMsg:
			if _, ok := cc.connectPeersPool.Load(msg.PeerID); ok {
				cc.connectPeersPool.Delete(msg.PeerID)
			}
			cc.connectPeersPool.Store(msg.PeerID, msg.PeerInfo)
		case *topic.EventDeleteP2PStreamMsg:
			if _, ok := cc.connectPeersPool.Load(msg.PeerID); ok {
				cc.connectPeersPool.Delete(msg.PeerID)
			}
		case *topic.EventBandwidthStats:
			cc.bandwidthStats = msg
		case *topic.EventP2PConnectPeersMsg:
			cc.connectPeersInfo = msg.PeersInfo
		case *topic.EventP2POnlinePeersMsg:
			cc.onlinePeersInfo = msg.PeersInfo
		}
	}), cc.EventBus())

	return cc.subscriber.Subscribe(topic.EventOnlinePeersInfo, topic.EventPeersInfo, topic.EventPovSyncState, topic.EventAddP2PStream, topic.EventDeleteP2PStream, topic.EventSyncStateChange, topic.EventConsensusSyncFinished, topic.EventGetBandwidthStats)
}

func (cc *ChainContext) Start() error {
	if !cc.PreStart() {
		return errors.New("pre start fail")
	}
	defer cc.PostStart()

	cc.services.Iter(func(name string, service common.Service) error {
		err := service.Start()
		if err != nil {
			return fmt.Errorf("%s, %s", name, err)
		}
		log.Root.Infof("%s start successfully", name)
		return nil
	})

	return nil
}

func (cc *ChainContext) Stop() error {
	if !cc.PreStop() {
		return errors.New("pre stop fail")
	}
	defer cc.PostStop()

	cc.services.ReverseIter(func(name string, service common.Service) error {
		err := service.Stop()
		if err != nil {
			return err
		}
		log.Root.Infof("%s stop successfully", name)
		return nil
	})

	if cc.subscriber != nil {
		return cc.subscriber.UnsubscribeAll()
	}

	return nil
}

func (cc *ChainContext) Status() int32 {
	return cc.State()
}

func (cc *ChainContext) SetAccounts(accounts []*types.Account) {
	cc.locker.Lock()
	defer cc.locker.Unlock()
	cc.accounts = accounts
}

func (cc *ChainContext) Accounts() []*types.Account {
	cc.locker.RLock()
	defer cc.locker.RUnlock()
	return cc.accounts
}

func (cc *ChainContext) Id() string {
	return cc.chainID
}

func (cc *ChainContext) Register(name string, service common.Service) error {
	return cc.services.Register(name, service)
}

func (cc *ChainContext) HasService(name string) bool {
	return cc.services.HasService(name)
}

func (cc *ChainContext) UnRegister(name string) error {
	return cc.services.UnRegister(name)
}

func (cc *ChainContext) AllServices() ([]common.Service, error) {
	var services []common.Service
	cc.services.Iter(func(name string, service common.Service) error {
		services = append(services, service)
		return nil
	})
	return services, nil
}

func (cc *ChainContext) WaitForever() {
	count := len(cc.services.services)
	for {
		counter := 0
		cc.services.Iter(func(name string, service common.Service) error {
			if service.Status() == int32(common.Started) {
				counter++
			} else {
				fmt.Println(name, service.Status())
			}
			// return fmt.Errorf("%s, %d", name, service.Status())
			return nil
		})
		if counter == count {
			return
		}
		time.Sleep(time.Duration(50) * time.Millisecond)
	}
}

func (cc *ChainContext) Service(name string) (common.Service, error) {
	return cc.services.Get(name)
}

func (cc *ChainContext) ReloadService(name string) error {
	service, err := cc.Service(name)
	if err != nil {
		return err
	}

	return reloadService(service)
}

func (*ChainContext) RestartAll() error {
	panic("implement me")
}

func (cc *ChainContext) Destroy() error {
	err := cc.Stop()
	if err != nil {
		return err
	}

	id := cc.Id()
	if _, ok := cache.GetStringKey(id); ok {
		cache.Del(id)
	}

	return nil
}

func (cc *ChainContext) ConfigManager(opts ...Option) (*config.CfgManager, error) {
	cc.locker.Lock()
	defer cc.locker.Unlock()
	if cc.cm == nil {
		cc.cm = config.NewCfgManagerWithFile(cc.cfgFile)
		_, err := cc.cm.Load(config.NewMigrationV1ToV2(), config.NewMigrationV2ToV3(), config.NewMigrationV3ToV4(),
			config.NewMigrationV4ToV5(), config.NewMigrationV5ToV6(), config.NewMigrationV6ToV7(), config.NewMigrationV7ToV8(),
			config.NewMigrationV8ToV9(), config.NewMigrationV9ToV10())
		if err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		_ = opt(cc.cm)
	}

	return cc.cm, nil
}

func (cc *ChainContext) Config() (*config.Config, error) {
	cm, err := cc.ConfigManager()
	if err != nil {
		return nil, err
	}
	return cm.Config()
}

func reloadService(s common.Service) error {
	err := s.Stop()
	if err != nil {
		return err
	}

	err = s.Init()
	if err != nil {
		return err
	}

	err = s.Start()
	if err != nil {
		return err
	}
	return nil
}

type serviceContainer struct {
	locker   sync.RWMutex
	services map[string]common.Service
	names    []string
}

func newServiceContainer() *serviceContainer {
	return &serviceContainer{
		locker:   sync.RWMutex{},
		services: make(map[string]common.Service),
		names:    []string{},
	}
}

func (sc *serviceContainer) Register(name string, s common.Service) error {
	sc.locker.Lock()
	defer sc.locker.Unlock()

	if _, ok := sc.services[name]; ok {
		return fmt.Errorf("service[%s] already exist", name)
	} else {
		sc.services[name] = s
		sc.names = append(sc.names, name)
		return nil
	}
}

func (sc *serviceContainer) UnRegister(name string) error {
	sc.locker.Lock()
	defer sc.locker.Unlock()

	if v, ok := sc.services[name]; ok {
		_ = v.Stop()
		delete(sc.services, name)
		for idx, n := range sc.names {
			if n == name {
				sc.names = append(sc.names[:idx], sc.names[idx+1:]...)
				break
			}
		}
		return nil
	} else {
		return fmt.Errorf("service[%s] not exist", name)
	}
}

func (sc *serviceContainer) Get(name string) (common.Service, error) {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	if v, ok := sc.services[name]; ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("service[%s] not exist", name)
	}
}

func (sc *serviceContainer) HasService(name string) bool {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	if _, ok := sc.services[name]; ok {
		return true
	}

	return false
}

func (sc *serviceContainer) Iter(fn func(name string, service common.Service) error) {
	_ = sc.IterWithPredicate(fn, func(name string) bool {
		return true
	})
}

func (sc *serviceContainer) IterWithPredicate(fn func(name string, service common.Service) error,
	predicate func(name string) bool) error {
	sc.locker.RLock()
	defer sc.locker.RUnlock()
	for idx := range sc.names {
		name := sc.names[idx]
		if service, ok := sc.services[name]; ok && predicate(name) {
			err := fn(name, service)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *serviceContainer) ReverseIter(fn func(name string, service common.Service) error) {
	sc.locker.RLock()
	defer sc.locker.RUnlock()

	for i := len(sc.names) - 1; i >= 0; i-- {
		name := sc.names[i]
		if service, ok := sc.services[name]; ok {
			err := fn(name, service)
			if err != nil {
				break
			}
		}
	}
}
