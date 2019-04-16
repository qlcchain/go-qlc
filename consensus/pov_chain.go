package consensus

import (
	"errors"
	lru "github.com/hashicorp/golang-lru"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
	"sync/atomic"
)

var (
	ErrNoGenesis = errors.New("Genesis not found in chain")

	// ErrUnknownAncestor is returned when validating a block requires an ancestor
	// that is unknown.
	ErrUnknownAncestor = errors.New("unknown ancestor")

	// ErrPrunedAncestor is returned when validating a block requires an ancestor
	// that is known, but the state of which is not available.
	ErrPrunedAncestor = errors.New("pruned ancestor")

	// ErrFutureBlock is returned when a block's timestamp is in the future according
	// to the current node.
	ErrFutureBlock = errors.New("block in the future")

	// ErrInvalidNumber is returned if a block's number doesn't equal it's parent's
	// plus one.
	ErrInvalidNumber = errors.New("invalid block number")

	// ErrInvalidHash is returned if a block's previous doesn't equal it's parent's
	// hash.
	ErrInvalidPrevious = errors.New("invalid block previous")
)

// WriteStatus status of write
type WriteStatus byte

const (
	NonStatTy WriteStatus = iota
	CanonStatTy
	SideStatTy
)

const (
	bodyCacheLimit      = 256
	blockCacheLimit     = 256
	maxFutureBlocks     = 256
	maxTimeFutureBlocks = 30 * 2
)

type PovBlockChain struct {
	povEngine *PoVEngine
	logger *zap.SugaredLogger

	genesisBlock *types.PovBlock
	latestBlock  *types.PovBlock

	currentBlock atomic.Value // Current head of the block chain
	bodyCache     *lru.Cache     // Cache for the most recent block bodies
	blockCache    *lru.Cache     // Cache for the most recent entire blocks
}

func NewPovBlockChain(povEngine *PoVEngine) *PovBlockChain {
	bodyCache, _ := lru.New(bodyCacheLimit)
	blockCache, _ := lru.New(blockCacheLimit)

	chain := &PovBlockChain{
		povEngine: povEngine,
		logger: log.NewLogger("pov_blockchain"),

		bodyCache:      bodyCache,
		blockCache:     blockCache,
	}
	return chain
}

func (bc *PovBlockChain) getLedger() ledger.Store {
	return bc.povEngine.GetLedger()
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

	bc.logger.Info("Loaded most recent local full block", "number", currentBlock.GetHeight(), "hash", currentBlock.GetHash())

	return nil
}

func (bc *PovBlockChain) Reset() error {
	return bc.ResetWithGenesisBlock(bc.genesisBlock)
}

func (bc *PovBlockChain) ResetWithGenesisBlock(genesis *types.PovBlock) error {
	// Prepare the genesis block and reinitialise the chain
	err := bc.getLedger().AddPovBlock(genesis)
	if err != nil {
		return err
	}

	bc.genesisBlock = genesis

	bc.logger.Infof("Reset with genesis block, height %d hash %s", genesis.Height, genesis.Hash)

	bc.latestBlock = genesis

	return nil
}

func (bc *PovBlockChain) Start() error {
	return nil
}

func (bc *PovBlockChain) Stop() error {
	return nil
}

func (bc *PovBlockChain) GetLatestBlock() *types.PovBlock {
	return bc.latestBlock
}

func (bc *PovBlockChain) GetGenesisBlock() *types.PovBlock {
	return bc.genesisBlock
}

func (bc *PovBlockChain) IsGenesisBlock(block *types.PovBlock) bool {
	return common.IsGenesisPovBlock(block)
}

func (bc *PovBlockChain) GetBlockByHeight(height uint64) (*types.PovBlock, error) {
	block, err := bc.povEngine.GetLedger().GetPovBlockByHeight(height)
	return block, err
}

func (bc *PovBlockChain) GetBlockByHash(hash types.Hash) (*types.PovBlock, error) {
	block, err := bc.povEngine.GetLedger().GetPovBlockByHash(hash)
	return block, err
}

// CurrentBlock retrieves the current head block of the canonical chain. The
// block is retrieved from the blockchain's internal cache.
func (bc *PovBlockChain) CurrentBlock() *types.PovBlock {
	return bc.currentBlock.Load().(*types.PovBlock)
}

func (bc *PovBlockChain) InsertBlock(block *types.PovBlock) error {
	return nil
}

func (bc *PovBlockChain) processFork(oldBlock *types.PovBlock, newBlock *types.PovBlock) error {
	return nil
}