package pov

import (
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/trie"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"
)

type PovProcessorChainReader interface {
	HasBestBlock(hash types.Hash, height uint64) bool
	GetBlockByHash(hash types.Hash) *types.PovBlock
	InsertBlock(block *types.PovBlock, stateTrie *trie.Trie) error
}

type PovProcessorVerifier interface {
	VerifyFull(block *types.PovBlock) *PovVerifyStat
}

type PovProcessorSyncer interface {
	requestBlocksByHashes(reqBlkHashes []*types.Hash, peerID string)
	requestTxsByHashes(reqTxHashes []*types.Hash, peerID string)
}

const (
	blockChanSize = 8192

	checkTxPendingTickerSec   = 5
	checkOrphanBlockTickerSec = 5
	checkReqOrphanTickerSec   = 1

	minPullTxIntervalSec    = 10
	minPullBlockIntervalSec = 15
)

type PovBlockResult struct {
	err error
}

type PovBlockSource struct {
	block   *types.PovBlock
	from    types.PovBlockFrom
	peerID  string
	replyCh chan PovBlockResult
}

type PovOrphanBlock struct {
	blockSrc     *PovBlockSource
	addTime      time.Time
	lastPullTime time.Time
}

type PovPendingBlock struct {
	addTime   time.Time
	blockSrc  *PovBlockSource
	txResults map[types.Hash]process.ProcessResult
}

type PovTxPendingEntry struct {
	pendingBlocks []*PovPendingBlock
	lastPullTime  time.Time
}

type PovBlockProcessor struct {
	eb       event.EventBus
	ledger   ledger.Store
	chain    PovProcessorChainReader
	verifier PovProcessorVerifier
	syncer   PovProcessorSyncer
	logger   *zap.SugaredLogger

	orphanBlocks  map[types.Hash]*PovOrphanBlock   // blockHash -> block
	parentOrphans map[types.Hash][]*PovOrphanBlock // blockHash -> child blocks

	pendingBlocks   map[types.Hash]*PovPendingBlock   // blockHash -> block
	txPendingBlocks map[types.Hash]*PovTxPendingEntry // txHash -> block list, multi side chains
	txPendingMux    sync.Mutex

	deOrphanBlocks  []*PovBlockSource
	reqOrphanBlocks map[types.Hash]*PovOrphanBlock

	blockInCh  map[types.Hash]struct{}
	blockChMux sync.Mutex

	blockHighCh chan *PovBlockSource
	blockNormCh chan *PovBlockSource
	quitCh      chan struct{}

	syncState common.SyncState
	syncMux   sync.RWMutex
}

func NewPovBlockProcessor(eb event.EventBus, ledger ledger.Store,
	chain PovProcessorChainReader,
	verifier PovProcessorVerifier,
	syncer PovProcessorSyncer) *PovBlockProcessor {
	bp := &PovBlockProcessor{
		eb:       eb,
		ledger:   ledger,
		logger:   log.NewLogger("pov_processor"),
		chain:    chain,
		verifier: verifier,
		syncer:   syncer,
	}

	bp.orphanBlocks = make(map[types.Hash]*PovOrphanBlock)
	bp.parentOrphans = make(map[types.Hash][]*PovOrphanBlock)
	bp.pendingBlocks = make(map[types.Hash]*PovPendingBlock)
	bp.txPendingBlocks = make(map[types.Hash]*PovTxPendingEntry)

	bp.deOrphanBlocks = make([]*PovBlockSource, 0)
	bp.reqOrphanBlocks = make(map[types.Hash]*PovOrphanBlock)

	bp.blockInCh = make(map[types.Hash]struct{})

	bp.blockHighCh = make(chan *PovBlockSource, blockChanSize/10)
	bp.blockNormCh = make(chan *PovBlockSource, blockChanSize)
	bp.quitCh = make(chan struct{})

	bp.syncState = common.SyncNotStart

	return bp
}

func (bp *PovBlockProcessor) Start() error {
	if bp.eb != nil {
		bp.eb.Subscribe(common.EventAddRelation, bp.onAddStateBlock)

		bp.eb.Subscribe(common.EventPovSyncState, bp.onPovSyncState)
	}

	common.Go(bp.loop)
	common.Go(bp.loopCheckTxPending)
	return nil
}

func (bp *PovBlockProcessor) Init() error {
	return nil
}

func (bp *PovBlockProcessor) Stop() error {
	if bp.eb != nil {
		bp.eb.Unsubscribe(common.EventAddRelation, bp.onAddStateBlock)

		bp.eb.Unsubscribe(common.EventPovSyncState, bp.onPovSyncState)
	}

	close(bp.quitCh)
	return nil
}

func (bp *PovBlockProcessor) onAddStateBlock(tx *types.StateBlock) {
	bp.txPendingMux.Lock()
	defer bp.txPendingMux.Unlock()

	txHash := tx.GetHash()

	txPendEntry, ok := bp.txPendingBlocks[txHash]
	if !ok {
		return
	}

	bp.logger.Debugf("recv add state block %s", txHash)

	delete(bp.txPendingBlocks, txHash)

	for _, pendingBlock := range txPendEntry.pendingBlocks {
		if _, ok := pendingBlock.txResults[txHash]; !ok {
			continue
		}
		delete(pendingBlock.txResults, txHash)

		if len(pendingBlock.txResults) <= 0 {
			bp.releaseTxPendingBlock(pendingBlock)
		}
	}
}

func (bp *PovBlockProcessor) onPovSyncState(state common.SyncState) {
	bp.syncMux.Lock()
	defer bp.syncMux.Unlock()

	if state.IsSyncExited() == false {
		return
	}

	bp.syncState = state
}

func (bp *PovBlockProcessor) checkAndSetBlockInChan(block *types.PovBlock) bool {
	bp.blockChMux.Lock()
	defer bp.blockChMux.Unlock()
	if _, ok := bp.blockInCh[block.GetHash()]; ok {
		return true
	}
	bp.blockInCh[block.GetHash()] = struct{}{}

	return false
}

func (bp *PovBlockProcessor) AddBlock(block *types.PovBlock, from types.PovBlockFrom, peerID string) error {
	if bp.checkAndSetBlockInChan(block) {
		return nil
	}

	blockSrc := &PovBlockSource{block: block, from: from, peerID: peerID}
	bp.blockNormCh <- blockSrc
	return nil
}

func (bp *PovBlockProcessor) AddMinedBlock(block *types.PovBlock) error {
	replyCh := make(chan PovBlockResult)
	bp.blockNormCh <- &PovBlockSource{block: block, from: types.PovBlockFromLocal, replyCh: replyCh}
	result := <-replyCh
	close(replyCh)
	return result.err
}

func (bp *PovBlockProcessor) loop() {
	checkOrphanTicker := time.NewTicker(checkOrphanBlockTickerSec * time.Second)
	defer checkOrphanTicker.Stop()

	requestOrphanTicker := time.NewTicker(checkReqOrphanTickerSec * time.Second)
	defer requestOrphanTicker.Stop()

	for {
		select {
		case <-bp.quitCh:
			bp.logger.Info("Exiting process blocks loop")
			return

		case blockSrc := <-bp.blockHighCh:
			bp.processOneBlock(blockSrc)

		case blockSrc := <-bp.blockNormCh:
			bp.processOneBlock(blockSrc)

		case <-checkOrphanTicker.C:
			bp.onCheckOrphanBlocksTimer()

		case <-requestOrphanTicker.C:
			bp.onRequestOrphanBlocksTimer()
		}
	}
}

func (bp *PovBlockProcessor) loopCheckTxPending() {
	txPendingTicker := time.NewTicker(checkTxPendingTickerSec * time.Second)

	for {
		select {
		case <-bp.quitCh:
			return
		case <-txPendingTicker.C:
			bp.onCheckTxPendingBlocksTimer()
		}
	}
}

func (bp *PovBlockProcessor) processOneBlock(blockSrc *PovBlockSource) {
	bp.blockChMux.Lock()
	delete(bp.blockInCh, blockSrc.block.GetHash())
	bp.blockChMux.Unlock()

	err := bp.processBlock(blockSrc)
	if blockSrc.replyCh != nil {
		blockSrc.replyCh <- PovBlockResult{err: err}
	}

	for len(bp.deOrphanBlocks) > 0 {
		bp.logger.Infof("processing orphan blocks %d", len(bp.deOrphanBlocks))
		needProcBlocks := make([]*PovBlockSource, len(bp.deOrphanBlocks))
		copy(needProcBlocks, bp.deOrphanBlocks)
		bp.deOrphanBlocks = make([]*PovBlockSource, 0)

		for _, blockSrc := range needProcBlocks {
			err := bp.processBlock(blockSrc)
			if blockSrc.replyCh != nil {
				blockSrc.replyCh <- PovBlockResult{err: err}
			}
		}
	}
}

func (bp *PovBlockProcessor) processBlock(blockSrc *PovBlockSource) error {
	block := blockSrc.block
	blockHash := blockSrc.block.GetHash()
	bp.logger.Debugf("process block, %d/%s", blockSrc.block.GetHeight(), blockHash)

	// check duplicate block
	if bp.HasOrphanBlock(blockHash) {
		bp.logger.Debugf("duplicate block %s exist in orphan", blockHash)
		return nil
	}
	if bp.HasPendingBlock(blockHash) {
		bp.logger.Debugf("duplicate block %s exist in pending", blockHash)
		return nil
	}
	if bp.chain.HasBestBlock(blockHash, block.GetHeight()) {
		bp.logger.Debugf("duplicate block %s exist in best chain", blockHash)
		return nil
	}

	prevBlock := bp.chain.GetBlockByHash(block.GetPrevious())
	if prevBlock == nil {
		bp.addOrphanBlock(blockSrc)
		return nil
	}

	// check block
	stat := bp.verifier.VerifyFull(block)
	if stat == nil {
		bp.logger.Errorf("failed to verify block %s", block.GetHash())
		return ErrPovFailedVerify
	}

	// orphan block
	if stat.Result == process.GapPrevious {
		bp.addOrphanBlock(blockSrc)
		return nil
	} else if stat.Result == process.GapTransaction {
		bp.addTxPendingBlock(blockSrc, stat)
		return nil
	} else if stat.Result != process.Progress {
		bp.logger.Errorf("failed to verify block %s, result %s, err %s", block.GetHash(), stat.Result, stat.ErrMsg)
		return ErrPovFailedVerify
	}

	err := bp.chain.InsertBlock(block, stat.StateTrie)

	if err == nil {
		_ = bp.enqueueOrphanBlocks(blockSrc)
	}

	return err
}

func (bp *PovBlockProcessor) addOrphanBlock(blockSrc *PovBlockSource) {
	blockHash := blockSrc.block.GetHash()

	nowTime := time.Now()

	for _, oBlock := range bp.orphanBlocks {
		if nowTime.After(oBlock.addTime.Add(time.Hour)) {
			bp.removeOrphanBlock(oBlock)
			continue
		}
	}

	oBlock := &PovOrphanBlock{
		blockSrc:     blockSrc,
		addTime:      nowTime,
		lastPullTime: nowTime,
	}
	bp.orphanBlocks[blockHash] = oBlock

	prevHash := blockSrc.block.GetPrevious()
	bp.parentOrphans[prevHash] = append(bp.parentOrphans[prevHash], oBlock)

	bp.logger.Infof("add orphan block %s prev %s", blockHash, prevHash)

	bp.requestOrphanBlock(oBlock)
}

func (bp *PovBlockProcessor) removeOrphanBlock(orphanBlock *PovOrphanBlock) {
	orphanHash := orphanBlock.blockSrc.block.GetHash()
	delete(bp.orphanBlocks, orphanHash)

	prevHash := orphanBlock.blockSrc.block.GetPrevious()
	orphans := bp.parentOrphans[prevHash]
	for i := 0; i < len(orphans); i++ {
		hash := orphans[i].blockSrc.block.GetHash()
		if hash == orphanHash {
			copy(orphans[i:], orphans[i+1:])
			orphans[len(orphans)-1] = nil
			orphans = orphans[:len(orphans)-1]
			i--
		}
	}

	bp.parentOrphans[prevHash] = orphans

	if len(bp.parentOrphans[prevHash]) == 0 {
		delete(bp.parentOrphans, prevHash)
	}
}

func (bp *PovBlockProcessor) HasOrphanBlock(blockHash types.Hash) bool {
	if bp.orphanBlocks[blockHash] != nil {
		return true
	}
	return false
}

func (bp *PovBlockProcessor) enqueueOrphanBlocks(blockSrc *PovBlockSource) error {
	blockHash := blockSrc.block.GetHash()
	orphans, ok := bp.parentOrphans[blockHash]
	if !ok {
		return nil
	}
	if len(orphans) <= 0 {
		delete(bp.parentOrphans, blockHash)
		return nil
	}

	bp.logger.Debugf("parent %s has %d orphan blocks", blockHash, len(orphans))

	for i := 0; i < len(orphans); i++ {
		orphan := orphans[i]
		if orphan == nil {
			continue
		}

		orphanHash := orphan.blockSrc.block.GetHash()
		bp.removeOrphanBlock(orphan)
		i--

		bp.logger.Debugf("move orphan block %s to queue", orphanHash)
		bp.deOrphanBlocks = append(bp.deOrphanBlocks, orphan.blockSrc)
	}

	return nil
}

func (bp *PovBlockProcessor) GetOrphanRoot(oHash types.Hash) types.Hash {
	// Keep looping while the parent of each orphaned block is
	// known and is an orphan itself.
	oRoot := oHash
	for {
		oBlock, exists := bp.orphanBlocks[oRoot]
		if !exists {
			break
		}
		oRoot = oBlock.blockSrc.block.GetPrevious()
	}
	return oRoot
}

func (bp *PovBlockProcessor) requestOrphanBlock(oBlock *PovOrphanBlock) {
	// no need request orphan in syncing phase
	if bp.syncState.IsSyncExited() == false {
		return
	}

	blockHash := oBlock.blockSrc.block.GetHash()

	bp.reqOrphanBlocks[blockHash] = oBlock
}

func (bp *PovBlockProcessor) onRequestOrphanBlocksTimer() {
	if len(bp.reqOrphanBlocks) <= 0 {
		return
	}
	// no need request orphan in syncing phase
	if bp.syncState.IsSyncExited() == false {
		return
	}

	// find orphan roots that are previous blocks
	peerOrphanRoots := make(map[string]map[types.Hash]struct{})
	for _, oBlock := range bp.reqOrphanBlocks {
		blockHash := oBlock.blockSrc.block.GetHash()
		oRoot := bp.GetOrphanRoot(blockHash)

		orphanRoots := peerOrphanRoots[oBlock.blockSrc.peerID]
		if orphanRoots == nil {
			orphanRoots = make(map[types.Hash]struct{})
			peerOrphanRoots[oBlock.blockSrc.peerID] = orphanRoots
		}
		orphanRoots[oRoot] = struct{}{}
	}

	// just request once by fast timer, retry many times by slow timer
	bp.reqOrphanBlocks = make(map[types.Hash]*PovOrphanBlock)

	for peerID, orphanRoots := range peerOrphanRoots {
		// check orphan roots are waiting txs or not
		var reqOrphanRoots []*types.Hash
		for oRoot := range orphanRoots {
			if bp.HasPendingBlock(oRoot) {
				continue
			}

			oRootCopy := oRoot
			reqOrphanRoots = append(reqOrphanRoots, &oRootCopy)
		}

		bp.syncer.requestBlocksByHashes(reqOrphanRoots, peerID)
	}
}

func (bp *PovBlockProcessor) onCheckOrphanBlocksTimer() {
	if len(bp.orphanBlocks) <= 0 {
		return
	}
	// no need request orphan in syncing phase
	if bp.syncState.IsSyncExited() == false {
		return
	}

	nowTime := time.Now()

	orphanRoots := make(map[types.Hash]struct{})
	for _, oBlock := range bp.orphanBlocks {
		// too old orphan block should be remove it
		if nowTime.After(oBlock.addTime.Add(time.Hour)) {
			continue
		}

		// too new orphan block should be ignore it
		if nowTime.Before(oBlock.lastPullTime.Add(minPullBlockIntervalSec * time.Second)) {
			continue
		}
		oBlock.lastPullTime = nowTime

		blockHash := oBlock.blockSrc.block.GetHash()
		oRoot := bp.GetOrphanRoot(blockHash)
		orphanRoots[oRoot] = struct{}{}
	}

	bp.logger.Infof("check orphan blocks %d, orphan roots %d", len(bp.orphanBlocks), len(orphanRoots))

	var reqOrphanRoots []*types.Hash
	for oRoot := range orphanRoots {
		if bp.HasPendingBlock(oRoot) {
			continue
		}

		oRootCopy := oRoot
		reqOrphanRoots = append(reqOrphanRoots, &oRootCopy)
	}

	bp.syncer.requestBlocksByHashes(reqOrphanRoots, "")
}

func (bp *PovBlockProcessor) addTxPendingBlock(blockSrc *PovBlockSource, stat *PovVerifyStat) {
	bp.txPendingMux.Lock()
	defer bp.txPendingMux.Unlock()

	blockHash := blockSrc.block.GetHash()
	pendingBlock := &PovPendingBlock{
		addTime:   time.Now(),
		blockSrc:  blockSrc,
		txResults: stat.TxResults,
	}

	bp.logger.Infof("add tx pending block %s txs %d", blockHash, len(stat.TxResults))

	var reqTxHashes []*types.Hash

	for txHashTmp, result := range stat.TxResults {
		if result == process.GapTransaction {
			txHash := txHashTmp
			txPendEntry, ok := bp.txPendingBlocks[txHash]
			if !ok {
				txPendEntry = &PovTxPendingEntry{
					lastPullTime: time.Now(),
				}
				bp.txPendingBlocks[txHash] = txPendEntry
			}
			txPendEntry.pendingBlocks = append(txPendEntry.pendingBlocks, pendingBlock)

			reqTxHashes = append(reqTxHashes, &txHash)
		}
	}

	bp.pendingBlocks[blockHash] = pendingBlock

	bp.syncer.requestTxsByHashes(reqTxHashes, blockSrc.peerID)
}

func (bp *PovBlockProcessor) removeTxPendingBlockNoLock(pendingBlock *PovPendingBlock) {
	blockHash := pendingBlock.blockSrc.block.GetHash()
	bp.logger.Infof("remove tx pending block %s txs %d", blockHash, len(pendingBlock.txResults))

	for txHash := range pendingBlock.txResults {
		delete(bp.txPendingBlocks, txHash)
	}
	delete(bp.pendingBlocks, blockHash)
}

func (bp *PovBlockProcessor) HasPendingBlock(blockHash types.Hash) bool {
	bp.txPendingMux.Lock()
	defer bp.txPendingMux.Unlock()

	if bp.pendingBlocks[blockHash] != nil {
		return true
	}
	return false
}

func (bp *PovBlockProcessor) onCheckTxPendingBlocksTimer() {
	bp.txPendingMux.Lock()
	defer bp.txPendingMux.Unlock()

	nowTime := time.Now()

	txPendingNum := len(bp.txPendingBlocks)
	blockPendingNum := len(bp.pendingBlocks)
	if txPendingNum+blockPendingNum > 0 {
		bp.logger.Infof("check tx pending, txs %d blocks %d", txPendingNum, blockPendingNum)
	}

	var reqTxHashes []*types.Hash

	for txHash, txPendEntry := range bp.txPendingBlocks {
		if len(txPendEntry.pendingBlocks) <= 0 {
			delete(bp.txPendingBlocks, txHash)
			continue
		}

		txBlock, _ := bp.ledger.GetStateBlock(txHash)
		// tx is not exist to pull from some peer
		if txBlock == nil {
			if nowTime.After(txPendEntry.lastPullTime.Add(minPullTxIntervalSec * time.Second)) {
				txHashCopy := txHash
				reqTxHashes = append(reqTxHashes, &txHashCopy)
				txPendEntry.lastPullTime = nowTime
			}
			continue
		}

		// tx is exist to release pending blocks
		for _, pendingBlock := range txPendEntry.pendingBlocks {
			delete(pendingBlock.txResults, txHash)
		}

		delete(bp.txPendingBlocks, txHash)
	}

	for _, pendingBlock := range bp.pendingBlocks {
		if len(pendingBlock.txResults) <= 0 {
			bp.releaseTxPendingBlock(pendingBlock)
		}
	}

	bp.syncer.requestTxsByHashes(reqTxHashes, "")
}

func (bp *PovBlockProcessor) releaseTxPendingBlock(pendingBlock *PovPendingBlock) {
	blockHash := pendingBlock.blockSrc.block.GetHash()
	bp.logger.Infof("release tx pending block %s", blockHash)

	delete(bp.pendingBlocks, blockHash)

	if bp.checkAndSetBlockInChan(pendingBlock.blockSrc.block) {
		return
	}
	bp.blockHighCh <- pendingBlock.blockSrc
}
