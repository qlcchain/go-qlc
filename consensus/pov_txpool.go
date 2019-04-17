package consensus

import (
	"container/list"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"sync"
	"time"
)

type PovTxPool struct {
	povEngine *PoVEngine
	txMu   sync.RWMutex
	accountTxs map[types.Address]*list.List
	allTxs map[types.Hash]*types.StateBlock
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
	err := ledger.GetStateBlocks(func(tx *types.StateBlock) error {
		txHash := tx.GetHash()

		if ledger.HasPovTxLookup(txHash) {
			return nil
		}

		logger.Infof("account block %s not in pov block", txHash)

		tp.addTx(txHash, tx)
		return nil
	})
	usedTime := time.Since(startTime)
	logger.Infof("scan all account blocks used time %s", usedTime.String())

	if err != nil {
		logger.Errorf("scan all account blocks failed")
	}
}

func (tp *PovTxPool) Start() {
	eb := tp.povEngine.GetEventBus()
	if eb != nil {
		eb.Subscribe(string(common.EventAddRelation), tp.onAddStateBlock)
		eb.Subscribe(string(common.EventDeleteRelation), tp.onDeleteStateBlock)
	}
}

func (tp *PovTxPool) Stop() {
	tp.allTxs = nil
	tp.accountTxs = nil
}

func (tp *PovTxPool) onAddStateBlock(tx *types.StateBlock) error {
	tp.povEngine.GetLogger().Infof("recv event, add state block hash %s", tx.GetHash())

	txHash := tx.GetHash()
	tp.addTx(txHash, tx)
	return nil
}

func (tp *PovTxPool) onDeleteStateBlock(tx *types.StateBlock) error {
	tp.povEngine.GetLogger().Infof("recv event, delete state block hash %s", tx.GetHash())

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

	accTxList.PushBack(tx)
	tp.allTxs[txHash] = tx
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
}

func (tp *PovTxPool) SelectPendingTxs(limit int) []*types.Block {
	tp.txMu.RLock()
	defer tp.txMu.RUnlock()

	var retTxs []*types.Block

	if limit <= 0 {
		return retTxs
	}

	for _, accTxList := range tp.accountTxs {
		for e := accTxList.Front(); e != nil; e = e.Next() {
			retTxs = append(retTxs, e.Value.(*types.Block))
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
