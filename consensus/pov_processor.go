package consensus

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"
	"sync"
	"time"
)

const (
	blockChanSize           = 1024
	maxOrphanBlocks         = 1000
	checkTxPendingTickerSec = 5
)

type PovBlockResult struct {
	err error
}

type PovBlockSource struct {
	block   *types.PovBlock
	from    types.PovBlockFrom
	replyCh chan PovBlockResult
}

type PovOrphanBlock struct {
	blockSrc   *PovBlockSource
	expiration time.Time
}

type PovPendingBlock struct {
	addTime   time.Time
	checkCnt  int
	blockSrc  *PovBlockSource
	txResults map[types.Hash]process.ProcessResult
}

type PovBlockProcessor struct {
	povEngine *PoVEngine

	orphanBlocks  map[types.Hash]*PovOrphanBlock   // blockHash -> block
	parentOrphans map[types.Hash][]*PovOrphanBlock // blockHash -> child blocks
	oldestOrphan  *PovOrphanBlock

	pendingBlocks   map[types.Hash]*PovPendingBlock // blockHash -> block
	txPendingBlocks map[types.Hash]*PovPendingBlock // txHash -> block
	txPendingMux    sync.Mutex

	waitingBlocks []*PovBlockSource
	blockCh       chan *PovBlockSource
	quitCh        chan struct{}
}

func NewPovBlockProcessor(povEngine *PoVEngine) *PovBlockProcessor {
	bp := &PovBlockProcessor{
		povEngine: povEngine,
	}

	bp.orphanBlocks = make(map[types.Hash]*PovOrphanBlock)
	bp.parentOrphans = make(map[types.Hash][]*PovOrphanBlock)
	bp.pendingBlocks = make(map[types.Hash]*PovPendingBlock)
	bp.txPendingBlocks = make(map[types.Hash]*PovPendingBlock)

	bp.waitingBlocks = make([]*PovBlockSource, 0, 1000)
	bp.blockCh = make(chan *PovBlockSource, blockChanSize)
	bp.quitCh = make(chan struct{})

	return bp
}

func (bp *PovBlockProcessor) Start() error {
	eb := bp.povEngine.GetEventBus()
	if eb != nil {
		eb.Subscribe(string(common.EventAddRelation), bp.onAddStateBlock)
	}

	common.Go(bp.loop)
	common.Go(bp.loopCheckTxPending)
	return nil
}

func (bp *PovBlockProcessor) Init() error {
	return nil
}

func (bp *PovBlockProcessor) Stop() error {
	eb := bp.povEngine.GetEventBus()
	if eb != nil {
		eb.Unsubscribe(string(common.EventAddRelation), bp.onAddStateBlock)
	}

	close(bp.quitCh)
	return nil
}

func (bp *PovBlockProcessor) onAddStateBlock(tx *types.StateBlock) {
	bp.txPendingMux.Lock()
	defer bp.txPendingMux.Unlock()

	txHash := tx.GetHash()

	pendingBlock := bp.txPendingBlocks[txHash]
	if pendingBlock == nil {
		return
	}

	bp.povEngine.GetLogger().Debugf("recv add state block %s", txHash)

	delete(bp.txPendingBlocks, txHash)

	if _, ok := pendingBlock.txResults[txHash]; !ok {
		return
	}
	delete(pendingBlock.txResults, txHash)

	if len(pendingBlock.txResults) <= 0 {
		bp.releaseTxPendingBlock(pendingBlock)
	}
}

func (bp *PovBlockProcessor) onRecvPovSyncState(state common.SyncState) {
	if state.IsSyncExited() {
		for _, blockSrc := range bp.waitingBlocks {
			bp.blockCh <- blockSrc
		}
		bp.waitingBlocks = nil
	}
}

func (bp *PovBlockProcessor) AddBlock(block *types.PovBlock, from types.PovBlockFrom) error {
	blockSrc := &PovBlockSource{block: block, from: from}

	needWait := false
	if from == types.PovBlockFromRemoteBroadcast {
		ss := bp.povEngine.GetSyncState()
		if !ss.IsSyncExited() {
			needWait = true
		}
	}

	if needWait {
		if len(bp.waitingBlocks) < cap(bp.waitingBlocks) {
			bp.waitingBlocks = append(bp.waitingBlocks, blockSrc)
		}
	} else {
		bp.blockCh <- blockSrc
	}

	return nil
}

func (bp *PovBlockProcessor) AddMinedBlock(block *types.PovBlock) error {
	replyCh := make(chan PovBlockResult)
	bp.blockCh <- &PovBlockSource{block: block, from: types.PovBlockFromLocal, replyCh: replyCh}
	result := <-replyCh
	close(replyCh)
	return result.err
}

func (bp *PovBlockProcessor) loop() {
	for {
		select {
		case <-bp.quitCh:
			bp.povEngine.GetLogger().Info("Exiting process blocks")
			return
		case blockSrc := <-bp.blockCh:
			err := bp.processBlock(blockSrc)
			if blockSrc.replyCh != nil {
				blockSrc.replyCh <- PovBlockResult{err: err}
			}
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
			bp.checkTxPendingBlocks()
		}
	}
}

func (bp *PovBlockProcessor) processBlock(blockSrc *PovBlockSource) error {
	block := blockSrc.block
	blockHash := blockSrc.block.GetHash()
	bp.povEngine.GetLogger().Debugf("process block, %d/%s", blockSrc.block.GetHeight(), blockHash)

	chain := bp.povEngine.GetChain()

	// check duplicate block
	if bp.HasOrphanBlock(blockHash) {
		bp.povEngine.GetLogger().Debugf("duplicate block %s exist in orphan", blockHash)
		return nil
	}
	if bp.HasPendingBlock(blockHash) {
		bp.povEngine.GetLogger().Debugf("duplicate block %s exist in pending", blockHash)
		return nil
	}
	if chain.HasBestBlock(blockHash, block.GetHeight()) {
		bp.povEngine.GetLogger().Debugf("duplicate block %s exist in best chain", blockHash)
		return nil
	}

	prevBlock := chain.GetBlockByHash(block.GetPrevious())
	if prevBlock == nil {
		bp.addOrphanBlock(blockSrc)
		return nil
	}

	// check block
	stat := bp.povEngine.GetVerifier().VerifyFull(block)
	if stat == nil {
		bp.povEngine.GetLogger().Errorf("failed to verify block %s", block.GetHash())
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
		bp.povEngine.GetLogger().Errorf("failed to verify block %s, result %s, err %s", block.GetHash(), stat.Result, stat.ErrMsg)
		return ErrPovFailedVerify
	}

	err := bp.povEngine.GetChain().InsertBlock(block, stat.StateTrie)

	if err == nil {
		_ = bp.processOrphanBlock(blockSrc)
	}

	return err
}

func (bp *PovBlockProcessor) addOrphanBlock(blockSrc *PovBlockSource) {
	blockHash := blockSrc.block.GetHash()

	for _, oBlock := range bp.orphanBlocks {
		if time.Now().After(oBlock.expiration) {
			bp.removeOrphanBlock(oBlock)
			continue
		}

		if bp.oldestOrphan == nil || oBlock.expiration.Before(bp.oldestOrphan.expiration) {
			bp.oldestOrphan = oBlock
		}
	}

	if len(bp.orphanBlocks)+1 > maxOrphanBlocks {
		bp.removeOrphanBlock(bp.oldestOrphan)
		bp.oldestOrphan = nil
	}

	expiration := time.Now().Add(time.Hour)
	oBlock := &PovOrphanBlock{
		blockSrc:   blockSrc,
		expiration: expiration,
	}
	bp.orphanBlocks[blockHash] = oBlock

	prevHash := blockSrc.block.GetPrevious()
	bp.parentOrphans[prevHash] = append(bp.parentOrphans[prevHash], oBlock)

	bp.povEngine.GetLogger().Infof("add orphan block %s prev %s", blockHash, prevHash)

	orphanRoot := bp.GetOrphanRoot(blockHash)
	if bp.HasPendingBlock(orphanRoot) == false {
		bp.povEngine.GetSyncer().requestBlocksByHash(orphanRoot, 1, false)
	}
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

func (bp *PovBlockProcessor) processOrphanBlock(blockSrc *PovBlockSource) error {
	blockHash := blockSrc.block.GetHash()
	orphans, ok := bp.parentOrphans[blockHash]
	if !ok {
		return nil
	}
	if len(orphans) <= 0 {
		delete(bp.parentOrphans, blockHash)
		return nil
	}

	processHashes := make([]*types.Hash, 0, 10)
	processHashes = append(processHashes, &blockHash)
	for len(processHashes) > 0 {
		processHash := processHashes[0]
		processHashes[0] = nil
		processHashes = processHashes[1:]

		orphans := bp.parentOrphans[*processHash]

		bp.povEngine.GetLogger().Infof("parent %s has %d orphan blocks", processHash, len(orphans))

		for i := 0; i < len(orphans); i++ {
			orphan := orphans[i]
			if orphan == nil {
				continue
			}

			orphanHash := orphan.blockSrc.block.GetHash()
			bp.removeOrphanBlock(orphan)
			i--

			bp.blockCh <- orphan.blockSrc

			processHashes = append(processHashes, &orphanHash)
		}
	}

	return nil
}

func (bp *PovBlockProcessor) GetOrphanRoot(hash types.Hash) types.Hash {
	// Keep looping while the parent of each orphaned block is
	// known and is an orphan itself.
	orphanRoot := hash
	orphanHash := hash
	for {
		orphan, exists := bp.orphanBlocks[orphanHash]
		if !exists {
			break
		}
		orphanRoot = orphan.blockSrc.block.GetPrevious()
		orphanHash = orphanRoot
	}
	return orphanRoot
}

func (bp *PovBlockProcessor) addTxPendingBlock(blockSrc *PovBlockSource, stat *process.PovVerifyStat) {
	bp.txPendingMux.Lock()
	defer bp.txPendingMux.Unlock()

	blockHash := blockSrc.block.GetHash()
	pendingBlock := &PovPendingBlock{
		addTime:   time.Now(),
		blockSrc:  blockSrc,
		txResults: stat.TxResults,
	}

	bp.povEngine.GetLogger().Infof("add tx pending block %s txs %d", blockHash, len(stat.TxResults))

	for txHashTmp, result := range stat.TxResults {
		if result == process.GapTransaction {
			txHash := txHashTmp
			bp.txPendingBlocks[txHash] = pendingBlock

			bp.povEngine.GetSyncer().requestTxsByHash(txHash, 1)
		}
	}

	bp.pendingBlocks[blockHash] = pendingBlock
}

func (bp *PovBlockProcessor) removeTxPendingBlockNoLock(pendingBlock *PovPendingBlock) {
	blockHash := pendingBlock.blockSrc.block.GetHash()
	bp.povEngine.GetLogger().Infof("remove tx pending block %s txs %d", blockHash, len(pendingBlock.txResults))

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

func (bp *PovBlockProcessor) checkTxPendingBlocks() {
	bp.txPendingMux.Lock()
	defer bp.txPendingMux.Unlock()

	txPendingNum := len(bp.txPendingBlocks)
	blockPendingNum := len(bp.pendingBlocks)
	if txPendingNum+blockPendingNum > 0 {
		bp.povEngine.GetLogger().Infof("check tx pending, txs %d blocks %d", txPendingNum, blockPendingNum)
	}

	for txHash, pendingBlock := range bp.txPendingBlocks {
		txBlock, _ := bp.povEngine.GetLedger().GetStateBlock(txHash)
		if txBlock != nil {
			delete(pendingBlock.txResults, txHash)
			delete(bp.txPendingBlocks, txHash)
		}
	}

	for _, pendingBlock := range bp.pendingBlocks {
		if len(pendingBlock.txResults) <= 0 {
			bp.releaseTxPendingBlock(pendingBlock)
		} else {
			pendingBlock.checkCnt++
			if pendingBlock.checkCnt > 60 {
				bp.removeTxPendingBlockNoLock(pendingBlock)
			} else if pendingBlock.checkCnt/3 == 0 {
				for txHash := range pendingBlock.txResults {
					bp.povEngine.GetSyncer().requestTxsByHash(txHash, 1)
				}
			}
		}
	}
}

func (bp *PovBlockProcessor) releaseTxPendingBlock(pendingBlock *PovPendingBlock) {
	blockHash := pendingBlock.blockSrc.block.GetHash()
	bp.povEngine.GetLogger().Infof("release tx pending block %s", blockHash)

	for txHash := range pendingBlock.txResults {
		delete(bp.txPendingBlocks, txHash)
	}
	delete(bp.pendingBlocks, blockHash)

	go func() {
		time.Sleep(time.Second)
		bp.blockCh <- pendingBlock.blockSrc
	}()
}
