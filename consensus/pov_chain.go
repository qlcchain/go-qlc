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
	"sync/atomic"
)

var (
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
	currentBlock atomic.Value // Current head of the block chain

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

func (bc *PovBlockChain) loadLastState() error {
	// Make sure the entire head block is available
	currentBlock, _ := bc.getLedger().GetLatestPovBlock()
	if currentBlock == nil {
		// Corrupt or empty database, init from scratch
		bc.logger.Warn("Head block missing, resetting chain")
		return bc.Reset()
	}

	// Everything seems to be fine, set as the head block
	bc.currentBlock.Store(currentBlock)

	bc.logger.Infof("Loaded most recent local full block, number %d hash %s", currentBlock.GetHeight(), currentBlock.GetHash())

	return nil
}

func (bc *PovBlockChain) Reset() error {
	return bc.ResetWithGenesisBlock(bc.genesisBlock)
}

func (bc *PovBlockChain) ResetWithGenesisBlock(genesis *types.PovBlock) error {
	bc.getLedger().DeletePovBlock(genesis)
	bc.getLedger().DeletePovBestHash(genesis.GetHeight())

	// Prepare the genesis block and reinitialise the chain
	err := bc.getLedger().AddPovBlock(genesis)
	if err != nil {
		return err
	}
	bc.getLedger().AddPovBestHash(genesis.GetHeight(), genesis.GetHash())

	bc.genesisBlock = genesis

	bc.logger.Infof("Reset with genesis block, height %d hash %s", genesis.Height, genesis.Hash)

	bc.currentBlock.Store(genesis)

	return nil
}

func (bc *PovBlockChain) Start() error {
	return nil
}

func (bc *PovBlockChain) Stop() error {
	return nil
}

func (bc *PovBlockChain) GenesisBlock() *types.PovBlock {
	return bc.genesisBlock
}

func (bc *PovBlockChain) CurrentBlock() *types.PovBlock {
	return bc.currentBlock.Load().(*types.PovBlock)
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

func (bc *PovBlockChain) insert(block *types.PovBlock) {
	// Add the block to the canonical chain number scheme and mark as the head
	bc.getLedger().AddPovBestHash(block.GetHeight(), block.GetHash())

	bc.currentBlock.Store(block)
}

func (bc *PovBlockChain) InsertBlock(block *types.PovBlock) {
	currentBlock := bc.CurrentBlock()

	bc.getLedger().AddPovBlock(block)

	if block.GetPrevious() == currentBlock.GetHash() {

	}

	if block.GetHeight() >= currentBlock.GetHeight() + uint64(bc.getConfig().PoV.ForkHeight) {

	}
}
