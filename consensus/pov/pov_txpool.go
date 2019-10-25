package pov

import (
	"container/list"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/trie"
)

const (
	MaxTxsInPool = 100000
)

type PovTxEvent struct {
	txHash  types.Hash
	txBlock *types.StateBlock
	event   common.TopicType
}

type PovTxEntry struct {
	txHash  types.Hash
	txBlock *types.StateBlock
}

type PovTxPool struct {
	logger      *zap.SugaredLogger
	eb          event.EventBus
	handlerIds  map[common.TopicType]string //topic->handler id
	ledger      ledger.Store
	chain       PovTxChainReader
	txMu        sync.RWMutex
	txEventCh   chan *PovTxEvent
	quitCh      chan struct{}
	accountTxs  map[types.AddressToken]*list.List
	allTxs      map[types.Hash]*PovTxEntry
	lastUpdated int64
	syncState   atomic.Value
}

type PovTxChainReader interface {
	RegisterListener(listener EventListener)
	UnRegisterListener(listener EventListener)
	GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState
}

func NewPovTxPool(eb event.EventBus, l ledger.Store, chain PovTxChainReader) *PovTxPool {
	txPool := &PovTxPool{
		logger:     log.NewLogger("pov_txpool"),
		eb:         eb,
		handlerIds: make(map[common.TopicType]string),
		ledger:     l,
		chain:      chain,
	}
	txPool.txEventCh = make(chan *PovTxEvent, 5000)
	txPool.quitCh = make(chan struct{})
	txPool.accountTxs = make(map[types.AddressToken]*list.List)
	txPool.allTxs = make(map[types.Hash]*PovTxEntry)
	txPool.syncState.Store(common.SyncNotStart)
	return txPool
}

func (tp *PovTxPool) Init() error {
	return nil
}

func (tp *PovTxPool) Start() error {
	if tp.ledger != nil {
		ebL := tp.ledger.EventBus()
		id, err := ebL.SubscribeSync(common.EventAddRelation, tp.onAddStateBlock)
		if err != nil {
			tp.logger.Errorf("failed to subscribe EventAddRelation")
			return err
		}
		tp.handlerIds[common.EventAddRelation] = id
		id, err = ebL.SubscribeSync(common.EventAddSyncBlocks, tp.onAddSyncStateBlock)
		if err != nil {
			tp.logger.Errorf("failed to subscribe EventAddSyncBlocks")
			return err
		}
		tp.handlerIds[common.EventAddSyncBlocks] = id
		id, err = ebL.SubscribeSync(common.EventDeleteRelation, tp.onDeleteStateBlock)
		if err != nil {
			tp.logger.Errorf("failed to subscribe EventDeleteRelation")
			return err
		}
		tp.handlerIds[common.EventDeleteRelation] = id
	}

	if tp.eb != nil {
		id, err := tp.eb.SubscribeSync(common.EventPovSyncState, tp.onPovSyncState)
		if err != nil {
			tp.logger.Errorf("failed to subscribe EventPovSyncState")
			return err
		}
		tp.handlerIds[common.EventPovSyncState] = id
	}

	tp.chain.RegisterListener(tp)

	common.Go(tp.loop)

	return nil
}

func (tp *PovTxPool) Stop() {
	tp.chain.UnRegisterListener(tp)

	if tp.ledger != nil {
		ebL := tp.ledger.EventBus()
		err := ebL.Unsubscribe(common.EventAddRelation, tp.handlerIds[common.EventAddRelation])
		if err != nil {
			tp.logger.Error(err)
		}
		err = ebL.Unsubscribe(common.EventAddSyncBlocks, tp.handlerIds[common.EventAddSyncBlocks])
		if err != nil {
			tp.logger.Error(err)
		}
		err = ebL.Unsubscribe(common.EventDeleteRelation, tp.handlerIds[common.EventDeleteRelation])
		if err != nil {
			tp.logger.Error(err)
		}
	}

	if tp.eb != nil {
		err := tp.eb.Unsubscribe(common.EventPovSyncState, tp.handlerIds[common.EventPovSyncState])
		if err != nil {
			tp.logger.Error(err)
		}
	}

	close(tp.quitCh)
}

func (tp *PovTxPool) onAddStateBlock(block *types.StateBlock) {
	if !tp.isPovSyncDone() {
		return
	}
	if tp.isTxExceedLimit() {
		return
	}
	txHash := block.GetHash()
	tp.txEventCh <- &PovTxEvent{event: common.EventAddRelation, txHash: txHash, txBlock: block}
}

func (tp *PovTxPool) onAddSyncStateBlock(block *types.StateBlock, done bool) {
	if done {
		return
	}
	if !tp.isPovSyncDone() {
		return
	}
	if tp.isTxExceedLimit() {
		return
	}
	txHash := block.GetHash()
	tp.txEventCh <- &PovTxEvent{event: common.EventAddSyncBlocks, txHash: txHash, txBlock: block}
}

func (tp *PovTxPool) onDeleteStateBlock(hash types.Hash) {
	tp.txEventCh <- &PovTxEvent{event: common.EventDeleteRelation, txHash: hash}
}

func (tp *PovTxPool) OnPovBlockEvent(evt byte, block *types.PovBlock) {
	if !tp.isPovSyncDone() {
		return
	}

	txs := block.GetAccountTxs()
	if len(txs) == 0 {
		return
	}

	if evt == EventConnectPovBlock {
		tp.logger.Debugf("connect pov block %s delete txs %d", block.GetHash(), len(txs))
		for _, tx := range txs {
			tp.delTx(tx.Hash)
		}
	} else if evt == EventDisconnectPovBlock {
		tp.logger.Debugf("disconnect pov block %s add txs %d", block.GetHash(), len(txs))
		for _, tx := range txs {
			tp.addTx(tx.Hash, tx.Block)
		}
	}
}

func (tp *PovTxPool) onPovSyncState(state common.SyncState) {
	tp.syncState.Store(state)
}

func (tp *PovTxPool) isPovSyncDone() bool {
	if tp.syncState.Load().(common.SyncState) == common.SyncDone {
		return true
	}
	return false
}

func (tp *PovTxPool) isTxExceedLimit() bool {
	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	txNum := len(tp.allTxs)
	if txNum >= MaxTxsInPool {
		return true
	}

	return false
}

func (tp *PovTxPool) loop() {
	// wait pov init sync done
	for {
		if tp.isPovSyncDone() {
			tp.recoverUnconfirmedTxs()
			break
		}
		time.Sleep(time.Second)
	}

	checkTicker := time.NewTicker(100 * time.Millisecond)
	defer checkTicker.Stop()

	txProcNum := 0

	for {
		select {
		case <-tp.quitCh:
			return
		case txEvent := <-tp.txEventCh:
			tp.processTxEvent(txEvent)
			txProcNum++
		case <-checkTicker.C:
			if txProcNum >= 100 {
				time.Sleep(1 * time.Millisecond)
			}
			txProcNum = 0
		}
	}
}

func (tp *PovTxPool) getUnconfirmedTxsByFast() (map[types.AddressToken][]*PovTxEntry, int) {
	stateBlockNum, _ := tp.ledger.CountStateBlocks()
	uncheckedStateBlockNum, _ := tp.ledger.CountUncheckedBlocks()

	tp.logger.Infof("begin to scan all state blocks, block %d, unchecked %d", stateBlockNum, uncheckedStateBlockNum)

	var accountTxs map[types.AddressToken][]*PovTxEntry
	unconfirmedTxNum := 0

	startTime := time.Now()

	err := tp.ledger.BatchView(func(txn db.StoreTxn) error {
		// scan all account metas
		var allAms []*types.AccountMeta
		err := tp.ledger.GetAccountMetas(func(am *types.AccountMeta) error {
			allAms = append(allAms, am)
			return nil
		}, txn)
		if err != nil {
			tp.logger.Errorf("failed to get account metas, err %s", err)
			return err
		}

		tp.logger.Infof("total %d account metas", len(allAms))

		accountTxs = make(map[types.AddressToken][]*PovTxEntry, len(allAms))

		// scan all blocks under user accounts
		for _, am := range allAms {
			for _, tm := range am.Tokens {
				prevHash := tm.Header
				for !prevHash.IsZero() {
					block, err := tp.ledger.GetStateBlock(prevHash, txn)
					if err != nil {
						break
					}

					txHash := prevHash // no need calc hash from block
					prevHash = block.GetPrevious()

					if tp.ledger.HasPovTxLookup(txHash, txn) {
						break // all previous tx had been packed into pov
					}

					addrToken := types.AddressToken{Address: block.GetAddress(), Token: block.GetToken()}
					accountTxs[addrToken] = append(accountTxs[addrToken], &PovTxEntry{txHash: txHash, txBlock: block})
					unconfirmedTxNum++
				}
			}
		}

		tp.logger.Infof("total %d unconfirmed blocks by scan %d account metas", unconfirmedTxNum, len(allAms))

		// reversing txs, open <- send/recv <- header
		for _, tokenTxs := range accountTxs {
			for i := len(tokenTxs)/2 - 1; i >= 0; i-- {
				opp := len(tokenTxs) - 1 - i
				tokenTxs[i], tokenTxs[opp] = tokenTxs[opp], tokenTxs[i]
			}
		}

		// scan all genesis blocks under contract accounts
		allGenesisBlocks := common.AllGenesisBlocks()
		for _, block := range allGenesisBlocks {
			if !types.IsContractAddress(block.GetAddress()) {
				continue
			}

			txHash := block.GetHash()
			if tp.ledger.HasPovTxLookup(txHash, txn) {
				continue
			}

			addrToken := types.AddressToken{Address: block.GetAddress(), Token: block.GetToken()}
			blockCopy := block
			accountTxs[addrToken] = append(accountTxs[addrToken], &PovTxEntry{txHash: txHash, txBlock: &blockCopy})
			unconfirmedTxNum++
		}

		return nil
	})
	if err != nil {
		tp.logger.Errorf("scan all state blocks failed")
	}

	usedTime := time.Since(startTime)

	tp.logger.Infof("finished to scan all state blocks used time %s, unconfirmed %d", usedTime.String(), unconfirmedTxNum)

	return accountTxs, unconfirmedTxNum
}

func (tp *PovTxPool) recoverUnconfirmedTxs() {
	unpackAccountTxs, unconfirmedTxNum := tp.getUnconfirmedTxsByFast()
	if unconfirmedTxNum <= 0 {
		return
	}

	txStepNum := 5000
	txAddNum := 0
	for _, accountTxs := range unpackAccountTxs {
		for _, txEntry := range accountTxs {
			tp.addTx(txEntry.txHash, txEntry.txBlock)
			txAddNum++
			if txAddNum%txStepNum == 0 {
				tp.logger.Infof("total %d unconfirmed state blocks have been added to tx pool", txAddNum)
			}
		}
	}

	tp.logger.Infof("total %d unconfirmed state blocks have been added to tx pool", txAddNum)
}

func (tp *PovTxPool) processTxEvent(txEvent *PovTxEvent) {
	if txEvent.event == common.EventAddRelation || txEvent.event == common.EventAddSyncBlocks {
		tp.addTx(txEvent.txHash, txEvent.txBlock)
	} else if txEvent.event == common.EventDeleteRelation {
		tp.delTx(txEvent.txHash)
	}
}

func (tp *PovTxPool) addTx(txHash types.Hash, txBlock *types.StateBlock) {
	if txBlock == nil {
		tp.logger.Errorf("add tx %s but block is nil", txHash)
		return
	}

	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	txExist, ok := tp.allTxs[txHash]
	if ok && txExist != nil {
		return
	}

	addrToken := types.AddressToken{Address: txBlock.GetAddress(), Token: txBlock.GetToken()}

	accTxList, ok := tp.accountTxs[addrToken]
	if !ok {
		accTxList = list.New()
		tp.accountTxs[addrToken] = accTxList
	}

	txEntry := &PovTxEntry{txHash: txHash, txBlock: txBlock}

	var childE *list.Element
	var parentE *list.Element

	prevHash := txBlock.GetPrevious()
	if prevHash.IsZero() {
		// contract address's blocks are all independent, no previous
		if types.IsContractAddress(txBlock.GetAddress()) {
		} else {
			// open block should be first
			childE = accTxList.Front()
		}
	} else {
		for itE := accTxList.Back(); itE != nil; itE = itE.Prev() {
			itTxEntry := itE.Value.(*PovTxEntry)
			if itTxEntry.txHash == prevHash {
				parentE = itE
				break
			}
			if itTxEntry.txBlock.GetPrevious() == txHash {
				childE = itE
				break
			}
		}
	}

	if parentE != nil {
		accTxList.InsertAfter(txEntry, parentE)
	} else if childE != nil {
		accTxList.InsertBefore(txEntry, childE)
	} else {
		accTxList.PushBack(txEntry)
	}
	tp.allTxs[txHash] = txEntry

	tp.lastUpdated = time.Now().Unix()
}

func (tp *PovTxPool) delTx(txHash types.Hash) {
	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	txEntry, ok := tp.allTxs[txHash]
	if !ok {
		return
	}

	delete(tp.allTxs, txHash)

	if txEntry == nil {
		return
	}
	txBlock := txEntry.txBlock

	addrToken := types.AddressToken{Address: txBlock.GetAddress(), Token: txBlock.GetToken()}
	accTxList, ok := tp.accountTxs[addrToken]
	if ok {
		var foundEle *list.Element
		for e := accTxList.Front(); e != nil; e = e.Next() {
			if e.Value == txEntry {
				foundEle = e
				break
			}
		}
		if foundEle != nil {
			accTxList.Remove(foundEle)
		}

		if accTxList.Len() <= 0 {
			delete(tp.accountTxs, addrToken)
		}
	}

	tp.lastUpdated = time.Now().Unix()
}

func (tp *PovTxPool) getTx(txHash types.Hash) *types.StateBlock {
	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	txEntry, ok := tp.allTxs[txHash]
	if !ok {
		return nil
	}

	return txEntry.txBlock
}

func (tp *PovTxPool) SelectPendingTxs(stateTrie *trie.Trie, limit int) []*types.StateBlock {
	return tp.selectPendingTxsByFair(stateTrie, limit)
}

func (tp *PovTxPool) selectPendingTxsByFair(stateTrie *trie.Trie, limit int) []*types.StateBlock {
	tp.txMu.RLock()
	defer tp.txMu.RUnlock()

	var retTxs []*types.StateBlock

	if limit <= 0 || len(tp.accountTxs) == 0 {
		return retTxs
	}
	addrTokenNum := len(tp.accountTxs)

	addrTokenNeedScans := make(map[types.AddressToken]struct{}, addrTokenNum)
	addrTokenPrevHashes := make(map[types.AddressToken]types.Hash, addrTokenNum)
	addrTokenSelectedTxHashes := make(map[types.AddressToken]map[types.Hash]struct{})
	addrTokenScanIters := make(map[types.AddressToken]*list.Element, addrTokenNum)
	addrTokenIsCA := make(map[types.AddressToken]bool)

	for addrToken, accTxList := range tp.accountTxs {
		if accTxList.Len() > 0 {
			addrTokenNeedScans[addrToken] = struct{}{}

			addrTokenIsCA[addrToken] = types.IsContractAddress(addrToken.Address)
		}
	}

LoopMain:
	for {
		if len(addrTokenNeedScans) == 0 {
			break
		}

	LoopAddrToken:
		for addrToken := range addrTokenNeedScans {
			accTxList := tp.accountTxs[addrToken]
			if accTxList.Len() == 0 {
				delete(addrTokenNeedScans, addrToken)
				continue
			}

			selectedTxHashes := addrTokenSelectedTxHashes[addrToken]
			if selectedTxHashes == nil {
				selectedTxHashes = make(map[types.Hash]struct{})
				addrTokenSelectedTxHashes[addrToken] = selectedTxHashes
			}
			if len(selectedTxHashes) >= accTxList.Len() {
				delete(addrTokenNeedScans, addrToken)
				continue
			}

			isCA := addrTokenIsCA[addrToken]

			notInOrderTxNum := 0
			inOrderTxNum := 0
			accTxIter := addrTokenScanIters[addrToken]
			if accTxIter == nil {
				accTxIter = accTxList.Front()
			} else {
				accTxIter = accTxIter.Next()
			}
		LoopAccTx:
			for ; accTxIter != nil; accTxIter = accTxIter.Next() {
				txEntry := accTxIter.Value.(*PovTxEntry)
				txBlock := txEntry.txBlock
				txHash := txEntry.txHash
				if _, ok := selectedTxHashes[txHash]; ok {
					continue
				}

				// contract address's blocks are all independent, no previous
				prevHashWant, ok := addrTokenPrevHashes[addrToken]
				if !ok {
					if isCA {
						prevHashWant = types.ZeroHash
					} else {
						as := tp.chain.GetAccountState(stateTrie, addrToken.Address)
						if as != nil {
							rs := as.GetTokenState(addrToken.Token)
							if rs != nil {
								prevHashWant = rs.Hash
							} else {
								prevHashWant = types.ZeroHash
							}
						} else {
							prevHashWant = types.ZeroHash
						}
					}
				}

				if txBlock.GetPrevious() == prevHashWant {
					retTxs = append(retTxs, txBlock)

					selectedTxHashes[txHash] = struct{}{}
					// contract address's blocks are all independent, no previous
					if !isCA {
						addrTokenPrevHashes[addrToken] = txHash
					}
					inOrderTxNum++

					// choose 1 tx in each loop
					limit--
					break LoopAccTx
				} else {
					notInOrderTxNum++
				}
			}
			if accTxIter == nil {
				addrTokenScanIters[addrToken] = accTxList.Back()
			} else {
				addrTokenScanIters[addrToken] = accTxIter
			}

			// check account token is skip or not
			if inOrderTxNum == 0 {
				delete(addrTokenNeedScans, addrToken)
				if notInOrderTxNum > 0 {
					tp.logger.Debugf("addrToken %s has txs %d not in order", addrToken, notInOrderTxNum)
				}
			}

			if limit == 0 {
				break LoopAddrToken
			}
		}

		if limit == 0 {
			break LoopMain
		}
	}

	return retTxs
}

func (tp *PovTxPool) LastUpdated() time.Time {
	return time.Unix(tp.lastUpdated, 0)
}

func (tp *PovTxPool) GetPendingTxNum() uint32 {
	tp.txMu.RLock()
	defer tp.txMu.RUnlock()

	return uint32(len(tp.allTxs))
}

func (tp *PovTxPool) GetDebugInfo() map[string]interface{} {
	// !!! be very careful about to map concurrent read !!!

	info := make(map[string]interface{})
	info["allTxs"] = len(tp.allTxs)
	info["accountTxs"] = len(tp.accountTxs)

	return info
}
