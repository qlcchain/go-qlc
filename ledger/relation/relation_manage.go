package relation

import (
	"context"
	"sync"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	chaincontext "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/storage/relationdb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

type Relation struct {
	Store      *sqlx.DB
	eb         event.EventBus
	subscriber *event.ActorSubscriber
	dir        string
	deleteChan chan types.Schema
	addChan    chan types.Schema
	tables     []types.Schema
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *zap.SugaredLogger
}

const batchMaxCount = 199

var (
	cache = make(map[string]*Relation)
	lock  = sync.RWMutex{}
)

func NewRelation(cfgFile string) (*Relation, error) {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[cfgFile]; !ok {
		cc := chaincontext.NewChainContext(cfgFile)
		cfg, _ := cc.Config()
		store, err := relationdb.NewDB(cfg)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithCancel(context.Background())
		relation := &Relation{
			Store:      store,
			eb:         cc.EventBus(),
			dir:        cfgFile,
			deleteChan: make(chan types.Schema, 10240),
			addChan:    make(chan types.Schema, 10240),
			ctx:        ctx,
			cancel:     cancel,
			logger:     log.NewLogger("relation"),
		}
		relation.tables = []types.Schema{new(types.StateBlock)}
		if err := relation.init(); err != nil {
			return nil, err
		}
		go relation.process()
		cache[cfgFile] = relation
	}
	//cache[dir].logger = log.NewLogger("ledger")
	return cache[cfgFile], nil
}

func (r *Relation) Close() error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[r.dir]; ok {
		r.cancel()
		err := r.Store.Close()
		if err != nil {
			return err
		}
		r.logger.Info("sqlite closed")
		delete(cache, r.dir)
		return err
	}
	return nil
}

func (r *Relation) Add(obj types.Schema) {
	r.addChan <- obj
}

func (r *Relation) Delete(obj types.Schema) {
	r.deleteChan <- obj
}

func (r *Relation) batchAdd(txn *sqlx.Tx, objs []types.Schema) error {
	for _, obj := range objs {
		s, i := obj.SetRelation()
		//tx.MustExec(s, i...)
		if _, err := txn.Exec(s, i...); err != nil {
			return err
		}
	}
	return nil
}

func (r *Relation) batchDelete(txn *sqlx.Tx, objs []types.Schema) error {
	for _, obj := range objs {
		s := obj.RemoveRelation()
		//tx.MustExec(s)
		if _, err := txn.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *Relation) process() {
	addObjs := make([]types.Schema, 0)
	deleteObjs := make([]types.Schema, 0)

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("sqlite ctx done")
			return
		case obj := <-r.addChan:
			addObjs = append(addObjs, obj)
			if len(r.addChan) > 0 {
				for b := range r.addChan {
					addObjs = append(addObjs, b)
					if len(addObjs) >= batchMaxCount {
						break
					}
					if len(r.addChan) == 0 {
						break
					}
				}
			}

			err := r.BatchUpdate(func(txn *sqlx.Tx) error {
				if err := r.batchAdd(txn, addObjs); err != nil {
					r.logger.Errorf("batch add error: %s", err)
				}
				return nil
			})
			if err != nil {
				r.logger.Errorf("batch update obj error: %s", err)
			}
			addObjs = addObjs[:0]
		case obj := <-r.deleteChan:
			deleteObjs = append(deleteObjs, obj)
			if len(r.deleteChan) > 0 {
				for b := range r.deleteChan {
					deleteObjs = append(deleteObjs, b)
					if len(deleteObjs) >= batchMaxCount {
						break
					}
					if len(r.deleteChan) == 0 {
						break
					}
				}
			}

			err := r.BatchUpdate(func(txn *sqlx.Tx) error {
				if err := r.batchDelete(txn, deleteObjs); err != nil {
					r.logger.Errorf("batch delete error: %s", err)
				}
				return nil
			})
			if err != nil {
				r.logger.Errorf("batch delete objs error: %s", err)
			}
			deleteObjs = deleteObjs[:0]
			//case blk := <-r.syncBlkChan:
			//	r.syncBlocks = append(r.syncBlocks, blk)
			//	if len(r.syncBlocks) == batchMaxCount {
			//		r.batchUpdateBlocks()
			//		r.syncBlocks = r.syncBlocks[:0]
			//	}
			//case <-r.syncDone:
			//	if len(r.syncBlocks) > 0 {
			//		r.batchUpdateBlocks()
			//		r.syncBlocks = r.syncBlocks[:0]
			//	}
			//}
		}
	}
}
