package consensus

import (
	"errors"
	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
	"math/big"
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

	genesisBlock, err := bc.getLedger().GetPovBlockByHeight(0)
	if genesisBlock == nil {
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
	return nil
}

func (bc *PovBlockChain) loadLastState() error {
	// Make sure the entire head block is available
	currentBlock, _ := bc.getLedger().GetLatestPovBlock()
	if currentBlock == nil {
		// Corrupt or empty database, init from scratch
		bc.logger.Warn("Head block missing, resetting chain")
		return bc.Reset()
	}

	// Everything seems to be fine, set as the head block
	bc.latestBlock.Store(currentBlock)

	bc.logger.Infof("Loaded most recent local full block, number %d hash %s", currentBlock.GetHeight(), currentBlock.GetHash())

	return nil
}

func (bc *PovBlockChain) Reset() error {
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

	bc.logger.Infof("Reset with genesis block, height %d hash %s", genesis.Height, genesis.Hash)

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
	currentBlock := bc.LatestBlock()

	err := bc.getLedger().AddPovBlock(block)
	if err != nil {
		return err
	}

	if block.GetPrevious() == currentBlock.GetHash() {
		if block.GetHeight() != currentBlock.GetHeight() + 1 {
			bc.logger.Errorf("invalid height, currentBlock %s/%d, newBlock %s/%d",
				currentBlock.GetHash(), currentBlock.GetHeight(), block.GetHash(), block.GetHeight())
			return ErrPovInvalidHeight
		}
		return bc.connectBestBlock(block)
	}

	if block.GetHeight() >= currentBlock.GetHeight() + uint64(bc.getConfig().PoV.ForkHeight) {
		return bc.processFork(block)
	}

	return nil
}

func (bc *PovBlockChain) connectBestBlock(block *types.PovBlock) error {
	err := bc.connectBlock(block)
	if err != nil {
		return err
	}

	err = bc.connectTransactions(block)
	if err != nil {
		return err
	}

	bc.latestBlock.Store(block)

	return nil
}

func (bc *PovBlockChain) disconnectBestBlock(block *types.PovBlock) error {
	err := bc.disconnectBlock(block)
	if err != nil {
		return err
	}

	err = bc.disconnectTransactions(block)
	if err != nil {
		return err
	}

	bc.latestBlock.Store(block)

	return nil
}

func (bc *PovBlockChain) processFork(newBlock *types.PovBlock) error {
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

	bc.logger.Infof("find fork block %s/%d, detach %d, attach %d",
		forkBlock.GetHash(), forkBlock.GetHeight(), len(detachBlocks), len(attachBlocks))

	for _, detachBlock := range detachBlocks {
		err := bc.disconnectTransactions(detachBlock)
		if err != nil {
			return err
		}

		err = bc.disconnectBlock(detachBlock)
		if err != nil {
			return err
		}
	}

	for _, attachBlock := range attachBlocks {
		err := bc.connectBestBlock(attachBlock)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bc *PovBlockChain) connectBlock(block *types.PovBlock) error {
	return bc.getLedger().AddPovBestHash(block.GetHeight(), block.GetHash())
}

func (bc *PovBlockChain) disconnectBlock(block *types.PovBlock) error {
	return bc.getLedger().DeletePovBestHash(block.GetHeight())
}

func (bc *PovBlockChain) connectTransactions(block *types.PovBlock) error {
	txpool := bc.getTxPool()

	for _, txPov := range block.Transactions {
		txpool.delTx(txPov.Hash)
	}

	return nil
}

func (bc *PovBlockChain) disconnectTransactions(block *types.PovBlock) error {
	txpool := bc.getTxPool()
	ldger := bc.getLedger()

	for _, txPov := range block.Transactions {
		txBlock, _ := ldger.GetStateBlock(txPov.Hash)
		if txBlock == nil {
			continue
		}

		txpool.addTx(txPov.Hash, txBlock)
	}

	return nil
}

func (bc *PovBlockChain) FindAncestor(block *types.PovBlock, height uint64) *types.PovBlock {
	if height < 0 || height > block.GetHeight() {
		return nil
	}

	for {
		prevBlock := bc.GetBlockByHash(block.GetPrevious())
		if prevBlock == nil {
			return nil
		}

		if prevBlock.GetHeight() == height {
			return prevBlock
		}
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

	firstBlock := bc.RelativeAncestor(block, uint64(bc.getConfig().PoV.TargetCycle - 1))
	if firstBlock == nil {
		return types.Signature{}, ErrPovUnknownAncestor
	}

	actualTimespan := block.Timestamp - firstBlock.Timestamp
	targetTimeSpan := int64(bc.getConfig().PoV.TargetCycle * bc.getConfig().PoV.BlockInterval)

	oldTargetInt := block.Target.ToBigInt()
	nextTargetInt := new(big.Int).Set(oldTargetInt)
	nextTargetInt.Mul(oldTargetInt, big.NewInt(actualTimespan))
	nextTargetInt.Div(nextTargetInt, big.NewInt(targetTimeSpan))

	var nextTarget types.Signature
	nextTarget.FromBigInt(nextTargetInt)

	bc.logger.Debugf("Difficulty retarget at block height %d", block.GetHeight()+1)
	bc.logger.Debugf("Old target %d (%s)", oldTargetInt.BitLen(), oldTargetInt.Text(16))
	bc.logger.Debugf("New target %d (%s)", nextTargetInt.BitLen(), nextTargetInt.Text(16))
	bc.logger.Debugf("Actual timespan %v, target timespan %v",
		time.Duration(actualTimespan)*time.Second, time.Duration(targetTimeSpan)*time.Second)

	return nextTarget, nil
}