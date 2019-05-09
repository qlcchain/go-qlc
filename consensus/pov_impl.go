package consensus

import (
	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"go.uber.org/zap"
	"time"
)

const (
	blkCacheSize       = 1024
	blkCacheExpireTime = 10 * time.Minute
)

type PoVEngine struct {
	logger   *zap.SugaredLogger
	cfg      *config.Config
	ledger   *ledger.Ledger
	eb       event.EventBus
	accounts []*types.Account

	blkCache gcache.Cache
	bp       *PovBlockProcessor
	txpool   *PovTxPool
	chain    *PovBlockChain
	verifier *process.PovVerifier
	syncer   *PovSyncer
}

func NewPovEngine(cfg *config.Config, accounts []*types.Account) (*PoVEngine, error) {
	ledger := ledger.NewLedger(cfg.LedgerDir())

	pov := &PoVEngine{
		logger:   log.NewLogger("pov_engine"),
		cfg:      cfg,
		eb:       event.GetEventBus(cfg.LedgerDir()),
		accounts: accounts,
		ledger:   ledger,
	}

	pov.blkCache = gcache.New(blkCacheSize).LRU().Expiration(blkCacheExpireTime).Build()
	pov.bp = NewPovBlockProcessor(pov)
	pov.txpool = NewPovTxPool(pov)
	pov.chain = NewPovBlockChain(pov)
	pov.verifier = process.NewPovVerifier(ledger, pov.chain)
	pov.syncer = NewPovSyncer(pov)

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

	pov.syncer.Start()

	pov.setEvent()

	return nil
}

func (pov *PoVEngine) Stop() error {
	pov.unsetEvent()

	pov.syncer.Stop()

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

func (pov *PoVEngine) GetVerifier() *process.PovVerifier {
	return pov.verifier
}

func (pov *PoVEngine) GetAccounts() []*types.Account {
	return pov.accounts
}

func (pov *PoVEngine) GetSyncState() common.SyncState {
	return pov.syncer.getState()
}

func (pov *PoVEngine) GetSyncer() *PovSyncer {
	return pov.syncer
}

func (pov *PoVEngine) AddMinedBlock(block *types.PovBlock) error {
	err := pov.bp.AddMinedBlock(block)
	if err == nil {
		pov.eb.Publish(string(common.EventBroadcast), p2p.PovPublishReq, block)
		pov.blkCache.Set(block.GetHash(), struct{}{})
	}
	return err
}

func (pov *PoVEngine) AddBlock(block *types.PovBlock, from types.PovBlockFrom) error {
	err := pov.bp.AddBlock(block, from)
	return err
}

func (pov *PoVEngine) setEvent() error {
	err := pov.eb.SubscribeAsync(string(common.EventPovRecvBlock), pov.onRecvPovBlock, false)
	if err != nil {
		return err
	}

	err = pov.eb.SubscribeAsync(string(common.EventPovSyncState), pov.onRecvPovSyncState, false)
	if err != nil {
		return err
	}
	return nil
}

func (pov *PoVEngine) unsetEvent() error {
	err := pov.eb.Unsubscribe(string(common.EventPovRecvBlock), pov.onRecvPovBlock)
	if err != nil {
		return err
	}
	err = pov.eb.Unsubscribe(string(common.EventPovSyncState), pov.onRecvPovSyncState)
	if err != nil {
		return err
	}
	return nil
}

func (pov *PoVEngine) onRecvPovBlock(block *types.PovBlock, msgHash types.Hash, msgPeer string) error {
	if pov.blkCache.Has(block.GetHash()) {
		return nil
	}

	pov.logger.Infof("receive block [%s] from [%s]", block.GetHash(), msgPeer)
	err := pov.bp.AddBlock(block, types.PovBlockFromRemoteBroadcast)
	if err == nil {
		pov.eb.Publish(string(common.EventSendMsgToPeers), p2p.PovPublishReq, block, msgPeer)
		pov.blkCache.Set(block.GetHash(), struct{}{})
	}

	return err
}

func (pov *PoVEngine) onRecvPovSyncState(state common.SyncState) {
	if pov.bp != nil {
		pov.bp.onRecvPovSyncState(state)
	}
}
