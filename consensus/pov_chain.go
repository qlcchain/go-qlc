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
	ErrPovFailedVerify    = errors.New("failed to verify block")
	ErrPovForkHashZero    = errors.New("fork point hash is zero")
	ErrPovInvalidHead     = errors.New("invalid pov head block")
)

const (
	blockCacheLimit  = 512
	numberCacheLimit = 2048
)

type PovBlockChain struct {
	povEngine *PoVEngine
	logger    *zap.SugaredLogger

	genesisBlock *types.PovBlock
	latestBlock  atomic.Value // Current head of the best block chain

	blockCache   gcache.Cache // Cache for the most recent entire blocks
	heightCache  gcache.Cache
	trieNodePool *trie.NodePool

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

		dbErr = bc.getLedger().AddPovBlock(genesis)
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

func (bc *PovBlockChain) InsertBlock(block *types.PovBlock, stateTrie *trie.Trie) error {
	err := bc.getLedger().BatchUpdate(func(txn db.StoreTxn) error {
		return bc.insertBlock(txn, block, stateTrie)
	})

	if err != nil {
		bc.logger.Errorf("insert block to chain, err %s", err)
	}

	return err
}

func (bc *PovBlockChain) insertBlock(txn db.StoreTxn, block *types.PovBlock, stateTrie *trie.Trie) error {
	currentBlock := bc.LatestBlock()

	// save block to db
	err := bc.getLedger().AddPovBlock(block, txn)
	if err != nil {
		bc.logger.Errorf("add pov block %d/%s failed, err %s", block.Height, block.Hash, err)
		return err
	}

	saveCallback, dbErr := stateTrie.SaveInTxn(txn)
	if dbErr != nil {
		return dbErr
	}
	if saveCallback != nil {
		saveCallback()
	}

	// try to grow best chain
	if block.GetPrevious() == currentBlock.GetHash() {
		return bc.connectBestBlock(txn, block)
	}

	// check fork side chain
	if block.GetHeight() >= currentBlock.GetHeight()+uint64(bc.getConfig().PoV.ForkHeight) {
		return bc.processFork(txn, block)
	} else {
		bc.logger.Debugf("block %d/%s in side chain, prev %s",
			block.GetHeight(), block.GetHash(), block.GetPrevious())
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

	//old: b1 <- b2 <- b3 <- b4 <- b5 <- b6-1 <- b7-1
	//new:                            <- b6-2 <- b7-2 <- b8 <- b9 <- b10

	// step1: attach b7-2(exclude) <- b10
	attachBlock := newBlock
	for height := attachBlock.GetHeight(); height > oldHeadBlock.GetHeight(); height-- {
		attachBlocks = append(attachBlocks, attachBlock)

		prevHash := attachBlock.GetPrevious()
		attachBlock = bc.GetBlockByHash(prevHash)
		if attachBlock == nil {
			bc.logger.Errorf("failed to get previous attach block %s", prevHash)
			return ErrPovInvalidPrevious
		}
	}

	detachBlock := oldHeadBlock

	// step2: attach b6-2 <- b7-2(include)
	//        detach b6-1 <- b7-1(include)
	//        fork point: b5
	for {
		attachBlocks = append(attachBlocks, attachBlock)
		detachBlocks = append(detachBlocks, detachBlock)

		if attachBlock.GetPrevious() == detachBlock.GetPrevious() {
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

	if forkHash.IsZero() {
		bc.logger.Errorf("fork block hash is zero")
		return ErrPovForkHashZero
	}

	forkBlock := bc.GetBlockByHash(forkHash)
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

	actualTimespan := block.Timestamp - firstBlock.Timestamp
	if actualTimespan <= 0 {
		actualTimespan = 1
	}
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
