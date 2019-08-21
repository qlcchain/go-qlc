package pov

import (
	"container/list"
	"sync"
	"time"

	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/trie"
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
	ledger      ledger.Store
	chain       PovTxChainReader
	txMu        sync.RWMutex
	txEventCh   chan *PovTxEvent
	quitCh      chan struct{}
	accountTxs  map[types.AddressToken]*list.List
	allTxs      map[types.Hash]*PovTxEntry
	lastUpdated int64
}

type PovTxChainReader interface {
	RegisterListener(listener EventListener)
	UnRegisterListener(listener EventListener)
	GetAccountState(trie *trie.Trie, address types.Address) *types.PovAccountState
}

func NewPovTxPool(eb event.EventBus, ledger ledger.Store, chain PovTxChainReader) *PovTxPool {
	txPool := &PovTxPool{
		logger: log.NewLogger("pov_txpool"),
		eb:     eb,
		ledger: ledger,
		chain:  chain,
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
	if tp.eb != nil {
		tp.eb.SubscribeSync(common.EventAddRelation, tp.onAddStateBlock)
		tp.eb.SubscribeSync(common.EventDeleteRelation, tp.onDeleteStateBlock)
	}

	tp.chain.RegisterListener(tp)

	common.Go(tp.loop)
}

func (tp *PovTxPool) Stop() {
	tp.chain.UnRegisterListener(tp)

	close(tp.quitCh)
}

func (tp *PovTxPool) onAddStateBlock(block *types.StateBlock) error {
	txHash := block.GetHash()
	//tp.logger.Debugf("recv event, add state block hash %s", txHash)
	tp.txEventCh <- &PovTxEvent{event: common.EventAddRelation, txHash: txHash, txBlock: block}
	return nil
}

func (tp *PovTxPool) onDeleteStateBlock(hash types.Hash) error {
	//tp.logger.Debugf("recv event, delete state block hash %s", hash)
	tp.txEventCh <- &PovTxEvent{event: common.EventDeleteRelation, txHash: hash}
	return nil
}

func (tp *PovTxPool) OnPovBlockEvent(event byte, block *types.PovBlock) {
	txs := block.GetAccountTxs()
	if len(txs) <= 0 {
		return
	}

	if event == EventConnectPovBlock {
		for _, tx := range txs {
			tp.delTx(tx.Hash)
		}
	} else if event == EventDisconnectPovBlock {
		for _, tx := range txs {
			tp.addTx(tx.Hash, tx.Block)
		}
	}
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
				for prevHash.IsZero() == false {
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
					//logger.Debugf("AddrToken %s block %s", addrToken, txHash)
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
			if types.IsContractAddress(block.GetAddress()) != true {
				continue
			}

			txHash := block.GetHash()
			if tp.ledger.HasPovTxLookup(txHash, txn) {
				continue
			}

			addrToken := types.AddressToken{Address: block.GetAddress(), Token: block.GetToken()}
			//logger.Debugf("AddrToken %s block %s", addrToken, txHash)
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

func (tp *PovTxPool) getUnconfirmedTxsBySlow() (map[types.AddressToken][]*PovTxEntry, int) {
	stateBlockNum, _ := tp.ledger.CountStateBlocks()
	uncheckedStateBlockNum, _ := tp.ledger.CountUncheckedBlocks()

	tp.logger.Infof("begin to scan all state blocks, block %d, unchecked %d", stateBlockNum, uncheckedStateBlockNum)

	accountTxs := make(map[types.AddressToken][]*PovTxEntry, 1000)
	unconfirmedTxNum := 0

	startTime := time.Now()

	err := tp.ledger.BatchView(func(txn db.StoreTxn) error {
		err := tp.ledger.GetStateBlocks(func(block *types.StateBlock) error {
			txHash := block.GetHash()
			if tp.ledger.HasPovTxLookup(txHash, txn) {
				return nil
			}

			addrToken := types.AddressToken{Address: block.GetAddress(), Token: block.GetToken()}
			accountTxs[addrToken] = append(accountTxs[addrToken], &PovTxEntry{txHash: txHash, txBlock: block})
			unconfirmedTxNum++
			return nil
		}, txn)

		if err != nil {
			tp.logger.Errorf("failed to get state blocks, err %s", err)
			return err
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

	addrToken := types.AddressToken{Address: txBlock.GetAddress(), Token: txBlock.GetToken()}

	tp.logger.Debugf("add tx %s", txHash)

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

	tp.logger.Debugf("delete tx %s", txHash)

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
				prevHashWant, ok := addrTokenPrevHashes[addrToken]
				if !ok {
					if isCA {
						prevHashWant = types.ZeroHash
					} else {
						as := tp.chain.GetAccountState(stateTrie, addrToken.Address)
						if as != nil {
							rs := as.GetTokenState(token)
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

				//tp.povEngine.GetLogger().Debugf("AddrToken %s block %s", addrToken, txHash)
				//tp.povEngine.GetLogger().Debugf("prevHashWant %s txPrevious %s", prevHashWant, txBlock.GetPrevious())
				if txBlock.GetPrevious() == prevHashWant {
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
					tp.logger.Debugf("AddrToken %s has txs %d not in order", addrToken, notInOrderTxNum)
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
