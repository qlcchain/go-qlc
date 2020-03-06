package relation

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	chaincontext "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/storage/relationdb"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type Relation struct {
	db         relationdb.RelationDB
	eb         event.EventBus
	subscriber *event.ActorSubscriber
	dir        string
	deleteChan chan types.Schema
	addChan    chan types.Schema
	tables     []types.Schema
	ctx        context.Context
	cancel     context.CancelFunc
	closedChan chan bool
	logger     *zap.SugaredLogger
}

const batchMaxCount = 199

var (
	cache = make(map[string]*Relation)
	lock  = sync.RWMutex{}
)

//TODO ctx as a parameter from ledger
func NewRelation(cfgFile string) (*Relation, error) {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[cfgFile]; !ok {
		cc := chaincontext.NewChainContext(cfgFile)
		cfg, _ := cc.Config()
		store, err := relationdb.NewDB(cfg)
		if err != nil {
			return nil, fmt.Errorf("open store fail: %s", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		relation := &Relation{
			db:         store,
			eb:         cc.EventBus(),
			dir:        cfgFile,
			deleteChan: make(chan types.Schema, 10240),
			addChan:    make(chan types.Schema, 10240),
			ctx:        ctx,
			cancel:     cancel,
			closedChan: make(chan bool),
			logger:     log.NewLogger("relation"),
		}
		relation.tables = []types.Schema{new(types.StateBlock)}
		if err := relation.init(); err != nil {
			return nil, fmt.Errorf("store init fail: %s", err)
		}
		go relation.process()
		cache[cfgFile] = relation
	}
	//cache[dir].logger = log.NewLogger("ledger")
	return cache[cfgFile], nil
}

func (r *Relation) init() error {
	for _, s := range r.tables {
		fields, key := s.TableSchema()
		sql := r.db.CreateTable(s.TableName(), fields, key)
		if _, err := r.db.Store().Exec(sql); err != nil {
			r.logger.Errorf("exec error, sql: %s, err: %s", sql, err.Error())
			return err
		}
	}
	return nil
}

func (r *Relation) Close() error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[r.dir]; ok {
		r.cancel()
		r.closed()
		err := r.db.Close()
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
		vals := obj.SetRelation()
		s, i := r.db.Set(obj.TableName(), vals)
		//tx.MustExec(s, i...)
		if _, err := txn.Exec(s, i...); err != nil {
			return err
		}
	}
	return nil
}

func (r *Relation) batchDelete(txn *sqlx.Tx, objs []types.Schema) error {
	for _, obj := range objs {
		vals := obj.RemoveRelation()
		s := r.db.Delete(obj.TableName(), vals)
		//tx.MustExec(s)
		if _, err := txn.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *Relation) closed() {
	<-r.closedChan
}

func (r *Relation) process() {
	addObjs := make([]types.Schema, 0)
	deleteObjs := make([]types.Schema, 0)

	for {
		select {
		case <-r.ctx.Done():
			//TODO write all chan data
			r.closedChan <- true
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

func (r *Relation) EmptyStore() error {
	r.logger.Info("empty store")
	for _, s := range r.tables {
		sql := fmt.Sprintf("delete from %s ", s.TableName())
		if _, err := r.db.Store().Exec(sql); err != nil {
			r.logger.Errorf("exec delete error, sql: %s, err: %s", sql, err.Error())
			return err
		}
	}
	return nil
}
