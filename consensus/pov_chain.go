package consensus

import (
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	blockCacheLimit = 1024
	medianTimeBlocks = 11
)

var (
	ErrPovInvalidHash     = errors.New("invalid pov block hash")
	ErrPovNoGenesis       = errors.New("pov genesis block not found in chain")
	ErrPovUnknownAncestor = errors.New("unknown pov block ancestor")
	ErrPovFutureBlock     = errors.New("pov block in the future")
	ErrPovInvalidHeight   = errors.New("invalid pov block height")
	ErrPovInvalidPrevious = errors.New("invalid pov block previous")
	ErrPovFailedVerify    = errors.New("failed to verify block")
	ErrPovForkHashZero    = errors.New("fork point hash is zero")
	ErrPovInvalidHead     = errors.New("invalid pov head block")
	ErrPovInvalidFork     = errors.New("invalid pov fork point")
)

var (
	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// oneLsh512 is 1 shifted left 512 bits.  It is defined here to avoid
	// the overhead of creating it multiple times.
	oneLsh512 = new(big.Int).Lsh(bigOne, 512)
)

// log2FloorMasks defines the masks to use when quickly calculating
// floor(log2(x)) in a constant log2(32) = 5 steps, where x is a uint32, using
// shifts.  They are derived from (2^(2^x) - 1) * (2^(2^x)), for x in 4..0.
var log2FloorMasks = []uint32{0xffff0000, 0xff00, 0xf0, 0xc, 0x2}

// fastLog2Floor calculates and returns floor(log2(x)) in a constant 5 steps.
func fastLog2Floor(n uint32) uint8 {
	rv := uint8(0)
	exponent := uint8(16)
	for i := 0; i < 5; i++ {
		if n&log2FloorMasks[i] != 0 {
			rv += exponent
			n >>= exponent
		}
		exponent >>= 1
	}
	return rv
}

type PovBlockChain struct {
	povEngine *PoVEngine
	logger    *zap.SugaredLogger

	genesisBlock *types.PovBlock
	latestBlock  atomic.Value // Current head of the best block chain

	blockCache   gcache.Cache // Cache for the most recent entire blocks
	heightCache  gcache.Cache
	tdCache      gcache.Cache
	trieNodePool *trie.NodePool

	wg sync.WaitGroup
}

func NewPovBlockChain(povEngine *PoVEngine) *PovBlockChain {
	blockCache := gcache.New(blockCacheLimit).Build()
	heightCache := gcache.New(blockCacheLimit).Build()
	tdCache := gcache.New(blockCacheLimit).Build()

	chain := &PovBlockChain{
		povEngine: povEngine,
		logger:    log.NewLogger("pov_chain"),

		blockCache:  blockCache,
		heightCache: heightCache,
		tdCache:     tdCache,
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

	bc.trieNodePool = trie.NewSimpleTrieNodePool()

	needReset := false
	genesisBlock, err := bc.getLedger().GetPovBlockByHeight(0)
	if genesisBlock == nil {
		needReset = true
	} else if !bc.IsGenesisBlock(genesisBlock) {
		needReset = true
	}

	if needReset {
		err := bc.ResetChainState()
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
	bc.trieNodePool.Clear()
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

	genesisBlock := common.GenesisPovBlock()

	return bc.resetWithGenesisBlock(&genesisBlock)
}

func (bc *PovBlockChain) resetWithGenesisBlock(genesis *types.PovBlock) error {
	var saveCallback func()

	err := bc.getLedger().BatchUpdate(func(txn db.StoreTxn) error {
		var dbErr error

		td := bc.CalcTotalDifficulty(genesis.Target)
		dbErr = bc.getLedger().AddPovBlock(genesis, td)
		if dbErr != nil {
			return dbErr
		}
		dbErr = bc.getLedger().AddPovBestHash(genesis.GetHeight(), genesis.GetHash())
		if dbErr != nil {
			return dbErr
		}

		stateTrie := trie.NewTrie(bc.getLedger().DBStore(), nil, bc.trieNodePool)
		stateKeys, stateValues := common.GenesisPovStateKVs()
		for i := range stateKeys {
			stateTrie.SetValue(stateKeys[i], stateValues[i])
		}

		if *stateTrie.Hash() != genesis.GetStateHash() {
			panic(fmt.Sprintf("genesis block state hash not same %s != %s", *stateTrie.Hash(), genesis.GetStateHash()))
		}

		saveCallback, dbErr = stateTrie.SaveInTxn(txn)
		if dbErr != nil {
			return dbErr
		}
		saveCallback()

		return nil
	})

	if err != nil {
		bc.logger.Fatalf("failed to reset with genesis block")
		return err
	}

	if saveCallback != nil {
		saveCallback()
	}

	bc.genesisBlock = genesis
	bc.latestBlock.Store(genesis)

	bc.logger.Infof("reset with genesis block %d/%s", genesis.Height, genesis.Hash)

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
		bc.blockCache.Set(block.GetHash(), block)
		return block.GetHeight()
	}

	return 0
}

func (bc *PovBlockChain) GetBlockByHash(hash types.Hash) *types.PovBlock {
	if hash.IsZero() {
		return nil
	}

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

func (bc *PovBlockChain) HasBlock(hash types.Hash, height uint64) bool {
	if bc.blockCache.Has(hash) {
		return true
	}

	if bc.getLedger().HasPovBlock(height, hash) {
		return true
	}

	return false
}

func (bc *PovBlockChain) GetBestBlockByHash(hash types.Hash) *types.PovBlock {
	block := bc.GetBlockByHash(hash)
	if block == nil {
		return nil
	}

	bestHashInDB, err := bc.getLedger().GetPovBestHash(block.GetHeight())
	if err != nil {
		return nil
	}
	if bestHashInDB != hash {
		return nil
	}

	return block
}

func (bc *PovBlockChain) HasBestBlock(hash types.Hash, height uint64) bool {
	if !bc.HasBlock(hash, height) {
		return false
	}

	bestHashInDB, err := bc.getLedger().GetPovBestHash(height)
	if err != nil {
		return false
	}
	if bestHashInDB != hash {
		return false
	}

	return true
}

func (bc *PovBlockChain) GetBlockTDByHash(hash types.Hash) *big.Int {
	blk := bc.GetBlockByHash(hash)
	if blk == nil {
		return nil
	}

	return bc.GetBlockTDByHashAndHeight(blk.GetHash(), blk.GetHeight())
}

func (bc *PovBlockChain) GetBlockTDByHashAndHeight(hash types.Hash, height uint64) *big.Int {
	v, _ := bc.tdCache.Get(hash)
	if v != nil {
		return v.(*big.Int)
	}

	td, err := bc.getLedger().GetPovTD(hash, height)
	if err != nil {
		return nil
	}

	bc.tdCache.Set(hash, td)
	return td
}

func (bc *PovBlockChain) GetBlockLocator(hash types.Hash) []types.Hash {
	var block *types.PovBlock

	if hash.IsZero() {
		hash = bc.LatestBlock().GetHash()
		block = bc.LatestBlock()
	} else {
		block = bc.GetBlockByHash(hash)
	}
	if block == nil {
		return nil
	}

	// Calculate the max number of entries that will ultimately be in the
	// block locator.  See the description of the algorithm for how these
	// numbers are derived.
	var maxEntries uint8
	if block.GetHeight() <= 12 {
		maxEntries = uint8(block.GetHeight()) + 1
	} else {
		// Requested hash itself + previous 10 entries + genesis block.
		// Then floor(log2(height-10)) entries for the skip portion.
		adjustedHeight := uint32(block.GetHeight()) - 10
		maxEntries = 12 + fastLog2Floor(adjustedHeight)
	}
	locator := make([]types.Hash, 0, maxEntries)

	step := uint64(1)
	for block != nil {
		locator = append(locator, block.GetHash())

		// Nothing more to add once the genesis block has been added.
		if block.GetHeight() == 0 {
			break
		}

		// Calculate height of previous node to include ensuring the
		// final node is the genesis block.
		height := block.GetHeight() - step
		if height < 0 {
			height = 0
		}

		block = bc.FindAncestor(block, height)

		// Once 11 entries have been included, start doubling the
		// distance between included hashes.
		if len(locator) > 10 {
			step *= 2
		}
	}

	return locator
}

func (bc *PovBlockChain) LocateBestBlock(locator []types.Hash) *types.PovBlock {
	startBlock := bc.GenesisBlock()

	for _, locHash := range locator {
		block := bc.GetBestBlockByHash(locHash)
		if block != nil {
			startBlock = block
			break
		}
	}

	return startBlock
}

func (bc *PovBlockChain) InsertBlock(block *types.PovBlock, stateTrie *trie.Trie) error {
	err := bc.getLedger().BatchUpdate(func(txn db.StoreTxn) error {
		return bc.insertBlock(txn, block, stateTrie)
	})

	if err != nil {
		bc.logger.Errorf("failed to insert block %d/%s to chain", block.GetHeight(), block.GetHash())
	} else {
		bc.logger.Infof("success to insert block %d/%s to chain", block.GetHeight(), block.GetHash())
	}

	return err
}

func (bc *PovBlockChain) insertBlock(txn db.StoreTxn, block *types.PovBlock, stateTrie *trie.Trie) error {
	currentBlock := bc.LatestBlock()

	bestTD, err := bc.getLedger().GetPovTD(currentBlock.GetHash(), currentBlock.GetHeight(), txn)
	if err != nil {
		bc.logger.Errorf("get pov best td %d/%s failed, err %s", currentBlock.GetHeight(), currentBlock.GetHash(), err)
		return err
	}

	prevTD, err := bc.getLedger().GetPovTD(block.GetPrevious(), block.GetHeight()-1, txn)
	if err != nil {
		bc.logger.Errorf("get pov previous td %d/%s failed, err %s", block.GetHeight()-1, block.GetPrevious(), err)
		return err
	}

	blockTarget := block.GetTarget()
	blockTD := new(big.Int).Add(blockTarget.ToBigInt(), prevTD)

	// save block to db
	if bc.getLedger().HasPovBlock(block.GetHeight(), block.GetHash(), txn) == false {
		err = bc.getLedger().AddPovBlock(block, blockTD, txn)
		if err != nil && err != ledger.ErrBlockExists {
			bc.logger.Errorf("add pov block %d/%s failed, err %s", block.Height, block.Hash, err)
			return err
		}
	}

	saveCallback, dbErr := stateTrie.SaveInTxn(txn)
	if dbErr != nil {
		return dbErr
	}
	if saveCallback != nil {
		saveCallback()
	}

	bc.tdCache.Set(block.GetHash(), blockTD)

	tdCmpRet := blockTD.Cmp(bestTD)
	isBest := tdCmpRet > 0
	if !isBest && tdCmpRet == 0 {
		if block.GetHeight() < currentBlock.GetHeight() {
			isBest = true
		} else if block.GetHeight() == currentBlock.GetHeight() {
			// If the total difficulty is higher than our known, add it to the canonical chain
			// Second clause in the if statement reduces the vulnerability to selfish mining.
			// Please refer to http://www.cs.cornell.edu/~ie53/publications/btcProcFC.pdf
			if rand.Float64() < 0.5 {
				isBest = true
			}
		}
	}

	// check fork side chain
	if isBest {
		// try to grow best chain
		if block.GetPrevious() == currentBlock.GetHash() {
			return bc.connectBestBlock(txn, block)
		}

		bc.logger.Infof("block %d/%s td %d/%s, need to doing fork, prev %s",
			block.GetHeight(), block.GetHash(), blockTD.BitLen(), blockTD.Text(16), block.GetPrevious())
		return bc.processFork(txn, block)
	} else {
		bc.logger.Debugf("block %d/%s td %d/%s in side chain, prev %s",
			block.GetHeight(), block.GetHash(), blockTD.BitLen(), blockTD.Text(16), block.GetPrevious())
	}

	return nil
}

func (bc *PovBlockChain) connectBestBlock(txn db.StoreTxn, block *types.PovBlock) error {
	latestBlock := bc.LatestBlock()
	if block.GetHeight() != latestBlock.GetHeight()+1 {
		return ErrPovInvalidHeight
	}

	err := bc.connectBlock(txn, block)
	if err != nil {
		return err
	}

	err = bc.connectTransactions(txn, block)
	if err != nil {
		return err
	}

	bc.latestBlock.Store(block)

	_ = bc.heightCache.Set(block.GetHeight(), block)

	return nil
}

func (bc *PovBlockChain) disconnectBestBlock(txn db.StoreTxn, block *types.PovBlock) error {
	latestBlock := bc.LatestBlock()
	if latestBlock.GetHash() != block.GetHash() {
		return ErrPovInvalidHash
	}

	err := bc.disconnectTransactions(txn, block)
	if err != nil {
		return err
	}

	err = bc.disconnectBlock(txn, block)
	if err != nil {
		return err
	}

	prevBlock := bc.GetBlockByHash(block.GetPrevious())
	if prevBlock == nil {
		return ErrPovInvalidPrevious
	}

	bc.heightCache.Remove(block.GetHeight())

	bc.latestBlock.Store(prevBlock)

	_ = bc.heightCache.Set(prevBlock.GetHeight(), prevBlock)

	return nil
}

func (bc *PovBlockChain) processFork(txn db.StoreTxn, newBlock *types.PovBlock) error {
	oldHeadBlock := bc.LatestBlock()
	if oldHeadBlock == nil {
		return ErrPovInvalidHead
	}

	bc.logger.Infof("before fork process, head %d/%s", oldHeadBlock.GetHeight(), oldHeadBlock.GetHash())

	var detachBlocks []*types.PovBlock
	var attachBlocks []*types.PovBlock
	var forkHash types.Hash

	attachBlock := newBlock
	if newBlock.GetHeight() > oldHeadBlock.GetHeight() {
		//old: b1 <- b2 <- b3 <- b4 <- b5 <- b6-1 <- b7-1
		//new:                            <- b6-2 <- b7-2 <- b8 <- b9 <- b10

		// step1: attach b7-2(exclude) <- b10
		for height := attachBlock.GetHeight(); height > oldHeadBlock.GetHeight(); height-- {
			attachBlocks = append(attachBlocks, attachBlock)

			prevHash := attachBlock.GetPrevious()
			if prevHash.IsZero() {
				break
			}

			attachBlock = bc.GetBlockByHash(prevHash)
			if attachBlock == nil {
				bc.logger.Errorf("failed to get previous attach block %s", prevHash)
				return ErrPovInvalidPrevious
			}
		}
	}

	detachBlock := oldHeadBlock
	if oldHeadBlock.GetHeight() > newBlock.GetHeight() {
		//old: b1 <- b2 <- b3 <- b4 <- b5 <- b6-1 <- b7-1 <- b8 <- b9 <- b10
		//new:                            <- b6-2 <- b7-2

		// step1: detach b7-1(exclude) <- b10
		for height := detachBlock.GetHeight(); height > newBlock.GetHeight(); height-- {
			detachBlocks = append(detachBlocks, detachBlock)

			prevHash := detachBlock.GetPrevious()
			if prevHash.IsZero() {
				break
			}

			detachBlock = bc.GetBlockByHash(prevHash)
			if detachBlock == nil {
				bc.logger.Errorf("failed to get previous detach block %s", prevHash)
				return ErrPovInvalidPrevious
			}
		}
	}

	//old: b1 <- b2 <- b3 <- b4 <- b5 <- b6-1 <- b7-1
	//new:                            <- b6-2 <- b7-2

	if attachBlock.GetHeight() != detachBlock.GetHeight() {
		bc.logger.Errorf("height not equal, attach %d detach %d", attachBlock.GetHeight(), detachBlock.GetHeight())
		return ErrPovInvalidHeight
	}

	// step2: attach b6-2 <- b7-2(include)
	//        detach b6-1 <- b7-1(include)
	//        fork point: b5
	foundForkPoint := false
	for {
		attachBlocks = append(attachBlocks, attachBlock)
		detachBlocks = append(detachBlocks, detachBlock)

		if attachBlock.GetPrevious() == detachBlock.GetPrevious() {
			foundForkPoint = true
			forkHash = attachBlock.GetPrevious()
			break
		}

		prevAttachHash := attachBlock.GetPrevious()
		attachBlock = bc.GetBlockByHash(prevAttachHash)
		if attachBlock == nil {
			bc.logger.Errorf("failed to get previous attach block %s", prevAttachHash)
			return ErrPovInvalidPrevious
		}

		prevDetachHash := detachBlock.GetPrevious()
		detachBlock = bc.GetBlockByHash(prevDetachHash)
		if detachBlock == nil {
			bc.logger.Errorf("failed to get previous detach block %s", prevDetachHash)
			return ErrPovInvalidPrevious
		}
	}

	if !foundForkPoint {
		bc.logger.Errorf("failed to find fork point")
		return ErrPovInvalidFork
	}

	var forkBlock *types.PovBlock
	if forkHash.IsZero() {
		forkBlock = bc.GenesisBlock()
	} else {
		forkBlock = bc.GetBlockByHash(forkHash)
	}
	if forkBlock == nil {
		bc.logger.Errorf("fork block %s not exist", forkHash)
		return ErrPovInvalidHash
	}

	bc.logger.Infof("find fork block %d/%s, detach %d, attach %d",
		forkBlock.GetHeight(), forkBlock.GetHash(), len(detachBlocks), len(attachBlocks))

	// detach from high to low
	for _, detachBlock := range detachBlocks {
		err := bc.disconnectBestBlock(txn, detachBlock)
		if err != nil {
			return err
		}
	}

	tmpHeadBlock := bc.LatestBlock()
	bc.logger.Infof("after detach process, head %d/%s", tmpHeadBlock.GetHeight(), tmpHeadBlock.GetHash())

	// reverse attach blocks
	for i := len(attachBlocks)/2 - 1; i >= 0; i-- {
		opp := len(attachBlocks) - 1 - i
		attachBlocks[i], attachBlocks[opp] = attachBlocks[opp], attachBlocks[i]
	}

	// attach from low to high
	for _, attachBlock := range attachBlocks {
		err := bc.connectBestBlock(txn, attachBlock)
		if err != nil {
			return err
		}
	}

	newHeadBlock := bc.LatestBlock()
	bc.logger.Infof("after attach process, head %d/%s", newHeadBlock.GetHeight(), newHeadBlock.GetHash())

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
			BlockHash:   block.GetHash(),
			BlockHeight: block.GetHeight(),
			TxIndex:     uint64(txIndex),
		}
		err := ledger.AddPovTxLookup(txPov.Hash, txLookup, txn)
		if err != nil {
			return err
		}
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

		err := ledger.DeletePovTxLookup(txPov.Hash, txn)
		if err != nil {
			return err
		}
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
		if prevBlock.GetHeight() < height {
			return nil
		}

		curBlock = prevBlock
	}
}

func (bc *PovBlockChain) RelativeAncestor(block *types.PovBlock, distance uint64) *types.PovBlock {
	return bc.FindAncestor(block, block.GetHeight()-distance)
}

func (bc *PovBlockChain) CalcNextRequiredTarget(block *types.PovBlock) (types.Signature, error) {
	if (block.GetHeight()+1)%uint64(bc.getConfig().PoV.TargetCycle) != 0 {
		return block.Target, nil
	}

	// nextTarget = prevTarget * (lastBlock.Timestamp - firstBlock.Timestamp) / (blockInterval * targetCycle)

	distance := uint64(bc.getConfig().PoV.TargetCycle - 1)
	firstBlock := bc.RelativeAncestor(block, distance)
	if firstBlock == nil {
		bc.logger.Errorf("failed to get relative ancestor at height %d distance %d", block.GetHeight(), distance)
		return types.ZeroSignature, ErrPovUnknownAncestor
	}

	targetTimeSpan := int64(bc.getConfig().PoV.TargetCycle * bc.getConfig().PoV.BlockInterval)
	minRetargetTimespan := targetTimeSpan / 4
	maxRetargetTimespan := targetTimeSpan * 4

	actualTimespan := block.Timestamp - firstBlock.Timestamp
	if actualTimespan < minRetargetTimespan {
		actualTimespan = minRetargetTimespan
	} else if actualTimespan > maxRetargetTimespan {
		actualTimespan = maxRetargetTimespan
	}

	oldTargetInt := block.Target.ToBigInt()
	nextTargetInt := new(big.Int).Set(oldTargetInt)
	nextTargetInt.Mul(oldTargetInt, big.NewInt(actualTimespan))
	nextTargetInt.Div(nextTargetInt, big.NewInt(targetTimeSpan))

	if nextTargetInt.Cmp(common.PovMinimumTargetInt) < 0 {
		nextTargetInt.SetBytes(common.PovMinimumTargetInt.Bytes())
	}
	if nextTargetInt.Cmp(common.PovMaximumTargetInt) > 0 {
		nextTargetInt.SetBytes(common.PovMaximumTargetInt.Bytes())
	}

	var nextTarget types.Signature
	err := nextTarget.FromBigInt(nextTargetInt)
	if err != nil {
		return types.ZeroSignature, err
	}

	bc.logger.Infof("Difficulty retarget at block height %d", block.GetHeight()+1)
	bc.logger.Infof("Old target %d (%s)", oldTargetInt.BitLen(), oldTargetInt.Text(16))
	bc.logger.Infof("New target %d (%s)", nextTargetInt.BitLen(), nextTargetInt.Text(16))
	bc.logger.Infof("Actual timespan %v, target timespan %v",
		time.Duration(actualTimespan)*time.Second, time.Duration(targetTimeSpan)*time.Second)

	return nextTarget, nil
}

// CalcTotalDifficulty calculates a total difficulty from target. PoV increases
// the difficulty for generating a block by decreasing the value which the
// generated vote signature must be less than. This difficulty target is stored in each
// block header as signature.  The main chain is selected by choosing the chain that has
// the most proof of work (highest difficulty).  Since a lower target difficulty
// value equates to higher actual difficulty, the work value which will be
// accumulated must be the inverse of the difficulty.  Also, in order to avoid
// potential division by zero and really small floating point numbers, the
// result adds 1 to the denominator and multiplies the numerator by 2^512.
func (bc *PovBlockChain) CalcTotalDifficulty(target types.Signature) *big.Int {
	// Return a work value of zero if the passed difficulty bits represent
	// a negative number. Note this should not happen in practice with valid
	// blocks, but an invalid block could trigger it.
	difficultyNum := target.ToBigInt()
	if difficultyNum.Sign() <= 0 {
		return big.NewInt(0)
	}

	// (1 << 512) / (difficultyNum + 1)
	denominator := new(big.Int).Add(difficultyNum, bigOne)
	return new(big.Int).Div(oneLsh512, denominator)
}

func (bc *PovBlockChain) CalcPastMedianTime(prevBlock *types.PovBlock) int64 {
	timestamps := make([]int64, medianTimeBlocks)
	numBlocks := 0
	iterBlock := prevBlock
	for i := 0; i < medianTimeBlocks && iterBlock != nil; i++ {
		timestamps[i] = iterBlock.GetTimestamp()
		numBlocks++

		iterBlock = bc.GetBlockByHash(iterBlock.GetPrevious())
	}

	timestamps = timestamps[:numBlocks]
	sort.Sort(types.TimeSorter(timestamps))

	medianTimestamp := timestamps[numBlocks/2]

	return medianTimestamp
}
