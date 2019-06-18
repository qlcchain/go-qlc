package consensus

import (
	"container/list"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/trie"
	"sync"
	"time"
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
	povEngine   *PoVEngine
	txMu        sync.RWMutex
	txEventCh   chan *PovTxEvent
	quitCh      chan struct{}
	accountTxs  map[types.AddressToken]*list.List
	allTxs      map[types.Hash]*PovTxEntry
	lastUpdated int64
}

func NewPovTxPool(povImpl *PoVEngine) *PovTxPool {
	txPool := &PovTxPool{
		povEngine: povImpl,
	}
	txPool.txEventCh = make(chan *PovTxEvent, 5000)
	txPool.quitCh = make(chan struct{})
	txPool.accountTxs = make(map[types.AddressToken]*list.List)
	txPool.allTxs = make(map[types.Hash]*PovTxEntry)
	return txPool
}

func (tp *PovTxPool) Init() {
	tp.recoverUnconfirmedTxs()
}

func (tp *PovTxPool) Start() {
	eb := tp.povEngine.GetEventBus()
	if eb != nil {
		eb.SubscribeSync(string(common.EventAddRelation), tp.onAddStateBlock)
		eb.SubscribeSync(string(common.EventDeleteRelation), tp.onDeleteStateBlock)
	}

	common.Go(tp.loop)
}

func (tp *PovTxPool) Stop() {
	close(tp.quitCh)
	tp.allTxs = nil
	tp.accountTxs = nil
}

func (tp *PovTxPool) onAddStateBlock(block *types.StateBlock) error {
	txHash := block.GetHash()
	//tp.povEngine.GetLogger().Debugf("recv event, add state block hash %s", txHash)
	tp.txEventCh <- &PovTxEvent{event: common.EventAddRelation, txHash: txHash, txBlock: block}
	return nil
}

func (tp *PovTxPool) onDeleteStateBlock(hash types.Hash) error {
	//tp.povEngine.GetLogger().Debugf("recv event, delete state block hash %s", hash)
	tp.txEventCh <- &PovTxEvent{event: common.EventDeleteRelation, txHash: hash}
	return nil
}

func (tp *PovTxPool) loop() {
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

func (tp *PovTxPool) recoverUnconfirmedTxs() {
	ledger := tp.povEngine.GetLedger()
	logger := tp.povEngine.GetLogger()

	stateBlockNum, _ := ledger.CountStateBlocks()
	uncheckedStateBlockNum, _ := ledger.CountUncheckedBlocks()

	logger.Infof("begin to scan all state blocks, block %d, unchecked %d", stateBlockNum, uncheckedStateBlockNum)

	startTime := time.Now()
	unpackStateBlocks := make(map[types.Hash]*types.StateBlock)
	err := ledger.BatchView(func(txn db.StoreTxn) error {
		err := ledger.GetStateBlocks(func(block *types.StateBlock) error {
			txHash := block.GetHash()
			if ledger.HasPovTxLookup(txHash, txn) {
				return nil
			}
			unpackStateBlocks[txHash] = block
			return nil
		}, txn)

		if err != nil {
			logger.Errorf("failed to get state blocks, err %s", err)
			return err
		}

		return nil
	})
	if err != nil {
		logger.Errorf("scan all state blocks failed")
	}

	usedTime := time.Since(startTime)
	unconfirmedStateBlockNum := len(unpackStateBlocks)

	logger.Infof("finished to scan all state blocks used time %s, unconfirmed %d", usedTime.String(), unconfirmedStateBlockNum)

	if len(unpackStateBlocks) <= 0 {
		return
	}

	txStepNum := len(unpackStateBlocks) / 100
	if txStepNum < 5000 {
		txStepNum = 5000
	}
	txAddNum := 0
	for txHash, block := range unpackStateBlocks {
		tp.addTx(txHash, block)
		txAddNum++
		if txAddNum%txStepNum == 0 {
			logger.Infof("total %d unconfirmed state blocks have been added to tx pool", txAddNum)
		}
	}

	logger.Infof("total %d unconfirmed state blocks have been added to tx pool", txAddNum)
}

func (tp *PovTxPool) processTxEvent(txEvent *PovTxEvent) {
	if txEvent.event == common.EventAddRelation {
		tp.addTx(txEvent.txHash, txEvent.txBlock)
	} else if txEvent.event == common.EventDeleteRelation {
		tp.delTx(txEvent.txHash)
	}
}

func (tp *PovTxPool) addTx(txHash types.Hash, txBlock *types.StateBlock) {
	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	txExist, ok := tp.allTxs[txHash]
	if ok && txExist != nil {
		return
	}

	tp.povEngine.GetLogger().Debugf("add tx %s", txHash)

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

	tp.povEngine.GetLogger().Debugf("delete tx %s", txHash)

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

func (tp *PovTxPool) SelectPendingTxs(stateTrie *trie.Trie, limit int) []*types.StateBlock {
	tp.txMu.RLock()
	defer tp.txMu.RUnlock()

	var retTxs []*types.StateBlock

	if limit <= 0 {
		return retTxs
	}

	//tp.povEngine.GetLogger().Debugf("select pending txs in pool, txs %d", len(tp.allTxs))

	addrTokenPrevHashes := make(map[types.AddressToken]types.Hash)
	for addrToken, accTxList := range tp.accountTxs {
		isCA := types.IsContractAddress(addrToken.Address)

		selectedTxHashes := make(map[types.Hash]struct{})
		for accTxList.Len() > len(selectedTxHashes) {
			notInOrderTxNum := 0
			inOrderTxNum := 0
			for e := accTxList.Front(); e != nil; e = e.Next() {
				txEntry := e.Value.(*PovTxEntry)
				txBlock := txEntry.txBlock
				txHash := txEntry.txHash
				if _, ok := selectedTxHashes[txHash]; ok {
					continue
				}
				token := addrToken.Token

				// contract address's blocks are all independent, no previous
				if isCA {
					addrTokenPrevHashes[addrToken] = types.ZeroHash
				} else {
					as := tp.povEngine.chain.GetAccountState(stateTrie, addrToken.Address)
					if as != nil {
						rs := as.GetTokenState(token)
						if rs != nil {
							addrTokenPrevHashes[addrToken] = rs.Hash
						} else {
							addrTokenPrevHashes[addrToken] = types.ZeroHash
						}
					} else {
						addrTokenPrevHashes[addrToken] = types.ZeroHash
					}
				}

				//tp.povEngine.GetLogger().Debugf("address %s token %s block %s", addrToken.Address, token, txHash)
				//tp.povEngine.GetLogger().Debugf("tokenPrevHash %s txPrevious %s", addrTokenPrevHashes[addrToken], txBlock.GetPrevious())
				if txBlock.GetPrevious() == addrTokenPrevHashes[addrToken] {
					retTxs = append(retTxs, txBlock)

					selectedTxHashes[txHash] = struct{}{}
					// contract address's blocks are all independent, no previous
					if !isCA {
						addrTokenPrevHashes[addrToken] = txHash
					}
					inOrderTxNum++

					limit--
					if limit == 0 {
						break
					}
				} else {
					notInOrderTxNum++
				}
			}
			if inOrderTxNum == 0 {
				if notInOrderTxNum > 0 {
					tp.povEngine.GetLogger().Debugf("address %s has txs %d not in order", addrToken.Address, notInOrderTxNum)
				}
				break
			}
			if limit == 0 {
				break
			}
		}

		if limit == 0 {
			break
		}
	}

	return retTxs
}

func (tp *PovTxPool) LastUpdated() time.Time {
	return time.Unix(tp.lastUpdated, 0)
}
