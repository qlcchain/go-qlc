package ledger

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/google/uuid"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
)

type Ledger struct {
	io.Closer
	Store          db.Store
	dir            string
	RollbackChan   chan types.Hash
	EB             event.EventBus
	representCache *RepresentationCache
	ctx            context.Context
	cancel         context.CancelFunc
	cache          *Cache
	logger         *zap.SugaredLogger
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
)

const (
	idPrefixBlock byte = iota
	idPrefixSmartContractBlock
	idPrefixUncheckedBlockPrevious
	idPrefixUncheckedBlockLink
	idPrefixAccount
	//idPrefixToken
	idPrefixFrontier
	idPrefixPending
	idPrefixRepresentation
	idPrefixPerformance
	idPrefixChild
	idPrefixVersion
	idPrefixStorage
	idPrefixToken    //discard
	idPrefixSender   //discard
	idPrefixReceiver //discard
	idPrefixMessage  //discard
	idPrefixMessageInfo
	idPrefixOnlineReps
	idPrefixPovHeader   // prefix + height + hash => header
	idPrefixPovBody     // prefix + height + hash => body
	idPrefixPovHeight   // prefix + hash => height (uint64)
	idPrefixPovTxLookup // prefix + txHash => TxLookup
	idPrefixPovBestHash // prefix + height => hash
	idPrefixPovTD       // prefix + height + hash => total difficulty (big int)
	idPrefixLink
	idPrefixBlockCache //block store this table before consensus complete
	idPrefixRepresentationCache
	idPrefixUncheckedTokenInfo
	idPrefixBlockCacheAccount
	idPrefixPovMinerStat // prefix + day index => miners of best blocks per day
	idPrefixUnconfirmedSync
	idPrefixUncheckedSync
	idPrefixSyncCacheBlock
	idPrefixUncheckedPovHeight
	idPrefixPovLatestHeight  // prefix => height
	idPrefixPovTxlScanCursor // prefix => height
)

var (
	cache = make(map[string]*Ledger)
	lock  = sync.RWMutex{}
)

const version = 10

func NewLedger(cfgFile string) *Ledger {
	lock.Lock()
	defer lock.Unlock()
	cc := chainctx.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	dir := cfg.LedgerDir()

	if _, ok := cache[dir]; !ok {
		store, err := db.NewBadgerStore(dir)
		if err != nil {
			fmt.Println(err.Error())
		}
		ctx, cancel := context.WithCancel(context.Background())
		l := &Ledger{
			Store:          store,
			dir:            dir,
			RollbackChan:   make(chan types.Hash, 100),
			EB:             cc.EventBus(),
			ctx:            ctx,
			cancel:         cancel,
			cache:          NewCache(),
			representCache: NewRepresentationCache(),
		}
		l.logger = log.NewLogger("ledger")

		if err := l.upgrade(); err != nil {
			l.logger.Error(err)
		}
		if err := l.initCache(); err != nil {
			l.logger.Error(err)
		}
		cache[dir] = l
	}
	//cache[dir].logger = log.NewLogger("ledger")
	return cache[dir]
}

//CloseLedger force release all ledger instance
func CloseLedger() {
	for k, v := range cache {
		if v != nil {
			v.Close()
			//logger.Debugf("release ledger from %s", k)
		}
		lock.Lock()
		delete(cache, k)
		lock.Unlock()
	}
}

func (l *Ledger) Close() error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[l.dir]; ok {
		err := l.Store.Close()
		l.cancel()
		l.logger.Info("badger closed")
		delete(cache, l.dir)
		return err
	}
	return nil
}

func (l *Ledger) upgrade() error {
	return l.BatchUpdate(func(txn db.StoreTxn) error {
		_, err := getVersion(txn)
		if err != nil {
			if err == ErrVersionNotFound {
				if err := setVersion(version, txn); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		ms := []db.Migration{new(MigrationV1ToV7), new(MigrationV7ToV8), new(MigrationV8ToV9), new(MigrationV9ToV10)}

		err = txn.Upgrade(ms)
		if err != nil {
			l.logger.Error(err)
		}
		return err
	})
}

func (l *Ledger) initCache() error {
	txn := l.Store.NewTransaction(true)
	defer func() {
		if err := txn.Commit(nil); err != nil {
			l.logger.Error(err)
		}
	}()
	err := l.representCache.cacheToConfirmed(txn)
	if err != nil {
		l.logger.Error(err)
		return err
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-l.ctx.Done():
				return
			case <-ticker.C:
				l.processCache()
			}
		}
	}()
	return nil
}

func (l *Ledger) processCache() {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[l.dir]; ok {
		txn := l.Store.NewTransaction(true)
		defer func() {
			if err := txn.Commit(nil); err != nil {
				l.logger.Error(err)
			}
		}()
		if err := l.representCache.memoryToConfirmed(txn); err != nil {
			l.logger.Errorf("cache to confirmed error : %s", err)
		}
	}
}

// Empty reports whether the database is empty or not.
func (l *Ledger) Empty(txns ...db.StoreTxn) (bool, error) {
	r := true
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		r = false
		return nil
	})
	if err != nil {
		return r, err
	}
	return r, nil
}

func (l *Ledger) DBStore() db.Store {
	return l.Store
}

func (l *Ledger) EventBus() event.EventBus {
	return l.EB
}

// BatchUpdate MUST pass the same txn
func (l *Ledger) BatchUpdate(fn func(txn db.StoreTxn) error) error {
	txn := l.Store.NewTransaction(true)
	//logger.Debugf("BatchUpdate NewTransaction %p", txn)
	defer func() {
		//logger.Debugf("BatchUpdate Discard %p", txn)
		txn.Discard()
	}()

	if err := fn(txn); err != nil {
		return err
	}
	return txn.Commit(nil)
}

// BatchView MUST pass the same txn
func (l *Ledger) BatchView(fn func(txn db.StoreTxn) error) error {
	txn := l.Store.NewTransaction(false)
	//logger.Debugf("BatchView NewTransaction %p", txn)
	defer func() {
		//logger.Debugf("BatchView Discard %p", txn)
		txn.Discard()
	}()

	if err := fn(txn); err != nil {
		return err
	}

	return txn.Commit(nil)
}

func (l *Ledger) BatchWrite(fn func(batch db.StoreBatch) error) error {
	batch := l.Store.NewWriteBatch()
	err := fn(batch)
	if err != nil {
		batch.Cancel()
		return err
	}
	return batch.Flush()
}

//getTxn get txn by `update` mode
func (l *Ledger) getTxn(update bool, txns ...db.StoreTxn) (db.StoreTxn, bool) {
	if len(txns) > 0 {
		//logger.Debugf("getTxn %p", txns[0])
		return txns[0], false
	} else {
		txn := l.Store.NewTransaction(update)
		//logger.Debugf("getTxn new %p", txn)
		return txn, true
	}
}

// releaseTxn commit change and close txn
func (l *Ledger) releaseTxn(txn db.StoreTxn, flag bool) {
	if flag {
		err := txn.Commit(nil)
		if err != nil {
			l.logger.Error(err)
		}
		txn.Discard()
	}
}

func getKeyOfParts(t byte, partList ...interface{}) ([]byte, error) {
	var buffer = []byte{t}
	for _, part := range partList {
		var src []byte
		switch part.(type) {
		case int:
			src = util.BE_Uint64ToBytes(uint64(part.(int)))
		case int32:
			src = util.BE_Uint64ToBytes(uint64(part.(int32)))
		case uint32:
			src = util.BE_Uint64ToBytes(uint64(part.(uint32)))
		case int64:
			src = util.BE_Uint64ToBytes(uint64(part.(int64)))
		case uint64:
			src = util.BE_Uint64ToBytes(part.(uint64))
		case []byte:
			src = part.([]byte)
		case types.Hash:
			hash := part.(types.Hash)
			src = hash[:]
		case *types.Hash:
			hash := part.(*types.Hash)
			src = hash[:]
		case types.Address:
			hash := part.(types.Address)
			src = hash[:]
		case *types.PendingKey:
			pk := part.(*types.PendingKey)
			var err error
			src, err = pk.Serialize()
			if err != nil {
				return nil, fmt.Errorf("pending key serialize: %s", err)
			}
		default:
			return nil, errors.New("key contains of invalid part")
		}

		buffer = append(buffer, src...)
	}

	return buffer, nil
}

func getVersionKey() []byte {
	return []byte{idPrefixVersion}
}

func setVersion(version int64, txn db.StoreTxn) error {
	key := getVersionKey()
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, version)
	return txn.Set(key, buf[:n])
}

func getVersion(txn db.StoreTxn) (int64, error) {
	var i int64
	key := getVersionKey()
	err := txn.Get(key, func(val []byte, b byte) error {
		i, _ = binary.Varint(val)
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return 0, ErrVersionNotFound
		}
		return i, err
	}
	return i, nil
}

func (l *Ledger) generateWork(hash types.Hash) types.Work {
	var work types.Work
	worker, _ := types.NewWorker(work, hash)
	return worker.NewWork()
	//
	////cache to Store
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
		return nil, err
	}
	if has {
		rxTm, err := l.GetTokenMeta(rxAccount, sendBlock.GetToken())
		if err != nil {
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
	tm := am.Token(common.ChainToken())
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
	tm := am.Token(common.ChainToken())
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

func (l *Ledger) Size() (int64, int64) {
	return l.Store.Size()
}

func (l *Ledger) GC() error {
	return l.Store.Purge()
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

type ActionType byte

const (
	Dump ActionType = iota
	GC
	BackUp
)

func (l *Ledger) Action(action ActionType) (string, error) {
	switch action {
	case Dump:
		return l.Dump()
	case GC:
		if err := l.GC(); err != nil {
			return "", err
		}
		return "", nil
	default:
		return "", nil
	}
}
