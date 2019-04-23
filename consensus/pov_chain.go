package consensus

import (
	"errors"
	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPovInvalidHash     = errors.New("invalid pov block hash")
	ErrPovNoGenesis       = errors.New("pov genesis block not found in chain")
	ErrPovUnknownAncestor = errors.New("unknown pov block ancestor")
	ErrPovFutureBlock     = errors.New("pov block in the future")
	ErrPovInvalidHeight   = errors.New("invalid pov block height")
	ErrPovInvalidPrevious = errors.New("invalid pov block previous")
)

const (
	blockCacheLimit  = 512
	numberCacheLimit = 2048
)

type PovBlockChain struct {
	povEngine *PoVEngine
	logger    *zap.SugaredLogger

	genesisBlock *types.PovBlock
	latestBlock atomic.Value // Current head of the best block chain

	blockCache   gcache.Cache // Cache for the most recent entire blocks
	heightCache  gcache.Cache

	wg sync.WaitGroup
}

func NewPovBlockChain(povEngine *PoVEngine) *PovBlockChain {
	blockCache := gcache.New(blockCacheLimit).Build()
	heightCache := gcache.New(numberCacheLimit).Build()

	chain := &PovBlockChain{
		povEngine: povEngine,
		logger:    log.NewLogger("pov_chain"),

		blockCache:  blockCache,
		heightCache: heightCache,
	}
	return chain
}

func (bc *PovBlockChain) getLedger() ledger.Store {
	return bc.povEngine.GetLedger()
}

func (bc *PovBlockChain) getConfig() *config.Config {
	return bc.povEngine.GetConfig()
}

func (bc *PovBlockChain) getTxPool() *PovTxPool {
	return bc.povEngine.GetTxPool()
}

func (bc *PovBlockChain) Init() error {
	var err error

	needReset := false
	genesisBlock, err := bc.getLedger().GetPovBlockByHeight(0)
	if genesisBlock == nil {
		needReset = true
	} else if !bc.IsGenesisBlock(genesisBlock) {
		needReset = true
	}

	if needReset {
		genesisBlock = common.GenesisPovBlock()
		err := bc.ResetWithGenesisBlock(genesisBlock)
		if err != nil {
			return err
		}
	} else {
		bc.genesisBlock = genesisBlock
	}

	err = bc.loadLastState()
	if err != nil {
		return err
	}

	return nil
}

func (bc *PovBlockChain) Start() error {
	return nil
}

func (bc *PovBlockChain) Stop() error {
	bc.wg.Wait()
	return nil
}

func (bc *PovBlockChain) loadLastState() error {
	// Make sure the entire head block is available
	latestBlock, _ := bc.getLedger().GetLatestPovBlock()
	if latestBlock == nil {
		// Corrupt or empty database, init from scratch
		bc.logger.Warn("Head block missing, resetting chain")
		return bc.ResetChainState()
	}

	// Everything seems to be fine, set as the head block
	bc.latestBlock.Store(latestBlock)

	bc.logger.Infof("Loaded latest block %d/%s", latestBlock.GetHeight(), latestBlock.GetHash())

	return nil
}

func (bc *PovBlockChain) ResetChainState() error {
	_ = bc.getLedger().DropAllPovBlocks()

	return bc.ResetWithGenesisBlock(bc.genesisBlock)
}

func (bc *PovBlockChain) ResetWithGenesisBlock(genesis *types.PovBlock) error {
	// Prepare the genesis block and reinitialise the chain
	err := bc.getLedger().AddPovBlock(genesis)
	if err != nil {
		return err
	}
	err = bc.getLedger().AddPovBestHash(genesis.GetHeight(), genesis.GetHash())
	if err != nil {
		return err
	}

	bc.genesisBlock = genesis

	bc.logger.Infof("Reset with genesis block %d/%s", genesis.Height, genesis.Hash)

	bc.latestBlock.Store(genesis)

	return nil
}

func (bc *PovBlockChain) GenesisBlock() *types.PovBlock {
	return bc.genesisBlock
}

func (bc *PovBlockChain) LatestBlock() *types.PovBlock {
	return bc.latestBlock.Load().(*types.PovBlock)
}

func (bc *PovBlockChain) IsGenesisBlock(block *types.PovBlock) bool {
	return common.IsGenesisPovBlock(block)
}

func (bc *PovBlockChain) GetBlockByHeight(height uint64) (*types.PovBlock, error) {
	v, _ := bc.heightCache.Get(height)
	if v != nil {
		return v.(*types.PovBlock), nil
	}

	block, err := bc.povEngine.GetLedger().GetPovBlockByHeight(height)
	if block != nil {
		bc.heightCache.Set(height, block)
		bc.blockCache.Set(block.GetHash(), block)
	}
	return block, err
}

func (bc *PovBlockChain) GetBlockHeight(hash types.Hash) uint64 {
	v, _ := bc.heightCache.Get(hash)
	if v != nil {
		return v.(uint64)
	}

	block, _ := bc.povEngine.GetLedger().GetPovBlockByHash(hash)
	if block != nil {
		bc.heightCache.Set(hash, block.GetHeight())
		return block.GetHeight()
	}

	return 0
}

func (bc *PovBlockChain) GetBlockByHash(hash types.Hash) *types.PovBlock {
	v, _ := bc.blockCache.Get(hash)
	if v != nil {
		return v.(*types.PovBlock)
	}

	block, _ := bc.povEngine.GetLedger().GetPovBlockByHash(hash)
	if block != nil {
		bc.blockCache.Set(hash, block)
		return block
	}

	return nil
}

func (bc *PovBlockChain) HasBlock(hash types.Hash, number uint64) bool {
	if bc.blockCache.Has(hash) {
		return true
	}
	return bc.getLedger().HasPovBody(number, hash)
}

func (bc *PovBlockChain) InsertBlock(block *types.PovBlock) error {
	err := bc.getLedger().BatchUpdate(func(txn db.StoreTxn) error {
		return bc.insertBlock(txn, block)
	})

	return err
}

func (bc *PovBlockChain) insertBlock(txn db.StoreTxn, block *types.PovBlock) error {
	currentBlock := bc.LatestBlock()

	err := bc.getLedger().AddPovBlock(block, txn)
	if err != nil {
		bc.logger.Errorf("add pov block %d/%s failed, err %s", block.Height, block.Hash, err)
		return err
	}

	if block.GetPrevious() == currentBlock.GetHash() {
		if block.GetHeight() != currentBlock.GetHeight() + 1 {
			bc.logger.Errorf("invalid height, currentBlock %s/%d, newBlock %s/%d",
				currentBlock.GetHash(), currentBlock.GetHeight(), block.GetHash(), block.GetHeight())
			return ErrPovInvalidHeight
		}
		return bc.connectBestBlock(txn, block)
	}

	if block.GetHeight() >= currentBlock.GetHeight() + uint64(bc.getConfig().PoV.ForkHeight) {
		return bc.processFork(txn, block)
	}

	return nil
}

func (bc *PovBlockChain) connectBestBlock(txn db.StoreTxn, block *types.PovBlock) error {
	err := bc.connectBlock(txn, block)
	if err != nil {
		return err
	}

	err = bc.connectTransactions(txn, block)
	if err != nil {
		return err
	}

	bc.latestBlock.Store(block)

	bc.heightCache.Set(block.GetHash(), block)
	bc.blockCache.Set(block.GetHeight(), block)

	return nil
}

func (bc *PovBlockChain) disconnectBestBlock(txn db.StoreTxn, block *types.PovBlock) error {
	err := bc.disconnectBlock(txn, block)
	if err != nil {
		return err
	}

	err = bc.disconnectTransactions(txn, block)
	if err != nil {
		return err
	}

	bc.latestBlock.Store(block)

	return nil
}

func (bc *PovBlockChain) processFork(txn db.StoreTxn, newBlock *types.PovBlock) error {
	oldBlock := bc.LatestBlock()

	var detachBlocks []*types.PovBlock
	var attachBlocks []*types.PovBlock
	var forkHash types.Hash

	attachBlock := newBlock
	for height := attachBlock.GetHeight(); height > oldBlock.GetHeight(); height-- {
		attachBlocks = append(attachBlocks, attachBlock)

		attachBlock = bc.GetBlockByHash(attachBlock.GetPrevious())
		if attachBlock == nil {
			return ErrPovInvalidPrevious
		}
	}

	detachBlock := oldBlock

	for {
		if attachBlock == nil || detachBlock == nil {
			return ErrPovInvalidPrevious
		}

		attachBlocks = append(attachBlocks, attachBlock)
		detachBlocks = append(detachBlocks, detachBlock)

		if attachBlock.GetPrevious() == detachBlock.GetPrevious() {
			forkHash = attachBlock.GetPrevious()
			break
		}

		attachBlock = bc.GetBlockByHash(attachBlock.GetPrevious())
		if attachBlock == nil {
			break
		}

		detachBlock = bc.GetBlockByHash(detachBlock.GetPrevious())
		if detachBlock == nil {
			break
		}
	}

	forkBlock := bc.GetBlockByHash(forkHash)
	if forkBlock == nil {
		return ErrPovInvalidHash
	}

	bc.logger.Infof("find fork block %d/%s, detach %d, attach %d",
		forkBlock.GetHeight(), forkBlock.GetHash(), len(detachBlocks), len(attachBlocks))

	for _, detachBlock := range detachBlocks {
		err := bc.disconnectTransactions(txn, detachBlock)
		if err != nil {
			return err
		}

		err = bc.disconnectBlock(txn, detachBlock)
		if err != nil {
			return err
		}
	}

	for _, attachBlock := range attachBlocks {
		err := bc.connectBestBlock(txn, attachBlock)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bc *PovBlockChain) connectBlock(txn db.StoreTxn, block *types.PovBlock) error {
	err := bc.getLedger().AddPovBestHash(block.GetHeight(), block.GetHash(), txn)
	if err != nil {
		bc.logger.Errorf("add pov best hash %d/%s failed, err %s", block.Height, block.Hash, err)
		return err
	}
	return nil
}

func (bc *PovBlockChain) disconnectBlock(txn db.StoreTxn, block *types.PovBlock) error {
	err := bc.getLedger().DeletePovBestHash(block.GetHeight(), txn)
	if err != nil {
		bc.logger.Errorf("delete pov best hash %d/%s failed, err %s", block.Height, block.Hash, err)
		return err
	}
	return nil
}

func (bc *PovBlockChain) connectTransactions(txn db.StoreTxn, block *types.PovBlock) error {
	txpool := bc.getTxPool()
	ledger := bc.getLedger()

	for txIndex, txPov := range block.Transactions {
		txpool.delTx(txPov.Hash)

		txLookup := &types.PovTxLookup{
			BlockHash: block.GetHash(),
			BlockHeight: block.GetHeight(),
			TxIndex: uint64(txIndex),
		}
		ledger.AddPovTxLookup(txPov.Hash, txLookup, txn)
	}

	return nil
}

func (bc *PovBlockChain) disconnectTransactions(txn db.StoreTxn, block *types.PovBlock) error {
	txpool := bc.getTxPool()
	ledger := bc.getLedger()

	for _, txPov := range block.Transactions {
		txBlock, _ := ledger.GetStateBlock(txPov.Hash, txn)
		if txBlock == nil {
			continue
		}

		txpool.addTx(txPov.Hash, txBlock)

		ledger.DeletePovTxLookup(txPov.Hash)
	}

	return nil
}

func (bc *PovBlockChain) FindAncestor(block *types.PovBlock, height uint64) *types.PovBlock {
	if height < 0 || height > block.GetHeight() {
		return nil
	}

	curBlock := block
	for {
		prevBlock := bc.GetBlockByHash(curBlock.GetPrevious())
		if prevBlock == nil {
			return nil
		}

		if prevBlock.GetHeight() == height {
			return prevBlock
		}

		curBlock = prevBlock
	}
}

func (bc *PovBlockChain) RelativeAncestor(block *types.PovBlock, distance uint64) *types.PovBlock {
	return bc.FindAncestor(block, block.GetHeight() - distance)
}

func (bc *PovBlockChain) CalcNextRequiredTarget(block *types.PovBlock) (types.Signature, error) {
	if (block.GetHeight() + 1) % uint64(bc.getConfig().PoV.TargetCycle) != 0 {
		return block.Target, nil
	}

	// nextTarget = prevTarget * (lastBlock.Timestamp - firstBlock.Timestamp) / (blockInterval * targetCycle)

	distance := uint64(bc.getConfig().PoV.TargetCycle - 1)
	firstBlock := bc.RelativeAncestor(block, distance)
	if firstBlock == nil {
		bc.logger.Infof("failed to get relative ancestor at height %d distance %d", block.GetHeight(), distance)
		return types.ZeroSignature, ErrPovUnknownAncestor
	}

	actualTimespan := block.Timestamp - firstBlock.Timestamp
	targetTimeSpan := int64(bc.getConfig().PoV.TargetCycle * bc.getConfig().PoV.BlockInterval)

	oldTargetInt := block.Target.ToBigInt()
	nextTargetInt := new(big.Int).Set(oldTargetInt)
	nextTargetInt.Mul(oldTargetInt, big.NewInt(actualTimespan))
	nextTargetInt.Div(nextTargetInt, big.NewInt(targetTimeSpan))

	var nextTarget types.Signature
	nextTarget.FromBigInt(nextTargetInt)

	bc.logger.Infof("Difficulty retarget at block height %d", block.GetHeight()+1)
	bc.logger.Infof("Old target %d (%s)", oldTargetInt.BitLen(), oldTargetInt.Text(16))
	bc.logger.Infof("New target %d (%s)", nextTargetInt.BitLen(), nextTargetInt.Text(16))
	bc.logger.Infof("Actual timespan %v, target timespan %v",
		time.Duration(actualTimespan)*time.Second, time.Duration(targetTimeSpan)*time.Second)

	return nextTarget, nil
}