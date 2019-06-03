package consensus

import (
	"container/list"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"sync"
	"time"
)

type PovTxPool struct {
	povEngine   *PoVEngine
	txMu        sync.RWMutex
	accountTxs  map[types.Address]*list.List
	allTxs      map[types.Hash]*types.StateBlock
	lastUpdated int64
}

func NewPovTxPool(povImpl *PoVEngine) *PovTxPool {
	txPool := &PovTxPool{
		povEngine: povImpl,
	}
	txPool.accountTxs = make(map[types.Address]*list.List)
	txPool.allTxs = make(map[types.Hash]*types.StateBlock)
	return txPool
}

func (tp *PovTxPool) Init() {
	ledger := tp.povEngine.GetLedger()
	logger := tp.povEngine.GetLogger()

	startTime := time.Now()
	totalStateBlockNum := 0
	unpackStateBlockNum := 0

	err := ledger.BatchView(func(txn db.StoreTxn) error {
		err := ledger.GetStateBlocks(func(block *types.StateBlock) error {
			totalStateBlockNum++

			txHash := block.GetHash()
			if ledger.HasPovTxLookup(txHash, txn) {
				return nil
			}

			unpackStateBlockNum++
			tp.addTx(txHash, block)
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

	logger.Infof("scan all account blocks used time %s", usedTime.String())
	logger.Infof("blocks %d, unpacked %d", totalStateBlockNum, unpackStateBlockNum)
}

func (tp *PovTxPool) Start() {
	eb := tp.povEngine.GetEventBus()
	if eb != nil {
		eb.SubscribeSync(string(common.EventAddRelation), tp.onAddStateBlock)
		eb.SubscribeSync(string(common.EventDeleteRelation), tp.onDeleteStateBlock)
	}
}

func (tp *PovTxPool) Stop() {
	tp.allTxs = nil
	tp.accountTxs = nil
}

func (tp *PovTxPool) onAddStateBlock(tx *types.StateBlock) error {
	tp.povEngine.GetLogger().Debugf("recv event, add state block hash %s", tx.GetHash())

	txHash := tx.GetHash()
	tp.addTx(txHash, tx)
	return nil
}

func (tp *PovTxPool) onDeleteStateBlock(tx *types.StateBlock) error {
	tp.povEngine.GetLogger().Debugf("recv event, delete state block hash %s", tx.GetHash())

	txHash := tx.GetHash()
	tp.delTx(txHash)
	return nil
}

func (tp *PovTxPool) addTx(txHash types.Hash, tx *types.StateBlock) {
	tp.txMu.Lock()
	defer tp.txMu.Unlock()

	txExist, ok := tp.allTxs[txHash]
	if ok && txExist != nil {
		return
	}

	accTxList, ok := tp.accountTxs[tx.Address]
	if !ok {
		accTxList = list.New()
		tp.accountTxs[tx.Address] = accTxList
	}

	var childE *list.Element
	for itE := accTxList.Back(); itE != nil; itE = itE.Prev() {
		itTx := itE.Value.(*types.StateBlock)
		if itTx.GetPrevious() == txHash {
			childE = itE
			break
		}
	}

	if childE != nil {
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

func (tp *PovTxPool) SelectPendingTxs(limit int) []*types.StateBlock {
	tp.txMu.RLock()
	defer tp.txMu.RUnlock()

	var retTxs []*types.StateBlock

	if limit <= 0 {
		return retTxs
	}

	for _, accTxList := range tp.accountTxs {
		for e := accTxList.Front(); e != nil; e = e.Next() {
			retTxs = append(retTxs, e.Value.(*types.StateBlock))
			limit--
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
