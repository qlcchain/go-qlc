package consensus

import (
	"fmt"
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
	blkCacheExpireTime = 15 * time.Second
)

type PoVEngine struct {
	logger   *zap.SugaredLogger
	cfg      *config.Config
	ledger   *ledger.Ledger
	eb       event.EventBus
	accounts []*types.Account

	blkRecvCache gcache.Cache

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

	pov.blkRecvCache = gcache.New(blkCacheSize).Simple().Expiration(blkCacheExpireTime).Build()
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
	pov.logger.Info("stop pov engine service")

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
	_ = pov.blkRecvCache.Set(block.GetHash(), struct{}{})
	err := pov.bp.AddMinedBlock(block)
	if err == nil {
		pov.eb.Publish(string(common.EventBroadcast), p2p.PovPublishReq, block)
	}
	return err
}

func (pov *PoVEngine) AddBlock(block *types.PovBlock, from types.PovBlockFrom) error {
	blockHash := block.GetHash()
	if pov.blkRecvCache.Has(blockHash) {
		return fmt.Errorf("block %s already exist in cache", blockHash)
	}
	_ = pov.blkRecvCache.Set(block.GetHash(), struct{}{})

	stat := pov.verifier.VerifyNet(block)
	if stat.Result != process.Progress {
		pov.logger.Infof("block %s verify net err %s", blockHash, stat.ErrMsg)
		return fmt.Errorf("block %s verify net err %s", blockHash, stat.ErrMsg)
	}

	err := pov.bp.AddBlock(block, from)
	return err
}

func (pov *PoVEngine) setEvent() error {
	err := pov.eb.Subscribe(string(common.EventPovRecvBlock), pov.onRecvPovBlock)
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

	return nil
}

func (pov *PoVEngine) onRecvPovBlock(block *types.PovBlock, msgHash types.Hash, msgPeer string) error {
	if pov.blkRecvCache.Has(block.GetHash()) {
		return nil
	}

	pov.logger.Infof("receive block %d/%s from %s", block.GetHeight(), block.GetHash(), msgPeer)

	err := pov.AddBlock(block, types.PovBlockFromRemoteBroadcast)
	if err == nil {
		pov.eb.Publish(string(common.EventSendMsgToPeers), p2p.PovPublishReq, block, msgPeer)
	}

	return err
}
