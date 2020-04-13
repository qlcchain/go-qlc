package relation

import (
	"context"
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"reflect"
	"sync"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	chaincontext "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/ledger/relation/db"
	"github.com/qlcchain/go-qlc/log"
)

type Relation struct {
	db         *sqlx.DB
	eb         event.EventBus
	subscriber *event.ActorSubscriber
	dir        string
	deleteChan chan types.Table
	addChan    chan types.Table
	drive      string
	ctx        context.Context
	cancel     context.CancelFunc
	closedChan chan bool
	tables     map[string]schema
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
		store, err := db.NewDB(cfg)
		if err != nil {
			return nil, fmt.Errorf("open store fail: %s", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		relation := &Relation{
			db:         store,
			drive:      cfg.DB.Driver,
			eb:         cc.EventBus(),
			dir:        cfgFile,
			deleteChan: make(chan types.Table, 10240),
			addChan:    make(chan types.Table, 10240),
			ctx:        ctx,
			cancel:     cancel,
			closedChan: make(chan bool),
			tables:     make(map[string]schema),
			logger:     log.NewLogger("relation"),
		}
		tables := []types.Table{new(types.BlockHash)}
		for _, t := range tables {
			if err := relation.Register(t); err != nil {
				return nil, fmt.Errorf("store init fail: %s", err)
			}
		}
		go relation.process()
		cache[cfgFile] = relation
	}
	//cache[dir].logger = log.NewLogger("ledger")
	return cache[cfgFile], nil
}

func (r *Relation) Register(t types.Table) error {
	if _, ok := r.tables[t.TableID()]; ok {
		return fmt.Errorf("table %s areadly exist", t.TableID())
	}
	var s schema
	rt := reflect.TypeOf(t).Elem()
	var columns []string
	columnsMap := make(map[string]string)
	key := ""
	for i := 0; i < rt.NumField(); i++ {
		column := rt.Field(i).Tag.Get("db")
		if column != "" {
			columns = append(columns, column)
			columnsMap[column] = convertSchemaType(r.drive, rt.Field(i).Tag.Get("typ"))
		}
		if k := rt.Field(i).Tag.Get("key"); k != "" {
			key = column
		}
	}
	s.tableName = rt.Name()
	s.create = create(s.tableName, columnsMap, key)
	s.insert = insert(s.tableName, columns)
	r.tables[t.TableID()] = s
	//r.logger.Debug(s)

	if _, err := r.db.Exec(s.create); err != nil {
		r.logger.Errorf("exec error, sql: %s, err: %s", s.create, err.Error())
		return err
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

func (r *Relation) Add(obj []types.Table) {
	for _, o := range obj {
		r.addChan <- o
	}
}

func (r *Relation) Delete(obj types.Table) {
	r.deleteChan <- obj
}

func (r *Relation) batchAdd(txn *sqlx.Tx, objs []types.Table) error {
	//objsMap := make(map[string][]*BlockHash)
	//for _, obj := range objs {
	//	objsMap[obj.TableID()] = append(objsMap[obj.TableID()], obj.(*BlockHash))
	//}
	//for _, objs := range objsMap {
	//	tableId := objs[0].TableID()
	//	sql := r.tables[tableId].insert
	//	for _, obj := range objs{
	//		if _, err := txn.Exec(sql, obj); err != nil {
	//			return fmt.Errorf("txn add exec: %s (%s)", err, sql)
	//		}
	//	}
	//}
	for _, obj := range objs {
		sql := r.tables[obj.TableID()].insert
		if _, err := txn.NamedExec(sql, obj); err != nil {
			return fmt.Errorf("txn add exec: %s [%s]", err, sql)
		}
	}
	return nil
}

func (r *Relation) batchDelete(txn *sqlx.Tx, objs []types.Table) error {
	for _, obj := range objs {
		if _, err := txn.Exec(obj.DeleteKey()); err != nil {
			return fmt.Errorf("txn delete exec: %s", err)
		}
	}
	return nil
}

func (r *Relation) closed() {
	<-r.closedChan
}

func (r *Relation) process() {
	addObjs := make([]types.Table, 0)
	deleteObjs := make([]types.Table, 0)

	for {
		select {
		case <-r.ctx.Done():
			r.flush()
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
				return r.batchAdd(txn, addObjs)
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

			if err := r.BatchUpdate(func(txn *sqlx.Tx) error {
				return r.batchDelete(txn, deleteObjs)
			}); err != nil {
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

func (r *Relation) flush() {
	//add chan
	if len(r.addChan) > 0 {
		addObjs := make([]types.Table, 0)
		for b := range r.addChan {
			addObjs = append(addObjs, b)
			if len(r.addChan) == 0 {
				break
			}
		}
		objs := make([]types.Table, 0)
		for _, obj := range addObjs {
			objs = append(objs, obj)
			if len(objs) == batchMaxCount {
				if err := r.BatchUpdate(func(txn *sqlx.Tx) error {
					return r.batchAdd(txn, addObjs)
				}); err != nil {
					r.logger.Errorf("batch update add error: %s", err)
				}
				objs = objs[:0]
			}
		}
		if err := r.BatchUpdate(func(txn *sqlx.Tx) error {
			return r.batchAdd(txn, addObjs)
		}); err != nil {
			r.logger.Errorf("batch update add error: %s", err)
		}
	}

	// delete chan
	if len(r.deleteChan) > 0 {
		deleteObjs := make([]types.Table, 0)
		for b := range r.deleteChan {
			deleteObjs = append(deleteObjs, b)
			if len(r.deleteChan) == 0 {
				break
			}
		}
		objs := make([]types.Table, 0)
		for _, obj := range deleteObjs {
			objs = append(objs, obj)
			if len(objs) == batchMaxCount {
				if err := r.BatchUpdate(func(txn *sqlx.Tx) error {
					return r.batchDelete(txn, deleteObjs)
				}); err != nil {
					r.logger.Errorf("batch update delete error: %s", err)
				}
				objs = objs[:0]
			}
		}
		if err := r.BatchUpdate(func(txn *sqlx.Tx) error {
			return r.batchDelete(txn, deleteObjs)
		}); err != nil {
			r.logger.Errorf("batch update delete error: %s", err)
		}
	}
}

func (r *Relation) EmptyStore() error {
	r.logger.Info("empty store")
	for _, s := range r.tables {
		sql := fmt.Sprintf("delete from %s ", s.tableName)
		if _, err := r.db.Exec(sql); err != nil {
			return fmt.Errorf("exec delete error, sql: %s, err: %s", sql, err.Error())
		}
	}
	return nil
}

func (r *Relation) DB() *sqlx.DB {
	return r.db
}
