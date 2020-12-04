package ledger

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/util"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	typelation "github.com/qlcchain/go-qlc/common/relation"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/storage/db"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/ledger/migration"
	"github.com/qlcchain/go-qlc/ledger/relation"
	"github.com/qlcchain/go-qlc/log"
)

type LedgerStore interface {
	Close() error
	DBStore() storage.Store
	EventBus() event.EventBus
	Get(k []byte, c ...storage.Cache) ([]byte, error)
	GetObject(k []byte, c ...storage.Cache) (interface{}, []byte, error)
	Iterator([]byte, []byte, func([]byte, []byte) error) error
	IteratorObject(prefix []byte, end []byte, fn func([]byte, interface{}) error) error
	GenerateSendBlock(block *types.StateBlock, amount types.Balance, prk ed25519.PrivateKey) (*types.StateBlock, error)
	GenerateReceiveBlock(sendBlock *types.StateBlock, prk ed25519.PrivateKey) (*types.StateBlock, error)
	GenerateChangeBlock(account types.Address, representative types.Address, prk ed25519.PrivateKey) (*types.StateBlock, error)
	GenerateOnlineBlock(account types.Address, prk ed25519.PrivateKey, povHeight uint64) (*types.StateBlock, error)
	GetVerifiedData() map[types.Hash]int
	Action(at storage.ActionType, t int) (interface{}, error)
	GetRelation(dest interface{}, query string) error
	SelectRelation(dest interface{}, query string) error
	Flush() error
	FlushU() error
	BlockConfirmed(blk *types.StateBlock)
	AddTrieCleanHeight(height uint64) error
	GetTrieCleanHeight() (uint64, error)
	NeedToWriteTrie(height uint64) bool
}

type Ledger struct {
	io.Closer
	dir            string
	cfg            *config.Config
	store          storage.Store
	cache          *MemoryCache
	unCheckCache   *MemoryCache
	rcache         *rCache
	cacheStats     []*CacheStat
	uCacheStats    []*CacheStat
	relation       *relation.Relation
	EB             event.EventBus
	blockConfirmed chan *types.StateBlock
	ctx            context.Context
	cancel         context.CancelFunc
	verifiedData   map[types.Hash]int
	deletedSchema  []types.Schema
	logger         *zap.SugaredLogger
	tokenCache     sync.Map
}

var (
	ErrStoreEmpty             = errors.New("the store is empty")
	ErrBlockExists            = errors.New("block already exists")
	ErrBlockNotFound          = errors.New("block not found")
	ErrUncheckedBlockExists   = errors.New("unchecked block already exists")
	ErrUncheckedBlockNotFound = errors.New("unchecked block not found")
	ErrAccountExists          = errors.New("account already exists")
	ErrAccountNotFound        = errors.New("account not found")
	ErrTokenExists            = errors.New("token already exists")
	ErrTokenNotFound          = errors.New("token not found")
	//ErrTokenInfoExists        = errors.New("token info already exists")
	//ErrTokenInfoNotFound      = errors.New("token info not found")
	ErrPendingExists          = errors.New("pending transaction already exists")
	ErrPendingNotFound        = errors.New("pending transaction not found")
	ErrFrontierExists         = errors.New("frontier already exists")
	ErrFrontierNotFound       = errors.New("frontier not found")
	ErrRepresentationNotFound = errors.New("representation not found")
	ErrPerformanceNotFound    = errors.New("performance not found")
	//ErrChildExists            = errors.New("child already exists")
	//ErrChildNotFound          = errors.New("child not found")
	ErrVersionNotFound = errors.New("version not found")
	ErrLinkNotFound    = errors.New("link not found")
	ErrPeerExists      = errors.New("peer already exists")
	ErrPeerNotFound    = errors.New("peer not found")
)

var (
	lcache = make(map[string]*Ledger)
	lock   = sync.RWMutex{}
)

const version = 16

func NewLedger(cfgFile string) *Ledger {
	lock.Lock()
	defer lock.Unlock()
	cc := chainctx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	dir := cfg.LedgerDir()

	if _, ok := lcache[dir]; !ok {
		ctx, cancel := context.WithCancel(context.Background())
		l := &Ledger{
			dir:            dir,
			EB:             cc.EventBus(),
			ctx:            ctx,
			cancel:         cancel,
			blockConfirmed: make(chan *types.StateBlock, 20480),
			deletedSchema:  make([]types.Schema, 0),
			logger:         log.NewLogger("ledger"),
			tokenCache:     sync.Map{},
			cfg:            cfg,
		}
		store, err := db.NewBadgerStore(dir)
		if err != nil {
			l.logger.Fatal(err.Error())
		}
		l.store = store
		r, err := relation.NewRelation(cfgFile)
		if err != nil {
			l.logger.Fatal(err.Error())
		}
		l.relation = r
		l.cache = NewMemoryCache(l, defaultBlockFlushSecs, 2000, block)
		l.unCheckCache = NewMemoryCache(l, defaultUncheckFlushSecs, 50, unchecked)
		l.rcache = NewrCache()
		l.cacheStats = make([]*CacheStat, 0)
		l.uCacheStats = make([]*CacheStat, 0)
		if err := l.init(); err != nil {
			l.logger.Fatal(err)
		}
		lcache[dir] = l
	}
	//cache2[dir].logger = log.NewLogger("ledger")
	return lcache[dir]
}

func DefaultStore() Store {
	lock.Lock()
	defer lock.Unlock()
	dir := filepath.Join(config.DefaultDataDir(), "ledger")
	if l, ok := lcache[dir]; ok {
		return l
	}
	return nil
}

func (l *Ledger) init() error {
	vd, err := l.getVerifiedData()
	if err != nil {
		return fmt.Errorf("get verified data: %s ", err)
	}
	l.verifiedData = vd

	if err := l.upgrade(); err != nil {
		return fmt.Errorf("upgrade: %s ", err)
	}

	if err := l.initRelation(); err != nil {
		return fmt.Errorf("init relation: %s ", err)
	}

	if err := l.removeBlockConfirmed(); err != nil {
		return fmt.Errorf("remove block confirmed: %s ", err)
	}

	if err := l.updateRepresentation(); err != nil {
		return fmt.Errorf("update representation: %s ", err)
	}

	return nil
}

func (l *Ledger) SetCacheCapacity() error {
	if err := l.Flush(); err != nil {
		return err
	}
	l.cache.ResetCapacity(defaultBlockCapacity)
	l.unCheckCache.ResetCapacity(defaultUncheckCapacity)
	return nil
}

func (l *Ledger) removeBlockConfirmed() error {
	err := l.GetBlockCaches(func(block *types.StateBlock) error {
		if b, _ := l.HasStateBlockConfirmed(block.GetHash()); b {
			if err := l.DeleteBlockCache(block.GetHash()); err != nil {
				return fmt.Errorf("delete block cache error: %s", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("removeBlockConfirmed error : %s", err)
	}
	go func() {
		timer := time.NewTicker(3 * time.Second)
		blocks := make([]*types.StateBlock, 0)
		for {
			select {
			case <-l.ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				if len(l.blockConfirmed) > 0 {
					for b := range l.blockConfirmed {
						blocks = append(blocks, b)
						if len(blocks) >= 10000 {
							break
						}
						if len(l.blockConfirmed) == 0 {
							break
						}
					}
				}

				if len(blocks) > 0 {
					err := l.store.BatchWrite(false, func(batch storage.Batch) error {
						for _, blk := range blocks {
							if err := l.DeleteBlockCache(blk.GetHash(), batch); err != nil {
								l.logger.Error(err)
							}
						}
						return nil
					})
					if err != nil {
						l.logger.Errorf("batch delete block cache error: %s", err)
					}
					blocks = blocks[:0]
				}
			}
		}
	}()
	return nil
}

func (l *Ledger) getVerifiedData() (map[types.Hash]int, error) {
	data, err := hex.DecodeString(verifieddata)
	if err != nil {
		return nil, err
	}
	verifiedMap := make(map[types.Hash]int)
	if err := json.Unmarshal(data, &verifiedMap); err != nil {
		return nil, err
	}
	return verifiedMap, nil
}

func (l *Ledger) GetVerifiedData() map[types.Hash]int {
	return l.verifiedData
}

func (l *Ledger) upgrade() error {
	v, err := l.getVersion()
	if err != nil {
		if err == ErrVersionNotFound {
			return l.setVersion(version)
		} else {
			return err
		}
	} else {
		if v >= version {
			return nil
		}
		ms := []migration.Migration{
			new(migration.MigrationV1ToV15),
			new(migration.MigrationV15ToV16),
		}

		err = migration.Upgrade(ms, l.store)
		if err != nil {
			l.logger.Error(err)
		}
		return err
	}
}

func (l *Ledger) initRelation() error {
	count1, err := l.relation.BlocksCount()
	if err != nil {
		return fmt.Errorf("get relation block count, %s ", err)
	}

	count2, err := l.CountStateBlocks()
	if err != nil {
		return fmt.Errorf("get block count, %s ", err)
	}

	if count1 != count2 {
		if err := l.relation.EmptyStore(); err != nil {
			return fmt.Errorf("relation emptystore, %s ", err)
		}
		return l.GetStateBlocksConfirmed(func(block *types.StateBlock) error {
			c, err := block.ConvertToSchema()
			if err != nil {
				return fmt.Errorf("relation convert, %s ", err)
			}
			l.relation.Add(c)
			return nil
		})
	}
	return nil
}

//CloseLedger force release all ledger instance
func CloseLedger() {
	for k, v := range lcache {
		if v != nil {
			v.Close()
			//logger.Debugf("release ledger from %s", k)
		}
		lock.Lock()
		delete(lcache, k)
		lock.Unlock()
	}
}

func (l *Ledger) Close() error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := lcache[l.dir]; ok {
		l.cancel()
		l.cache.closed()
		l.unCheckCache.closed()
		if err := l.relation.Close(); err != nil {
			l.logger.Error(err)
		}
		if err := l.store.Close(); err != nil {
			return err
		}
		l.logger.Info("badger closed")
		delete(lcache, l.dir)
		return nil
	}
	return nil
}

func (l *Ledger) DBStore() storage.Store {
	return l.store
}

func (l *Ledger) EventBus() event.EventBus {
	return l.EB
}

func getVersionKey() []byte {
	return []byte{byte(storage.KeyPrefixVersion)}
}

func (l *Ledger) setVersion(version int64) error {
	key := getVersionKey()
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, version)
	return l.store.Put(key, buf[:n])
}

func (l *Ledger) getVersion() (int64, error) {
	var i int64
	key := getVersionKey()
	val, err := l.store.Get(key)
	if err != nil {
		if err == storage.KeyNotFound {
			return 0, ErrVersionNotFound
		}
		return i, err
	}
	i, _ = binary.Varint(val)
	return i, nil
}

func (l *Ledger) generateWork(hash types.Hash) types.Work {
	var work types.Work
	worker, _ := types.NewWorker(work, hash)
	return worker.NewWork()
	//
	////cache2 to DBStore
	//_ = s.setWork(hash, work)
}

func (l *Ledger) GenerateSendBlock(block *types.StateBlock, amount types.Balance, prk ed25519.PrivateKey) (*types.StateBlock, error) {
	tm, err := l.GetTokenMeta(block.GetAddress(), block.GetToken())
	if err != nil {
		return nil, errors.New("token not found")
	}
	prev, err := l.GetStateBlock(tm.Header)
	if err != nil {
		return nil, err
	}
	povHeader, err := l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}
	if tm.Balance.Compare(amount) != types.BalanceCompSmaller {
		block.Type = types.Send
		block.Balance = tm.Balance.Sub(amount)
		block.Previous = tm.Header
		block.Representative = tm.Representative
		block.Timestamp = common.TimeNow().Unix()
		block.Vote = prev.GetVote()
		block.Network = prev.GetNetwork()
		block.Oracle = prev.GetOracle()
		block.Storage = prev.GetStorage()
		block.PoVHeight = povHeader.GetHeight()

		if prk != nil {
			acc := types.NewAccount(prk)
			addr := acc.Address()
			if addr.String() != block.Address.String() {
				return nil, fmt.Errorf("send address (%s) is mismatch privateKey (%s)", block.Address.String(), acc.Address().String())
			}
			block.Signature = acc.Sign(block.GetHash())
			block.Work = l.generateWork(block.Root())
		}
		return block, nil
	} else {
		return nil, fmt.Errorf("not enought balance(%s) of %s", tm.Balance, amount)
	}
}

func (l *Ledger) GenerateReceiveBlock(sendBlock *types.StateBlock, prk ed25519.PrivateKey) (*types.StateBlock, error) {
	hash := sendBlock.GetHash()
	var sb types.StateBlock
	var acc *types.Account
	if !sendBlock.GetType().Equal(types.Send) {
		return nil, fmt.Errorf("(%s) is not send block", hash.String())
	}
	if exist, err := l.HasStateBlock(hash); !exist || err != nil {
		return nil, fmt.Errorf("send block(%s) does not exist", hash.String())
	}
	if prk != nil {
		acc = types.NewAccount(prk)
		addr := acc.Address()
		if addr.ToHash().String() != sendBlock.Link.String() {
			return nil, fmt.Errorf("receive address (%s) is mismatch privateKey (%s)", types.Address(sendBlock.Link).String(), acc.Address().String())
		}
	}
	rxAccount := types.Address(sendBlock.Link)
	info, err := l.GetPending(&types.PendingKey{Address: rxAccount, Hash: hash})
	if err != nil {
		return nil, err
	}
	has, err := l.HasTokenMeta(rxAccount, sendBlock.Token)
	if err != nil {
		l.logger.Errorf("token not found (%s), address: %s, token: %s", err, rxAccount.String(), sendBlock.Token.String())
		return nil, err
	}
	if has {
		rxTm, err := l.GetTokenMeta(rxAccount, sendBlock.GetToken())
		if err != nil {
			l.logger.Error(err)
			return nil, err
		}
		prev, err := l.GetStateBlock(rxTm.Header)
		if err != nil {
			return nil, err
		}
		if rxTm != nil {
			sb = types.StateBlock{
				Type:           types.Receive,
				Address:        rxAccount,
				Balance:        rxTm.Balance.Add(info.Amount),
				Vote:           prev.GetVote(),
				Oracle:         prev.GetOracle(),
				Network:        prev.GetNetwork(),
				Storage:        prev.GetStorage(),
				Previous:       rxTm.Header,
				Link:           hash,
				Representative: rxTm.Representative,
				Token:          rxTm.Type,
				Extra:          types.ZeroHash,
				Timestamp:      common.TimeNow().Unix(),
			}
		}
	} else {
		sb = types.StateBlock{
			Type:           types.Open,
			Address:        rxAccount,
			Balance:        info.Amount,
			Vote:           types.ZeroBalance,
			Oracle:         types.ZeroBalance,
			Network:        types.ZeroBalance,
			Storage:        types.ZeroBalance,
			Previous:       types.ZeroHash,
			Link:           hash,
			Representative: sendBlock.GetRepresentative(), //Representative: genesis.Owner,
			Token:          sendBlock.GetToken(),
			Extra:          types.ZeroHash,
			Timestamp:      common.TimeNow().Unix(),
		}
	}
	povHeader, err := l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}
	sb.PoVHeight = povHeader.GetHeight()

	if prk != nil && acc != nil {
		sb.Signature = acc.Sign(sb.GetHash())
		sb.Work = l.generateWork(sb.Root())
	}
	return &sb, nil
}

func (l *Ledger) GenerateChangeBlock(account types.Address, representative types.Address, prk ed25519.PrivateKey) (*types.StateBlock, error) {
	if _, err := l.GetAccountMeta(representative); err != nil {
		return nil, fmt.Errorf("invalid representative[%s]", representative.String())
	}

	am, err := l.GetAccountMeta(account)
	if err != nil {
		return nil, fmt.Errorf("account[%s] is not exist", account.String())
	}
	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("account[%s] has no chain token", account.String())
	}
	prev, err := l.GetStateBlock(tm.Header)
	if err != nil {
		return nil, fmt.Errorf("token header block not found")
	}
	povHeader, err := l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}
	sb := types.StateBlock{
		Type:           types.Change,
		Address:        account,
		Balance:        tm.Balance,
		Vote:           prev.GetVote(),
		Oracle:         prev.GetOracle(),
		Network:        prev.GetNetwork(),
		Storage:        prev.GetStorage(),
		Previous:       tm.Header,
		Link:           types.ZeroHash,
		PoVHeight:      povHeader.GetHeight(),
		Representative: representative,
		Token:          tm.Type,
		Extra:          types.ZeroHash,
		Timestamp:      common.TimeNow().Unix(),
	}
	if prk != nil {
		acc := types.NewAccount(prk)
		addr := acc.Address()
		if addr.String() != account.String() {
			return nil, fmt.Errorf("change address (%s) is mismatch privateKey (%s)", account.String(), acc.Address().String())
		}
		sb.Signature = acc.Sign(sb.GetHash())
		sb.Work = l.generateWork(sb.Root())
	}
	return &sb, nil
}

func (l *Ledger) GenerateOnlineBlock(account types.Address, prk ed25519.PrivateKey, povHeight uint64) (*types.StateBlock, error) {
	am, err := l.GetAccountMeta(account)
	if err != nil {
		return nil, fmt.Errorf("account[%s] is not exist", account.String())
	}
	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("account[%s] has no chain token", account.String())
	}
	prev, err := l.GetStateBlock(tm.Header)
	if err != nil {
		return nil, fmt.Errorf("token header block not found")
	}

	sb := types.StateBlock{
		Type:           types.Online,
		Address:        account,
		Balance:        tm.Balance,
		Vote:           prev.GetVote(),
		Oracle:         prev.GetOracle(),
		Network:        prev.GetNetwork(),
		Storage:        prev.GetStorage(),
		Previous:       tm.Header,
		Link:           types.ZeroHash,
		Representative: tm.Representative,
		Token:          tm.Type,
		Extra:          types.ZeroHash,
		Timestamp:      common.TimeNow().Unix(),
		PoVHeight:      povHeight,
	}
	if prk != nil {
		acc := types.NewAccount(prk)
		addr := acc.Address()
		if addr.String() != account.String() {
			return nil, fmt.Errorf("change address (%s) is mismatch privateKey (%s)", account.String(), acc.Address().String())
		}
		sb.Signature = acc.Sign(sb.GetHash())
		sb.Work = l.generateWork(sb.Root())
	}
	return &sb, nil
}

func (l *Ledger) Get(k []byte, c ...storage.Cache) ([]byte, error) {
	if len(c) > 0 && c[0] != nil {
		if r, err := c[0].Get(k); r != nil {
			return r.([]byte), nil
		} else {
			if err == ErrKeyDeleted {
				return nil, storage.KeyNotFound
			}
		}
	}

	if r, err := l.cache.Get(k); r != nil {
		return r.([]byte), nil
	} else {
		if err == ErrKeyDeleted {
			return nil, storage.KeyNotFound
		}
	}

	v, err := l.store.Get(k)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (l *Ledger) GetObject(k []byte, c ...storage.Cache) (interface{}, []byte, error) {
	if len(c) > 0 && c[0] != nil {
		if r, err := c[0].Get(k); r != nil {
			return r, nil, nil
		} else {
			if err == ErrKeyDeleted {
				return nil, nil, storage.KeyNotFound
			}
		}
	}

	if r, err := l.cache.Get(k); r != nil {
		return r, nil, nil
	} else {
		if err == ErrKeyDeleted {
			return nil, nil, storage.KeyNotFound
		}
	}

	v, err := l.store.Get(k)
	if err != nil {
		return nil, nil, err
	}
	return nil, v, nil
}

func (l *Ledger) Put(key []byte, value interface{}) error {
	c := l.cache.GetCache()
	return c.Put(key, value)
}

func (l *Ledger) Delete(k []byte) error {
	c := l.cache.GetCache()
	return c.Delete(k)
}

func (l *Ledger) getFromStore(key []byte, batch ...storage.Batch) ([]byte, error) {
	if len(batch) > 0 {
		v, err := batch[0].Get(key)
		if err != nil {
			return nil, err
		} else {
			return v.([]byte), nil
		}
	} else {
		return l.store.Get(key)
	}
}

func (l *Ledger) getFromCache(k []byte, c ...storage.Cache) (interface{}, error) {
	if len(c) > 0 && c[0] != nil {
		if r, err := c[0].Get(k); r != nil {
			return r, nil
		} else {
			if err == ErrKeyDeleted {
				return nil, ErrKeyDeleted
			}
		}
	}
	if r, err := l.cache.Get(k); r != nil {
		return r, nil
	} else {
		if err == ErrKeyDeleted {
			return nil, ErrKeyDeleted
		}
	}
	return nil, ErrKeyNotInCache
}

func (l *Ledger) Iterator(prefix []byte, end []byte, fn func(k []byte, v []byte) error) error {
	keys, err := l.cache.prefixIterator(prefix, fn)
	if err != nil {
		return fmt.Errorf("cache iterator : %s", err)
	}
	if err := l.DBStore().Iterator(prefix, end, func(k, v []byte) error {
		if !contain(keys, k) {
			if err := fn(k, v); err != nil {
				return fmt.Errorf("ledger iterator: %s", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("ledger store iterator: %s", err)
	}

	return nil
}

func (l *Ledger) IteratorObject(prefix []byte, end []byte, fn func(k []byte, v interface{}) error) error {
	keys, err := l.cache.prefixIteratorObject(prefix, fn)
	if err != nil {
		return fmt.Errorf("cache iterator : %s", err)
	}
	if err := l.DBStore().Iterator(prefix, end, func(k, v []byte) error {
		if !contain(keys, k) {
			if err := fn(k, v); err != nil {
				return fmt.Errorf("unchecked iterator: %s", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("unchecked store iterator: %s", err)
	}

	return nil
}

func NewTestLedger() (func(), *Ledger) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	l := NewLedger(cm.ConfigFile)

	return func() {
		_ = l.Close()
		_ = os.RemoveAll(dir)
	}, l
}

func (l *Ledger) Action(at storage.ActionType, t int) (interface{}, error) {
	switch at {
	case storage.Dump:
		return l.Dump(t)
	case storage.GC:
		return l.store.Action(at)
	case storage.Size:
		return l.store.Action(at)
	default:
		return "", errors.New("invalid action type")
	}
}

type CacheStat struct {
	Index  int
	Key    int
	Block  int
	Delete int
	Start  int64
	End    int64
}

func (l *Ledger) updateCacheStat(c *CacheStat, typ cacheType) {
	switch typ {
	case block:
		l.cacheStats = append(l.cacheStats, c)
		if len(l.cacheStats) > 200 {
			l.cacheStats = l.cacheStats[:100]
		}
	case unchecked:
		l.uCacheStats = append(l.uCacheStats, c)
		if len(l.uCacheStats) > 200 {
			l.uCacheStats = l.uCacheStats[:100]
		}
	}
}

func (l *Ledger) Flush() error {
	lock.Lock()
	defer lock.Unlock()
	return l.cache.rebuild()
}

func (l *Ledger) FlushU() error {
	lock.Lock()
	defer lock.Unlock()
	return l.unCheckCache.rebuild()
}

func (l *Ledger) BlockConfirmed(blk *types.StateBlock) {
	l.blockConfirmed <- blk
}

func (l *Ledger) RegisterRelation(objs []types.Schema) error {
	for _, obj := range objs {
		if err := l.relation.Register(obj); err != nil {
			return fmt.Errorf("relation register fail: %s", err)
		}
	}
	return nil
}

func (l *Ledger) RegisterInterface(con types.Convert, objs []types.Schema) error {
	if err := typelation.RegisterInterface(con); err != nil {
		return err
	}
	return l.RegisterRelation(objs)
}

func (l *Ledger) AddTrieCleanHeight(height uint64) error {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixTrieClean)
	if err != nil {
		return err
	}
	value := util.BE_Uint64ToBytes(height)
	return l.store.Put(key, value)
}

func (l *Ledger) GetTrieCleanHeight() (uint64, error) {
	key, err := storage.GetKeyOfParts(storage.KeyPrefixTrieClean)
	if err != nil {
		return 0, err
	}
	v, err := l.store.Get(key)
	if err != nil {
		return 0, err
	}
	return util.BE_BytesToUint64(v), nil
}

func (l *Ledger) NeedToWriteTrie(height uint64) bool {
	if !l.cfg.TrieClean.Enable {
		return true
	}
	return l.cfg.TrieClean.SyncWriteHeight < height
}
