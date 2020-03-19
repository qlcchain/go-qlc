package miner

import (
	"time"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
)

type Miner struct {
	cc     *context.ChainContext
	logger *zap.SugaredLogger
	cfg    *config.Config
	eb     event.EventBus
	l      ledger.Store

	chain  PovChainReader
	txPool PovTxPoolReader
	cs     PovConsensusReader

	povWorker *PovWorker
	syncState topic.SyncState
}

type PovChainReader interface {
	LatestHeader() *types.PovHeader
	LatestBlock() *types.PovBlock
	TransitStateDB(height uint64, txs []*types.PovTransaction, gsdb *statedb.PovGlobalStateDB) error
	CalcBlockReward(header *types.PovHeader) (types.Balance, types.Balance, error)
	CalcPastMedianTime(prevHeader *types.PovHeader) uint32
}

type PovTxPoolReader interface {
	SelectPendingTxs(gsdb *statedb.PovGlobalStateDB, limit int) []*types.StateBlock
	LastUpdated() time.Time
	GetPendingTxNum() uint32
}

type PovConsensusReader interface {
	PrepareHeader(header *types.PovHeader) error
}

func NewMiner(cfgFile string, chain PovChainReader, txPool PovTxPoolReader, cs PovConsensusReader) *Miner {
	cc := context.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	miner := &Miner{
		cc:     cc,
		logger: log.NewLogger("miner"),
		cfg:    cfg,
		eb:     cc.EventBus(),

		chain:  chain,
		txPool: txPool,
		cs:     cs,
	}
	miner.l = ledger.NewLedger(cfgFile)

	miner.povWorker = NewPovWorker(cc, miner)

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

	err := miner.povWorker.Start()
	if err != nil {
		return err
	}

	return nil
}

func (miner *Miner) Stop() error {
	miner.logger.Info("stop miner service")

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

func (miner *Miner) GetTxPool() PovTxPoolReader {
	return miner.txPool
}

func (miner *Miner) GetChain() PovChainReader {
	return miner.chain
}

func (miner *Miner) GetPovConsensus() PovConsensusReader {
	return miner.cs
}

func (miner *Miner) GetAccounts() []*types.Account {
	return miner.cc.Accounts()
}

func (miner *Miner) GetSyncState() topic.SyncState {
	return miner.cc.PoVState()
}
