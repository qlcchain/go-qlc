package pov

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/qlcchain/go-qlc/common/event"

	"github.com/bluele/gcache"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
)

const (
	blockCacheLimit  = 1024
	headerCacheLimit = 4096
	medianTimeBlocks = 11
)

type ChainState uint

const (
	ChainStateNone ChainState = iota
	ChainStateMain
	ChainStateSide
)

var chainStateInfo = [...]string{
	ChainStateNone: "none",
	ChainStateMain: "main",
	ChainStateSide: "side",
}

func (s ChainState) String() string {
	if s > ChainStateSide {
		return "unknown"
	}
	return chainStateInfo[s]
}

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
	config *config.Config
	ledger ledger.Store
	logger *zap.SugaredLogger
	em     *eventManager
	eb     event.EventBus

	genesisBlock *types.PovBlock
	latestBlock  atomic.Value // Current head of the best block chain
	latestHeader atomic.Value

	hashBlockCache    gcache.Cache // hash => block
	heightBlockCache  gcache.Cache // height => best block
	hashTdCache       gcache.Cache // hash => td
	hashHeaderCache   gcache.Cache // hash => header
	heightHeaderCache gcache.Cache // height => best header

	trieCache    gcache.Cache // stateHash => trie
	trieNodePool *trie.NodePool

	doingMinerStat *atomic.Bool

	sideProcNum int
	forkProcNum int

	quitCh chan struct{}
	wg     sync.WaitGroup
}

func NewPovBlockChain(cfg *config.Config, eb event.EventBus, l ledger.Store) *PovBlockChain {
	chain := &PovBlockChain{
		config:         cfg,
		eb:             eb,
		ledger:         l,
		logger:         log.NewLogger("pov_chain"),
		doingMinerStat: atomic.NewBool(false),
	}
	chain.em = newEventManager()

	chain.hashBlockCache = gcache.New(blockCacheLimit).LRU().Build()
	chain.heightBlockCache = gcache.New(blockCacheLimit).LRU().Build()
	chain.hashTdCache = gcache.New(blockCacheLimit).LRU().Build()

	chain.hashHeaderCache = gcache.New(headerCacheLimit).LRU().Build()
	chain.heightHeaderCache = gcache.New(headerCacheLimit).LRU().Build()

	chain.trieCache = gcache.New(128).LRU().Build()

	chain.quitCh = make(chan struct{})

	return chain
}

func (bc *PovBlockChain) getLedger() ledger.Store {
	return bc.ledger
}

func (bc *PovBlockChain) Init() error {
	var err error

	bc.trieNodePool = trie.NewSimpleTrieNodePool()

	needReset := false
	genesisBlock, _ := bc.getLedger().GetPovBlockByHeight(0)
	if genesisBlock == nil {
		needReset = true
	} else if !bc.IsGenesisBlock(genesisBlock) {
		needReset = true
	}

	if needReset {
		bc.logger.Warn("pov genesis block incorrect, resetting chain")
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
	common.Go(bc.statLoop)
	return nil
}

func (bc *PovBlockChain) Stop() error {
	close(bc.quitCh)
	bc.wg.Wait()
	bc.trieNodePool.Clear()
	return nil
}

func (bc *PovBlockChain) statLoop() {
	checkTicker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-bc.quitCh:
			return

		case <-checkTicker.C:
			bc.onMinerDayStatTimer()
		}
	}
}

func (bc *PovBlockChain) onMinerDayStatTimer() {
	if !bc.doingMinerStat.CAS(false, true) {
		return
	}
	defer bc.doingMinerStat.CAS(true, false)

	latestBlock := bc.LatestBlock()

	curDayIndex := uint32(0)
	latestDayStat, err := bc.getLedger().GetLatestPovMinerStat()
	if err != nil {
		bc.logger.Errorf("failed to get latest pov miner stat, err %s", err)
		return
	}
	if latestDayStat != nil {
		curDayIndex = latestDayStat.DayIndex + 1
	}

	for {
		dayStartHeight := uint64(curDayIndex) * uint64(common.POVChainBlocksPerDay)
		dayEndHeight := dayStartHeight + uint64(common.POVChainBlocksPerDay) - 1

		if dayEndHeight+(common.PovMinerRewardHeightGapToLatest) > latestBlock.GetHeight() {
			break
		}

		bc.logger.Infof("curDayIndex: %d, latestBlock: %d", curDayIndex, latestBlock.GetHeight())
		dayStat := types.NewPovMinerDayStat()
		dayStat.DayIndex = curDayIndex

		for height := dayStartHeight; height <= dayEndHeight; height++ {
			header, err := bc.getLedger().GetPovHeaderByHeight(height)
			if err != nil {
				bc.logger.Warnf("failed to get pov header %d, err %s", height, err)
				return
			}

			//calc repReward
			if height%common.DPosOnlinePeriod == common.DPosOnlinePeriod-1 {
				repStates := bc.GetAllOnlineRepStates(header)
				bc.logger.Debugf("get online rep states %d at block height %d", len(repStates), height)

				// calc total weight of all reps
				repWeightTotal := types.NewBalance(0)
				for _, rep := range repStates {
					repWeightTotal = repWeightTotal.Add(rep.CalcTotal())
				}
				repWeightTotalBig := big.NewInt(repWeightTotal.Int64())

				// calc total reward of all blocks in period
				var i uint64
				for i = 0; i < common.DPosOnlinePeriod; i++ {
					povHead, err := bc.getLedger().GetPovHeaderByHeight(height - i)
					if err != nil {
						bc.logger.Warnf("failed to get pov header %d, err %s", height-i, err)
						return
					}
					repRewardAllBig := povHead.GetRepReward()

					// divide reward to each rep
					for _, rep := range repStates {
						// calc rule is : repReward = totalReward / repWeight * totalWeight
						repRewardBig := big.NewInt(rep.CalcTotal().Int64())
						repRewardBig = repRewardBig.Mul(repRewardBig, repRewardAllBig.Int)
						amountBig := repRewardBig.Div(repRewardBig, repWeightTotalBig)
						amount := types.NewBalance(amountBig.Int64())

						repAddrStr := rep.Account.String()
						minerStat := dayStat.MinerStats[repAddrStr]
						if minerStat == nil {
							minerStat = types.NewPovMinerStatItem()
							dayStat.MinerStats[repAddrStr] = minerStat
						}
						minerStat.RepBlockNum++
						minerStat.RepReward = minerStat.RepReward.Add(amount)
					}
				}
			}

			cbAddrStr := header.GetMinerAddr().String()
			minerStat := dayStat.MinerStats[cbAddrStr]
			if minerStat == nil {
				minerStat = types.NewPovMinerStatItem()
				dayStat.MinerStats[cbAddrStr] = minerStat
			}

			if minerStat.IsMiner {
				minerStat.LastHeight = header.GetHeight()
			} else {
				minerStat.FirstHeight = header.GetHeight()
				minerStat.IsMiner = true
			}

			minerStat.BlockNum++
			minerStat.RewardAmount = minerStat.RewardAmount.Add(header.GetMinerReward())
		}

		minerNum := uint32(0)
		for _, stat := range dayStat.MinerStats {
			if stat.IsMiner {
				minerNum++
			}
		}
		dayStat.MinerNum = minerNum

		err = bc.getLedger().AddPovMinerStat(dayStat)
		if err != nil {
			bc.logger.Errorf("failed to add pov miner stat, day %d, err %s", dayStat.DayIndex, err)
			return
		}

		bc.logger.Infof("add pov miner stat, %d-%d-%d", curDayIndex, dayStartHeight, dayEndHeight)
		curDayIndex++
	}
}

func (bc *PovBlockChain) loadLastState() error {
	// Make sure the entire head block is available
	latestBlock, err := bc.getLedger().GetLatestPovBlock()
	if err != nil {
		bc.logger.Errorf("failed to get latest block, err %s", err)
	}
	if latestBlock == nil {
		// Corrupt or empty database, init from scratch
		bc.logger.Warn("head block missing, resetting chain")
		return bc.ResetChainState()
	}

	// Everything seems to be fine, set as the head block
	bc.StoreLatestBlock(latestBlock)

	bc.logger.Infof("loaded latest block %d/%s", latestBlock.GetHeight(), latestBlock.GetHash())

	latestStateHash := latestBlock.GetStateHash()
	bc.logger.Infof("loaded latest state hash %s", latestStateHash)
	latestTrie := bc.GetStateTrie(&latestStateHash)
	if latestTrie == nil || latestTrie.Root == nil {
		panic(fmt.Errorf("invalid latest state hash %s", latestStateHash))
	}

	return nil
}

func (bc *PovBlockChain) ResetChainState() error {
	var err error

	// clear all pov blocks
	allBlkCnt, _ := bc.getLedger().CountPovBlocks()
	bc.logger.Infof("drop all pov blocks %d", allBlkCnt)
	err = bc.getLedger().DropAllPovBlocks()
	if err != nil {
		return err
	}

	_ = bc.getLedger().DBStore().Purge()

	// init with genesis block
	genesisBlock := common.GenesisPovBlock()

	return bc.resetWithGenesisBlock(&genesisBlock)
}

func (bc *PovBlockChain) resetWithGenesisBlock(genesis *types.PovBlock) error {
	bc.logger.Infof("reset with genesis block %d/%s", genesis.GetHeight(), genesis.GetHash())

	var saveCallback func()

	err := bc.getLedger().BatchUpdate(func(txn db.StoreTxn) error {
		var dbErr error

		td := bc.CalcTotalDifficulty(types.NewPovTD(), &genesis.Header)
		dbErr = bc.getLedger().AddPovBlock(genesis, td, txn)
		if dbErr != nil {
			return dbErr
		}
		dbErr = bc.getLedger().AddPovBestHash(genesis.GetHeight(), genesis.GetHash(), txn)
		if dbErr != nil {
			return dbErr
		}

		for txIdx, txPov := range genesis.GetAllTxs() {
			txl := new(types.PovTxLookup)
			txl.BlockHash = genesis.GetHash()
			txl.BlockHeight = genesis.GetHeight()
			txl.TxIndex = uint64(txIdx)
			dbErr = bc.getLedger().AddPovTxLookup(txPov.GetHash(), txl, txn)
			if dbErr != nil {
				return dbErr
			}
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

		if saveCallback != nil {
			saveCallback()
		}

		return nil
	})

	if err != nil {
		bc.logger.Fatalf("failed to reset with genesis block")
		return err
	}

	bc.genesisBlock = genesis
	bc.StoreLatestBlock(genesis)

	_ = bc.getLedger().DBStore().Purge()

	return nil
}

func (bc *PovBlockChain) GenesisBlock() *types.PovBlock {
	return bc.genesisBlock
}

func (bc *PovBlockChain) LatestBlock() *types.PovBlock {
	return bc.latestBlock.Load().(*types.PovBlock)
}

func (bc *PovBlockChain) StoreLatestBlock(block *types.PovBlock) {
	header := block.GetHeader()

	bc.latestBlock.Store(block)
	bc.latestHeader.Store(header)

	// set best block in cache
	_ = bc.heightBlockCache.Set(block.GetHeight(), block)
	_ = bc.heightHeaderCache.Set(block.GetHeight(), header)
}

func (bc *PovBlockChain) IsGenesisBlock(block *types.PovBlock) bool {
	return common.IsGenesisPovBlock(block)
}

func (bc *PovBlockChain) GetBlockByHeight(height uint64) (*types.PovBlock, error) {
	v, _ := bc.heightBlockCache.Get(height)
	if v != nil {
		return v.(*types.PovBlock), nil
	}

	block, err := bc.getLedger().GetPovBlockByHeight(height)
	if block != nil {
		_ = bc.heightBlockCache.Set(height, block)
		_ = bc.hashBlockCache.Set(block.GetHash(), block)
	}
	return block, err
}

func (bc *PovBlockChain) GetBlockByHash(hash types.Hash) *types.PovBlock {
	if hash.IsZero() {
		return nil
	}

	v, _ := bc.hashBlockCache.Get(hash)
	if v != nil {
		return v.(*types.PovBlock)
	}

	block, _ := bc.getLedger().GetPovBlockByHash(hash)
	if block != nil {
		_ = bc.hashBlockCache.Set(hash, block)
		return block
	}

	return nil
}

func (bc *PovBlockChain) HasFullBlock(hash types.Hash, height uint64) bool {
	if bc.hashBlockCache.Has(hash) {
		return true
	}

	if bc.getLedger().HasPovBlock(height, hash) {
		return true
	}

	return false
}

func (bc *PovBlockChain) LatestHeader() *types.PovHeader {
	return bc.latestHeader.Load().(*types.PovHeader)
}

func (bc *PovBlockChain) GetBestBlockByHash(hash types.Hash) *types.PovBlock {
	block, _ := bc.getLedger().GetPovBlockByHash(hash)
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
	if !bc.HasFullBlock(hash, height) {
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

func (bc *PovBlockChain) GetBlockTDByHash(hash types.Hash) *types.PovTD {
	hdr := bc.GetHeaderByHash(hash)
	if hdr == nil {
		return nil
	}

	return bc.GetBlockTDByHashAndHeight(hdr.GetHash(), hdr.GetHeight())
}

func (bc *PovBlockChain) GetBlockTDByHashAndHeight(hash types.Hash, height uint64) *types.PovTD {
	v, _ := bc.hashTdCache.Get(hash)
	if v != nil {
		return v.(*types.PovTD)
	}

	td, err := bc.getLedger().GetPovTD(hash, height)
	if err != nil {
		return nil
	}

	_ = bc.hashTdCache.Set(hash, td)
	return td
}

func (bc *PovBlockChain) GetHeaderByHash(hash types.Hash) *types.PovHeader {
	if hash.IsZero() {
		return nil
	}

	v, _ := bc.hashHeaderCache.Get(hash)
	if v != nil {
		return v.(*types.PovHeader)
	}

	header, _ := bc.getLedger().GetPovHeaderByHash(hash)
	if header != nil {
		_ = bc.hashHeaderCache.Set(hash, header)
		return header
	}

	return nil
}

func (bc *PovBlockChain) GetHeaderByHeight(height uint64) *types.PovHeader {
	v, _ := bc.heightHeaderCache.Get(height)
	if v != nil {
		return v.(*types.PovHeader)
	}

	header, _ := bc.getLedger().GetPovHeaderByHeight(height)
	if header != nil {
		_ = bc.heightHeaderCache.Set(height, header)
		_ = bc.hashHeaderCache.Set(header.GetHash(), header)
		return header
	}

	return nil
}

func (bc *PovBlockChain) HasHeader(hash types.Hash, height uint64) bool {
	if bc.hashHeaderCache.Has(hash) {
		return true
	}

	if bc.getLedger().HasPovHeader(height, hash) {
		return true
	}

	return false
}

func (bc *PovBlockChain) GetBlockLocator(hash types.Hash) []*types.Hash {
	var header *types.PovHeader

	if hash.IsZero() {
		header = bc.LatestHeader()
		hash = header.GetHash()
	} else {
		header = bc.GetHeaderByHash(hash)
	}
	if header == nil {
		return nil
	}

	// Calculate the max number of entries that will ultimately be in the
	// block locator.  See the description of the algorithm for how these
	// numbers are derived.
	var maxEntries uint8
	if header.GetHeight() <= 12 {
		maxEntries = uint8(header.GetHeight()) + 1
	} else {
		// Requested hash itself + previous 10 entries + genesis block.
		// Then floor(log2(height-10)) entries for the skip portion.
		adjustedHeight := uint32(header.GetHeight()) - 10
		maxEntries = 12 + fastLog2Floor(adjustedHeight)
	}
	locator := make([]*types.Hash, 0, maxEntries)

	step := uint64(1)
	for header != nil {
		locHash := header.GetHash()
		locator = append(locator, &locHash)

		// Nothing more to add once the genesis block has been added.
		if header.GetHeight() == 0 {
			break
		}

		// Calculate height of previous node to include ensuring the
		// final node is the genesis block.
		height := uint64(0)
		if header.GetHeight() > step {
			height = header.GetHeight() - step
		}

		header = bc.FindAncestor(header, height)

		// Once 11 entries have been included, start doubling the
		// distance between included hashes.
		if len(locator) > 10 {
			step *= 2
		}
	}

	return locator
}

func (bc *PovBlockChain) LocateBestBlock(locator []*types.Hash) *types.PovBlock {
	startBlock := bc.GenesisBlock()

	for _, locHash := range locator {
		if locHash == nil {
			continue
		}
		block := bc.GetBestBlockByHash(*locHash)
		if block != nil {
			startBlock = block
			break
		}
	}

	return startBlock
}

func (bc *PovBlockChain) InsertBlock(block *types.PovBlock, stateTrie *trie.Trie) error {
	chainState := ChainStateNone

	err := bc.getLedger().BatchUpdate(func(txn db.StoreTxn) error {
		var dbErr error
		chainState, dbErr = bc.insertBlock(txn, block, stateTrie)
		return dbErr
	})

	if err != nil {
		bc.logger.Errorf("failed to insert block %d/%s to chain", block.GetHeight(), block.GetHash())
	} else {
		bc.logger.Infof("success to insert block %d/%s to %s chain", block.GetHeight(), block.GetHash(), chainState)
	}

	if (block.GetHeight()+1)%uint64(common.POVChainBlocksPerDay) == 0 {
		bc.onMinerDayStatTimer()
	}

	return err
}

func (bc *PovBlockChain) insertBlock(txn db.StoreTxn, block *types.PovBlock, stateTrie *trie.Trie) (ChainState, error) {
	currentBlock := bc.LatestBlock()

	bestTD, err := bc.getLedger().GetPovTD(currentBlock.GetHash(), currentBlock.GetHeight(), txn)
	if err != nil {
		bc.logger.Errorf("get pov best td %d/%s failed, err %s", currentBlock.GetHeight(), currentBlock.GetHash(), err)
		return ChainStateNone, err
	}

	prevTD, err := bc.getLedger().GetPovTD(block.GetPrevious(), block.GetHeight()-1, txn)
	if err != nil {
		bc.logger.Errorf("get pov previous td %d/%s failed, err %s", block.GetHeight()-1, block.GetPrevious(), err)
		return ChainStateNone, err
	}

	blockTD := bc.CalcTotalDifficulty(prevTD, block.GetHeader())

	// save block to db
	if !bc.getLedger().HasPovBlock(block.GetHeight(), block.GetHash(), txn) {
		err = bc.getLedger().AddPovBlock(block, blockTD, txn)
		if err != nil && err != ledger.ErrBlockExists {
			bc.logger.Errorf("add pov block %d/%s failed, err %s", block.GetHeight(), block.GetHash(), err)
			return ChainStateNone, err
		}
	}

	saveCallback, dbErr := stateTrie.SaveInTxn(txn)
	if dbErr != nil {
		return ChainStateNone, dbErr
	}
	if saveCallback != nil {
		saveCallback()
	}

	_ = bc.hashTdCache.Set(block.GetHash(), blockTD)

	tdCmpRet := blockTD.Chain.Cmp(&bestTD.Chain)
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
			err := bc.connectBestBlock(txn, block)
			return ChainStateMain, err
		}

		bc.forkProcNum++
		bc.logger.Infof("block %d/%s td %d/%s, need to doing fork, prev %s",
			block.GetHeight(), block.GetHash(), blockTD.Chain.BitLen(), blockTD.Chain.Text(16), block.GetPrevious())
		err := bc.processFork(txn, block)
		return ChainStateMain, err
	}

	bc.sideProcNum++
	bc.logger.Debugf("block %d/%s td %d/%s in side chain, prev %s",
		block.GetHeight(), block.GetHash(), blockTD.Chain.BitLen(), blockTD.Chain.Text(16), block.GetPrevious())

	return ChainStateSide, nil
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

	bc.StoreLatestBlock(block)

	bc.eb.Publish(common.EventPovConnectBestBlock, block)

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

	// remove old best block in cache
	bc.heightBlockCache.Remove(block.GetHeight())
	bc.heightHeaderCache.Remove(block.GetHeight())

	bc.StoreLatestBlock(prevBlock)

	bc.eb.Publish(common.EventPovDisconnectBestBlock, block)

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

	bc.logger.Infof("find fork block %d/%s, detach blocks %d, attach blocks %d",
		forkBlock.GetHeight(), forkBlock.GetHash(), len(detachBlocks), len(attachBlocks))

	if len(detachBlocks) > int(common.PoVMaxForkHeight) {
		bc.logger.Errorf("fork detach blocks %d exceed max limit %d", len(detachBlocks), common.PoVMaxForkHeight)
		return ErrPovInvalidFork
	}

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
		bc.logger.Errorf("add pov best hash %d/%s failed, err %s", block.GetHeight(), block.GetHash(), err)
		return err
	}
	return nil
}

func (bc *PovBlockChain) disconnectBlock(txn db.StoreTxn, block *types.PovBlock) error {
	err := bc.getLedger().DeletePovBestHash(block.GetHeight(), txn)
	if err != nil {
		bc.logger.Errorf("delete pov best hash %d/%s failed, err %s", block.GetHeight(), block.GetHash(), err)
		return err
	}
	return nil
}

func (bc *PovBlockChain) connectTransactions(txn db.StoreTxn, block *types.PovBlock) error {
	allTxs := block.GetAllTxs()
	for txIndex, txPov := range allTxs {
		// txIndex == 0 is CoinBaseTX
		if txIndex > 0 && txPov.Block == nil {
			txPov.Block, _ = bc.getLedger().GetStateBlock(txPov.Hash, txn)
			if txPov.Block == nil {
				bc.logger.Errorf("connect txs failed to get state block %s", txPov.Hash)
			}
		}
		txLookup := &types.PovTxLookup{
			BlockHash:   block.GetHash(),
			BlockHeight: block.GetHeight(),
			TxIndex:     uint64(txIndex),
		}
		err := bc.getLedger().AddPovTxLookup(txPov.Hash, txLookup, txn)
		if err != nil {
			return err
		}
	}

	bc.em.TriggerBlockEvent(EventConnectPovBlock, block)

	return nil
}

func (bc *PovBlockChain) disconnectTransactions(txn db.StoreTxn, block *types.PovBlock) error {
	allTxs := block.GetAllTxs()
	for txIndex, txPov := range allTxs {
		// txIndex == 0 is CoinBaseTX
		if txIndex > 0 && txPov.Block == nil {
			txPov.Block, _ = bc.getLedger().GetStateBlock(txPov.Hash, txn)
			if txPov.Block == nil {
				bc.logger.Errorf("disconnect txs failed to get state block %s", txPov.Hash)
			}
		}

		err := bc.getLedger().DeletePovTxLookup(txPov.Hash, txn)
		if err != nil {
			return err
		}
	}

	bc.em.TriggerBlockEvent(EventDisconnectPovBlock, block)

	return nil
}

func (bc *PovBlockChain) FindAncestor(header *types.PovHeader, height uint64) *types.PovHeader {
	if height > header.GetHeight() {
		return nil
	}

	curHeader := header
	for {
		prevHeader := bc.GetHeaderByHash(curHeader.GetPrevious())
		if prevHeader == nil {
			return nil
		}

		if prevHeader.GetHeight() == height {
			return prevHeader
		}
		if prevHeader.GetHeight() < height {
			return nil
		}

		curHeader = prevHeader
	}
}

func (bc *PovBlockChain) RelativeAncestor(header *types.PovHeader, distance uint64) *types.PovHeader {
	return bc.FindAncestor(header, header.GetHeight()-distance)
}

func (bc *PovBlockChain) CalcTotalDifficulty(prevTD *types.PovTD, header *types.PovHeader) *types.PovTD {
	curTD := prevTD.Copy()

	curWorkAlgo := types.CalcWorkIntToBigNum(header.GetAlgoTargetInt())

	curTD.Chain.Add(&prevTD.Chain, curWorkAlgo)

	algoType := header.GetAlgoType()
	switch algoType {
	case types.ALGO_SHA256D:
		curTD.Sha256d.Add(&prevTD.Sha256d, curWorkAlgo)
	case types.ALGO_SCRYPT:
		curTD.Scrypt.Add(&prevTD.Scrypt, curWorkAlgo)
	case types.ALGO_X11:
		curTD.X11.Add(&prevTD.X11, curWorkAlgo)
	}

	return curTD
}

func (bc *PovBlockChain) CalcPastMedianTime(prevHeader *types.PovHeader) uint32 {
	timestamps := make([]uint32, medianTimeBlocks)
	numHeaders := 0
	iterHeader := prevHeader
	for i := 0; i < medianTimeBlocks && iterHeader != nil; i++ {
		timestamps[i] = iterHeader.GetTimestamp()
		numHeaders++

		iterHeader = bc.GetHeaderByHash(iterHeader.GetPrevious())
	}

	timestamps = timestamps[:numHeaders]
	sort.Sort(types.TimeSorter(timestamps))

	medianTimestamp := timestamps[numHeaders/2]

	return medianTimestamp
}

func (bc *PovBlockChain) CalcBlockReward(header *types.PovHeader) (types.Balance, types.Balance) {
	return bc.CalcBlockRewardByQLC(header)
}

func (bc *PovBlockChain) CalcBlockRewardByQLC(header *types.PovHeader) (types.Balance, types.Balance) {
	miner1 := new(big.Int).Mul(common.PovMinerRewardPerBlockInt, big.NewInt(int64(common.PovMinerRewardRatioMiner)))
	miner2 := new(big.Int).Div(miner1, big.NewInt(100))

	rep1 := new(big.Int).Mul(common.PovMinerRewardPerBlockInt, big.NewInt(int64(common.PovMinerRewardRatioRep)))
	rep2 := new(big.Int).Div(rep1, big.NewInt(100))

	return types.NewBalanceFromBigInt(miner2), types.NewBalanceFromBigInt(rep2)
}

func (bc *PovBlockChain) GetDebugInfo() map[string]interface{} {
	// !!! be very careful about to map concurrent read !!!

	info := make(map[string]interface{})
	info["hashBlockCache"] = bc.hashBlockCache.Len(false)
	info["hashHeaderCache"] = bc.hashHeaderCache.Len(false)
	info["heightBlockCache"] = bc.heightBlockCache.Len(false)
	info["heightHeaderCache"] = bc.heightHeaderCache.Len(false)
	info["hashTdCache"] = bc.hashTdCache.Len(false)
	info["trieCache"] = bc.trieCache.Len(false)
	info["trieNodePool"] = bc.trieNodePool.Len()

	info["sideProcNum"] = bc.sideProcNum
	info["forkProcNum"] = bc.forkProcNum

	latestHdr := bc.LatestHeader()
	if latestHdr != nil {
		info["latestHeight"] = latestHdr.GetHeight()
	}

	return info
}
