package pov

import (
	"container/list"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/config"

	"github.com/qlcchain/go-qlc/common/statedb"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/qlcchain/go-qlc/common/topic"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
)

const (
	MaxTxsInPool = 10000000

	maxSelectRunTimeInPool  = 60 * time.Second
	maxSelectKeepTimeInPool = 5 * time.Second
	maxSelectWaitTimeInPool = 120 * time.Second
)

type PovTxEvent struct {
	txHash  types.Hash
	txBlock *types.StateBlock
	event   topic.TopicType
}

type PovTxEntry struct {
	txHash  types.Hash
	txBlock *types.StateBlock
}

type txPoolWorkResponse struct {
	selectedTxs []*types.StateBlock
}

type txPoolWorkRequest struct {
	gsdb   *statedb.PovGlobalStateDB
	limit  int
	respCh chan *txPoolWorkResponse
}

type PovTxPool struct {
	logger      *zap.SugaredLogger
	eb          event.EventBus
	subscriber  *event.ActorSubscriber
	ledger      ledger.Store
	chain       PovTxChainReader
	txMu        sync.RWMutex
	txEventCh   chan *PovTxEvent
	quitCh      chan struct{}
	accountTxs  map[types.AddressToken]*list.List
	allTxs      map[types.Hash]*PovTxEntry
	lastUpdated time.Time
	syncState   atomic.Value
	poolMu      sync.Mutex

	workCh       chan *txPoolWorkRequest
	lastWorkRsp  *txPoolWorkResponse
	lastWorkTime time.Time

	statLimitTxNum     int64
	statNotInOrderNum  int64
	statMaxSelectTime  int64 // Microseconds
	statLastSelectTime int64 // Microseconds
}

type PovTxChainReader interface {
	RegisterListener(listener EventListener)
	UnRegisterListener(listener EventListener)
}

func NewPovTxPool(eb event.EventBus, l ledger.Store, chain PovTxChainReader) *PovTxPool {
	txPool := &PovTxPool{
		logger: log.NewLogger("pov_txpool"),
		eb:     eb,
		ledger: l,
		chain:  chain,
	}
	txPool.txEventCh = make(chan *PovTxEvent, 5000)
	txPool.quitCh = make(chan struct{})
	txPool.accountTxs = make(map[types.AddressToken]*list.List)
	txPool.allTxs = make(map[types.Hash]*PovTxEntry)
	txPool.syncState.Store(topic.SyncNotStart)
	txPool.workCh = make(chan *txPoolWorkRequest, 10)
	return txPool
}

func (tp *PovTxPool) Init() error {
	eb := tp.eb
	if eb == nil {
		eb = tp.ledger.EventBus()
	}
	tp.subscriber = event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *types.Tuple:
			tp.onAddSyncStateBlock(msg.First.(*types.StateBlock), msg.Second.(bool))
		case *types.StateBlock:
			tp.onAddStateBlock(msg)
		case topic.SyncState:
			tp.onPovSyncState(msg)
		}
	}), eb)

	if err := tp.subscriber.Subscribe(topic.EventAddRelation,
		topic.EventAddSyncBlocks, topic.EventPovSyncState); err != nil {
		return err
	}

	tp.chain.RegisterListener(tp)

	return nil
}

func (tp *PovTxPool) Start() error {
	common.Go(tp.loop)

	return nil
}

func (tp *PovTxPool) Stop() {
	tp.chain.UnRegisterListener(tp)
	if err := tp.subscriber.UnsubscribeAll(); err != nil {
		tp.logger.Error(err)
	}

	close(tp.quitCh)
}

func (tp *PovTxPool) onAddStateBlock(block *types.StateBlock) {
	if !tp.isPovSyncDone() {
		return
	}
	if tp.isTxExceedLimit() {
		tp.statLimitTxNum++
		return
	}
	txHash := block.GetHash()
	tp.txEventCh <- &PovTxEvent{event: topic.EventAddRelation, txHash: txHash, txBlock: block}
}

func (tp *PovTxPool) onAddSyncStateBlock(block *types.StateBlock, done bool) {
	if done {
		return
	}
	if !tp.isPovSyncDone() {
		return
	}
	if tp.isTxExceedLimit() {
		tp.statLimitTxNum++
		return
	}
	txHash := block.GetHash()
	tp.txEventCh <- &PovTxEvent{event: topic.EventAddSyncBlocks, txHash: txHash, txBlock: block}
}

func (tp *PovTxPool) onDeleteStateBlock(hash types.Hash) {
	tp.txEventCh <- &PovTxEvent{event: topic.EventDeleteRelation, txHash: hash}
}

func (tp *PovTxPool) OnPovBlockEvent(evt byte, block *types.PovBlock) {
	if !tp.isPovSyncDone() {
		return
	}

	txs := block.GetAccountTxs()
	if len(txs) == 0 {
		return
	}

	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	tp.lastWorkRsp = nil

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

func (tp *PovTxPool) onPovSyncState(state topic.SyncState) {
	tp.syncState.Store(state)
}

func (tp *PovTxPool) isPovSyncDone() bool {
	return tp.syncState.Load().(topic.SyncState) == topic.SyncDone
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

	for {
		select {
		case <-tp.quitCh:
			return
		case txEvent := <-tp.txEventCh:
			tp.processTxEvent(txEvent)
		case txWork := <-tp.workCh:
			tp.processTxWorkRequest(txWork)
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

	//err := tp.ledger.BatchView(func(txn db.StoreTxn) error {
	//	// scan all account metas
	var allAms []*types.AccountMeta
	err := tp.ledger.GetAccountMetas(func(am *types.AccountMeta) error {
		allAms = append(allAms, am)
		return nil
	})
	if err != nil {
		tp.logger.Errorf("failed to get account metas, err %s", err)
		//return err
	}

	tp.logger.Infof("total %d account metas", len(allAms))

	accountTxs = make(map[types.AddressToken][]*PovTxEntry, len(allAms))

	// scan all blocks under user accounts
	for _, am := range allAms {
		for _, tm := range am.Tokens {
			prevHash := tm.Header
			for !prevHash.IsZero() {
				block, err := tp.ledger.GetStateBlockConfirmed(prevHash)
				if err != nil {
					break
				}

				txHash := prevHash // no need calc hash from block
				prevHash = block.GetPrevious()

				if tp.ledger.HasPovTxLookup(txHash) {
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
	allGenesisBlocks := config.AllGenesisBlocks()
	for _, block := range allGenesisBlocks {
		if !contractaddress.IsContractAddress(block.GetAddress()) {
			continue
		}

		txHash := block.GetHash()
		if tp.ledger.HasPovTxLookup(txHash) {
			continue
		}

		addrToken := types.AddressToken{Address: block.GetAddress(), Token: block.GetToken()}
		blockCopy := block
		accountTxs[addrToken] = append(accountTxs[addrToken], &PovTxEntry{txHash: txHash, txBlock: &blockCopy})
		unconfirmedTxNum++
	}

	//	return nil
	//})
	if err != nil {
		tp.logger.Errorf("scan all state blocks failed")
	}

	usedTime := time.Since(startTime)

	tp.logger.Infof("finished to scan all state blocks used time %s, unconfirmed %d", usedTime.String(), unconfirmedTxNum)

	return accountTxs, unconfirmedTxNum
}

func (tp *PovTxPool) recoverUnconfirmedTxs() {
	tp.poolMu.Lock()
	defer tp.poolMu.Unlock()

	unpackAccountTxs, unconfirmedTxNum := tp.getUnconfirmedTxsByFast()
	if unconfirmedTxNum <= 0 {
		return
	}

	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	tp.lastWorkRsp = nil

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
	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	if txEvent.event == topic.EventAddRelation || txEvent.event == topic.EventAddSyncBlocks {
		tp.addTx(txEvent.txHash, txEvent.txBlock)
	} else if txEvent.event == topic.EventDeleteRelation {
		tp.delTx(txEvent.txHash)
	}

	if tp.lastWorkTime.Add(maxSelectKeepTimeInPool).Before(tp.lastUpdated) {
		tp.lastWorkRsp = nil
	}
}

func (tp *PovTxPool) addTx(txHash types.Hash, txBlock *types.StateBlock) {
	if txBlock == nil {
		tp.logger.Errorf("add tx %s but block is nil", txHash)
		return
	}

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
		if contractaddress.IsContractAddress(txBlock.GetAddress()) {
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

	tp.lastUpdated = time.Now()
}

func (tp *PovTxPool) delTx(txHash types.Hash) {
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

	tp.lastUpdated = time.Now()
}

func (tp *PovTxPool) getTx(txHash types.Hash) *types.StateBlock {
	txEntry, ok := tp.allTxs[txHash]
	if !ok {
		return nil
	}

	return txEntry.txBlock
}

func (tp *PovTxPool) SelectPendingTxs(gsdb *statedb.PovGlobalStateDB, limit int) []*types.StateBlock {
	// send request
	workReq := &txPoolWorkRequest{gsdb: gsdb, limit: limit}
	workReq.respCh = make(chan *txPoolWorkResponse, 0)

	select {
	case tp.workCh <- workReq:
		// ok, just wait response
	case <-time.After(maxSelectWaitTimeInPool):
		return nil
	}

	// receive response
	select {
	case workRsp := <-workReq.respCh:
		return workRsp.selectedTxs
	case <-time.After(maxSelectWaitTimeInPool):
		return nil
	}
}

func (tp *PovTxPool) selectPendingTxsByFair(gsdb *statedb.PovGlobalStateDB, limit int) []*types.StateBlock {
	var retTxs []*types.StateBlock

	if limit <= 0 || len(tp.accountTxs) == 0 {
		return retTxs
	}

	startTm := time.Now()
	defer func() {
		tp.statLastSelectTime = time.Since(startTm).Microseconds()
		if tp.statLastSelectTime > tp.statMaxSelectTime {
			tp.statMaxSelectTime = tp.statLastSelectTime
		}
	}()

	addrTokenNum := len(tp.accountTxs)

	addrTokenNeedScans := make(map[types.AddressToken]struct{}, addrTokenNum)
	addrTokenPrevHashes := make(map[types.AddressToken]types.Hash, addrTokenNum)
	addrTokenSelectedTxHashes := make(map[types.AddressToken]map[types.Hash]struct{})
	addrTokenScanIters := make(map[types.AddressToken]*list.Element, addrTokenNum)
	addrTokenIsCA := make(map[types.AddressToken]bool)

	for addrToken, accTxList := range tp.accountTxs {
		if accTxList.Len() > 0 {
			addrTokenNeedScans[addrToken] = struct{}{}

			addrTokenIsCA[addrToken] = contractaddress.IsContractAddress(addrToken.Address)
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
						as, _ := gsdb.GetAccountState(addrToken.Address)
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
					tp.statNotInOrderNum++
					tp.logger.Debugf("addrToken %s has txs %d not in order", addrToken, notInOrderTxNum)
				}
			}

			// check too much time for this selection
			if len(retTxs)%100 == 0 {
				if time.Since(startTm) >= maxSelectRunTimeInPool {
					limit = 0
				}
			}

			if limit == 0 {
				break LoopAddrToken
			}
		}

		if limit == 0 {
			break LoopMain
		}

		if time.Since(startTm) >= maxSelectRunTimeInPool {
			limit = 0
		}
	}

	return retTxs
}

func (tp *PovTxPool) processTxWorkRequest(workReq *txPoolWorkRequest) {
	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	if workReq.respCh == nil {
		return
	}

	if workReq.gsdb == nil {
		return
	}

	if tp.lastWorkRsp != nil {
		if tp.lastWorkTime.Add(maxSelectKeepTimeInPool).Before(tp.lastUpdated) {
			tp.lastWorkRsp = nil
		}
	}

	if tp.lastWorkRsp == nil {
		tp.lastWorkRsp = &txPoolWorkResponse{}
		tp.lastWorkRsp.selectedTxs = tp.selectPendingTxsByFair(workReq.gsdb, workReq.limit)
		tp.lastWorkTime = time.Now()
	}

	workReq.respCh <- tp.lastWorkRsp
}

func (tp *PovTxPool) LastUpdated() time.Time {
	return tp.lastUpdated
}

func (tp *PovTxPool) GetPendingTxNum() uint32 {
	return uint32(len(tp.allTxs))
}

func (tp *PovTxPool) GetDebugInfo() map[string]interface{} {
	// !!! be very careful about to map concurrent read !!!

	info := make(map[string]interface{})
	info["allTxs"] = len(tp.allTxs)
	info["accountTxs"] = len(tp.accountTxs)
	if tp.lastWorkRsp != nil {
		info["lastSelectedTxs"] = len(tp.lastWorkRsp.selectedTxs)
	}
	info["lastUpdated"] = tp.lastUpdated
	info["lastWorkTime"] = tp.lastWorkTime
	info["statLimitTxNum"] = tp.statLimitTxNum
	info["statNotInOrderNum"] = tp.statNotInOrderNum
	info["statLastSelectTime"] = tp.statLastSelectTime
	info["statMaxSelectTime"] = tp.statMaxSelectTime

	return info
}
