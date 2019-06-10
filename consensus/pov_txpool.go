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

type PovTxPool struct {
	povEngine   *PoVEngine
	txMu        sync.RWMutex
	txEventCh   chan *PovTxEvent
	quitCh      chan struct{}
	accountTxs  map[types.Address]*list.List
	allTxs      map[types.Hash]*types.StateBlock
	lastUpdated int64
}

func NewPovTxPool(povImpl *PoVEngine) *PovTxPool {
	txPool := &PovTxPool{
		povEngine: povImpl,
	}
	txPool.txEventCh = make(chan *PovTxEvent, 5000)
	txPool.quitCh = make(chan struct{})
	txPool.accountTxs = make(map[types.Address]*list.List)
	txPool.allTxs = make(map[types.Hash]*types.StateBlock)
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
	for {
		select {
		case <-tp.quitCh:
			return
		case txEvent := <-tp.txEventCh:
			tp.processTxEvent(txEvent)
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
		logger.Errorf("scan all account state blocks failed")
	}

	usedTime := time.Since(startTime)
	unconfirmedStateBlockNum := len(unpackStateBlocks)

	logger.Infof("finished to scan all account blocks used time %s, unconfirmed %d", usedTime.String(), unconfirmedStateBlockNum)

	for txHash, block := range unpackStateBlocks {
		tp.addTx(txHash, block)
	}
}

func (tp *PovTxPool) processTxEvent(txEvent *PovTxEvent) {
	if txEvent.event == common.EventAddRelation {
		tp.addTx(txEvent.txHash, txEvent.txBlock)
	} else if txEvent.event == common.EventDeleteRelation {
		tp.delTx(txEvent.txHash)
	}
}

func (tp *PovTxPool) addTx(txHash types.Hash, tx *types.StateBlock) {
	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	txExist, ok := tp.allTxs[txHash]
	if ok && txExist != nil {
		return
	}

	tp.povEngine.GetLogger().Debugf("add tx %s", txHash)

	accTxList, ok := tp.accountTxs[tx.Address]
	if !ok {
		accTxList = list.New()
		tp.accountTxs[tx.Address] = accTxList
	}

	var childE *list.Element
	var parentE *list.Element

	prevHash := tx.GetPrevious()
	if prevHash.IsZero() {
		// contract address's blocks are all independent, no previous
		if types.IsContractAddress(tx.GetAddress()) {
		} else {
			// open block should be first
			childE = accTxList.Front()
		}
	} else {
		for itE := accTxList.Front(); itE != nil; itE = itE.Next() {
			itTx := itE.Value.(*types.StateBlock)
			if itTx.GetHash() == prevHash {
				parentE = itE
				break
			}
			if itTx.GetPrevious() == txHash {
				childE = itE
				break
			}
		}
	}

	if parentE != nil {
		accTxList.InsertAfter(tx, parentE)
	} else if childE != nil {
		accTxList.InsertBefore(tx, childE)
	} else {
		accTxList.PushBack(tx)
	}
	tp.allTxs[txHash] = tx

	tp.lastUpdated = time.Now().Unix()
}

func (tp *PovTxPool) delTx(txHash types.Hash) {
	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	tx, ok := tp.allTxs[txHash]
	if !ok {
		return
	}

	tp.povEngine.GetLogger().Debugf("delete tx %s", txHash)

	delete(tp.allTxs, txHash)

	if tx == nil {
		return
	}

	accTxList, ok := tp.accountTxs[tx.Address]
	if ok {
		var foundEle *list.Element
		for e := accTxList.Front(); e != nil; e = e.Next() {
			if e.Value == tx {
				foundEle = e
				break
			}
		}
		if foundEle != nil {
			accTxList.Remove(foundEle)
		}

		if accTxList.Len() <= 0 {
			delete(tp.accountTxs, tx.Address)
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

	addressPrevHashes := make(map[types.Address]types.Hash)
	for address, accTxList := range tp.accountTxs {
		isCA := types.IsContractAddress(address)
		// contract address's blocks are all independent, no previous
		if isCA {
			addressPrevHashes[address] = types.ZeroHash
		} else {
			as := tp.povEngine.chain.GetAccountState(stateTrie, address)
			if as != nil {
				addressPrevHashes[address] = as.Hash
			} else {
				addressPrevHashes[address] = types.ZeroHash
			}
		}

		selectedTxHashes := make(map[types.Hash]struct{})
		for accTxList.Len() > len(selectedTxHashes) {
			notInOrderTxNum := 0
			inOrderTxNum := 0
			for e := accTxList.Front(); e != nil; e = e.Next() {
				tx := e.Value.(*types.StateBlock)
				txHash := tx.GetHash()
				if _, ok := selectedTxHashes[txHash]; ok {
					continue
				}
				//tp.povEngine.GetLogger().Debugf("addressPrevHash %s", addressPrevHashes[address])
				//tp.povEngine.GetLogger().Debugf("address %s block %s previous %s", address, txHash, tx.GetPrevious())
				if tx.GetPrevious() == addressPrevHashes[address] {
					retTxs = append(retTxs, tx)

					selectedTxHashes[txHash] = struct{}{}
					// contract address's blocks are all independent, no previous
					if !isCA {
						addressPrevHashes[address] = txHash
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
					tp.povEngine.GetLogger().Debugf("address %s has txs %d not in order", address, notInOrderTxNum)
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
