package ledger

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type Ledger struct {
	io.Closer
	Store          db.Store
	dir            string
	EB             event.EventBus
	representCache *RepresentationCache
	ctx            context.Context
	cancel         context.CancelFunc
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
)

var (
	cache = make(map[string]*Ledger)
	lock  = sync.RWMutex{}
)

const version = 7

func NewLedger(dir string) *Ledger {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[dir]; !ok {
		store, err := db.NewBadgerStore(dir)
		if err != nil {
			fmt.Println(err.Error())
		}
		ctx, cancel := context.WithCancel(context.Background())
		l := &Ledger{
			Store:          store,
			dir:            dir,
			EB:             event.GetEventBus(dir),
			ctx:            ctx,
			cancel:         cancel,
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
		ms := []db.Migration{new(MigrationV1ToV7)}

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
		ticker := time.NewTicker(15 * time.Second)
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
		if err := l.representCache.cacheToConfirmed(txn); err != nil {
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
			return nil, errors.New("Key contains of invalid part.")
		}

		buffer = append(buffer, src...)
	}

	return buffer, nil
}

func (l *Ledger) AddStateBlock(value *types.StateBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)

	k, err := getKeyOfParts(idPrefixBlock, value.GetHash())
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	if err := txn.Set(k, v); err != nil {
		return err
	}

	if b, err := l.HasBlockCache(value.GetHash()); b && err == nil {
		if err := l.DeleteBlockCache(value.GetHash(), txn); err != nil {
			return fmt.Errorf("delete block cache error: %s", err)
		}
	}
	if err := addChild(value, txn); err != nil {
		return fmt.Errorf("add block child error: %s", err)
	}
	if err := addLink(value, txn); err != nil {
		return fmt.Errorf("add block link error: %s", err)
	}
	l.releaseTxn(txn, flag)
	l.logger.Debug("publish addRelation,", value.GetHash())
	l.EB.Publish(common.EventAddRelation, value)
	return nil
}

func addChild(cBlock *types.StateBlock, txn db.StoreTxn) error {
	pHash := cBlock.Parent()
	cHash := cBlock.GetHash()
	if !common.IsGenesisBlock(cBlock) && pHash != types.ZeroHash && !cBlock.IsOpen() {
		// is parent block existed
		kCache, err := getKeyOfParts(idPrefixBlockCache, pHash)
		if err != nil {
			return err
		}
		err = txn.Get(kCache, func(v []byte, b byte) error {
			return nil
		})
		if err != nil {
			kConfirmed, err := getKeyOfParts(idPrefixBlock, pHash)
			if err != nil {
				return err
			}
			err = txn.Get(kConfirmed, func(v []byte, b byte) error {
				return nil
			})
			if err != nil {
				return fmt.Errorf("%s can not find parent %s", cHash.String(), pHash.String())
			}
		}

		// is parent have used
		k, err := getKeyOfParts(idPrefixChild, pHash)
		if err != nil {
			return err
		}

		err = txn.Get(k, func(val []byte, b byte) error {
			return nil
		})
		if err == nil {
			return fmt.Errorf("%s already have child ", pHash.String())
		}

		// add new relationship
		v, err := cHash.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := txn.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func addLink(block *types.StateBlock, txn db.StoreTxn) error {
	if block.GetType() == types.Open || block.GetType() == types.Receive || block.GetType() == types.ContractReward {
		k, err := getKeyOfParts(idPrefixLink, block.GetLink())
		if err != nil {
			return err
		}
		h := block.GetHash()
		v, err := h.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := txn.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) GetStateBlock(key types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	if blkCache, err := l.GetBlockCache(key); err == nil {
		return blkCache, nil
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return nil, err
	}

	value := new(types.StateBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetStateBlockConfirmed(key types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return nil, err
	}

	value := new(types.StateBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetStateBlocks(fn func(*types.StateBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			l.logger.Errorf("deserialize block error: %s", err)
			errStr = append(errStr, err.Error())
			return nil
		}
		if err := fn(blk); err != nil {
			l.logger.Errorf("process block error: %s", err)
			errStr = append(errStr, err.Error())
		}
		return nil
	})

	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) DeleteStateBlock(key types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return err
	}

	blk := new(types.StateBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := blk.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := txn.Delete(k); err != nil {
		return err
	}

	if err := l.deleteChild(blk, txn); err != nil {
		return fmt.Errorf("delete child error: %s", err)
	}
	if err := l.deleteLink(blk, txn); err != nil {
		return fmt.Errorf("delete link error: %s", err)
	}

	l.releaseTxn(txn, flag)
	l.logger.Info("publish deleteRelation,", key.String())
	l.EB.Publish(common.EventDeleteRelation, key)
	return nil
}

func (l *Ledger) deleteChild(blk *types.StateBlock, txn db.StoreTxn) error {
	pHash := blk.Parent()
	if !pHash.IsZero() {

		k, err := getKeyOfParts(idPrefixChild, pHash)
		if err != nil {
			return err
		}
		if err := txn.Delete(k); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) deleteLink(blk *types.StateBlock, txn db.StoreTxn) error {
	k, err := getKeyOfParts(idPrefixLink, blk.GetLink())
	if err != nil {
		return err
	}
	if err := txn.Delete(k); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) HasStateBlock(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	if exit, err := l.HasBlockCache(key); err == nil && exit {
		return exit, nil
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) HasStateBlockConfirmed(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlock, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) CountStateBlocks(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Count([]byte{idPrefixBlock})
}

func (l *Ledger) GetRandomStateBlock(txns ...db.StoreTxn) (*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	c, err := l.CountStateBlocks()
	if err != nil {
		return nil, err
	}
	if c == 0 {
		return nil, ErrStoreEmpty
	}
	blk := new(types.StateBlock)
	errFound := errors.New("state block found")

	for i := 0; i < 3; i++ {
		index := rand.Int63n(int64(c))
		var temp int64
		err = txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
			if temp == index {
				var b = new(types.StateBlock)
				if err = b.Deserialize(val); err != nil {
					return err
				}
				if !common.IsGenesisBlock(b) {
					blk = b
					return errFound
				}
			}
			temp++
			return nil
		})
		if err != nil && err != errFound {
			return nil, err
		}
		if !blk.Token.IsZero() {
			break
		}
	}
	if blk.Token.IsZero() {
		return nil, errors.New("state block not found")
	}
	return blk, nil
}

func (l *Ledger) AddSmartContractBlock(value *types.SmartContractBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixSmartContractBlock, value.GetHash())
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetSmartContractBlock(key types.Hash, txns ...db.StoreTxn) (*types.SmartContractBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixSmartContractBlock, key)
	if err != nil {
		return nil, err
	}

	value := new(types.SmartContractBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetSmartContractBlocks(fn func(block *types.SmartContractBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixSmartContractBlock, func(key []byte, val []byte, b byte) error {
		blk := new(types.SmartContractBlock)
		if err := blk.Deserialize(val); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		if err := fn(blk); err != nil {
			errStr = append(errStr, err.Error())
		}
		return nil
	})

	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) HasSmartContractBlock(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixSmartContractBlock, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) CountSmartContractBlocks(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	return txn.Count([]byte{idPrefixSmartContractBlock})
}

func (l *Ledger) uncheckedKindToPrefix(kind types.UncheckedKind) byte {
	switch kind {
	case types.UncheckedKindPrevious:
		return idPrefixUncheckedBlockPrevious
	case types.UncheckedKindLink:
		return idPrefixUncheckedBlockLink
	case types.UncheckedKindTokenInfo:
		return idPrefixUncheckedTokenInfo
	default:
		panic("bad unchecked block kind")
	}
}

func (l *Ledger) AddUncheckedBlock(key types.Hash, value *types.StateBlock, kind types.UncheckedKind, sync types.SynchronizedKind, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(l.uncheckedKindToPrefix(kind), key)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrUncheckedBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.SetWithMeta(k, v, byte(sync))
}

func (l *Ledger) GetUncheckedBlock(key types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (*types.StateBlock, types.SynchronizedKind, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(l.uncheckedKindToPrefix(kind), key)
	if err != nil {
		return nil, 0, err
	}

	value := new(types.StateBlock)
	var sync types.SynchronizedKind
	err = txn.Get(k, func(val []byte, b byte) (err error) {
		if err = value.Deserialize(val); err != nil {
			return err
		}
		sync = types.SynchronizedKind(b)
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, 0, ErrUncheckedBlockNotFound
		}
		return nil, 0, err
	}
	return value, sync, nil
}

func (l *Ledger) DeleteUncheckedBlock(key types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(l.uncheckedKindToPrefix(kind), key)
	if err != nil {
		return err
	}
	return txn.Delete(k)
}

func (l *Ledger) HasUncheckedBlock(key types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(l.uncheckedKindToPrefix(kind), key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) walkUncheckedBlocks(kind types.UncheckedKind, visit types.UncheckedBlockWalkFunc, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	prefix := l.uncheckedKindToPrefix(kind)
	err := txn.Iterator(prefix, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		h, err := types.BytesToHash(key[1:])
		if err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		if err := visit(blk, h, kind, types.SynchronizedKind(b)); err != nil {
			l.logger.Error("visit error %s", err)
			errStr = append(errStr, err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc, txns ...db.StoreTxn) error {
	if err := l.walkUncheckedBlocks(types.UncheckedKindPrevious, visit, txns...); err != nil {
		return err
	}

	if err := l.walkUncheckedBlocks(types.UncheckedKindLink, visit, txns...); err != nil {
		return err
	}

	return l.walkUncheckedBlocks(types.UncheckedKindTokenInfo, visit, txns...)
}

func (l *Ledger) CountUncheckedBlocks(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	var count uint64
	count, err := txn.Count([]byte{idPrefixUncheckedBlockLink})
	if err != nil {
		return 0, err
	}

	count2, err := txn.Count([]byte{idPrefixUncheckedBlockPrevious})
	if err != nil {
		return 0, err
	}

	return count + count2, nil
}

func (l *Ledger) AddAccountMeta(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrAccountExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetAccountMeta(key types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	am, err := l.GetAccountMetaCache(key)
	if err != nil {
		am = nil
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, key)
	if err != nil {
		return nil, err
	}
	meta := new(types.AccountMeta)
	err = txn.Get(k, func(v []byte, b byte) (err error) {
		if err = meta.Deserialize(v); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		meta = nil
	}
	if am != nil && meta == nil {
		return am, nil
	}
	if am == nil && meta != nil {
		return meta, nil
	}
	if am != nil && meta != nil {
		for _, v := range meta.Tokens {
			temp := am.Token(v.Type)
			if temp != nil {
				if temp.BlockCount < v.BlockCount {
					temp.BlockCount = v.BlockCount
					temp.Type = v.Type
					temp.BelongTo = v.BelongTo
					temp.Header = v.Header
					temp.Balance = v.Balance
					temp.Modified = v.Modified
					temp.OpenBlock = v.OpenBlock
					temp.Representative = v.Representative
				}
			} else {
				am.Tokens = append(am.Tokens, v)
			}
		}
		return am, nil
	}
	return nil, ErrAccountNotFound
}

func (l *Ledger) GetAccountMetaConfirmed(key types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, key)
	if err != nil {
		return nil, err
	}

	value := new(types.AccountMeta)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetAccountMetas(fn func(am *types.AccountMeta) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixAccount, func(key []byte, val []byte, b byte) error {
		am := new(types.AccountMeta)
		if err := am.Deserialize(val); err != nil {
			return err
		}
		if err := fn(am); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) CountAccountMetas(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Count([]byte{idPrefixAccount})
}

func (l *Ledger) UpdateAccountMeta(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrAccountNotFound
		}
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) AddOrUpdateAccountMeta(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) DeleteAccountMeta(key types.Address, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, key)
	if err != nil {
		return err
	}

	return txn.Delete(k)
}

func (l *Ledger) HasAccountMeta(key types.Address, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixAccount, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) AddTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		return err
	}

	if am.Token(meta.Type) != nil {
		return ErrTokenExists
	}

	am.Tokens = append(am.Tokens, meta)
	return l.UpdateAccountMeta(am, txns...)
}

func (l *Ledger) GetTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) (*types.TokenMeta, error) {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		return nil, err
	}

	tm := am.Token(tokenType)
	if tm == nil {
		return nil, ErrTokenNotFound
	}

	return tm, nil
}

func (l *Ledger) GetTokenMetaConfirmed(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) (*types.TokenMeta, error) {
	am, err := l.GetAccountMetaConfirmed(address, txns...)
	if err != nil {
		return nil, err
	}

	tm := am.Token(tokenType)
	if tm == nil {
		return nil, ErrTokenNotFound
	}

	return tm, nil
}

func (l *Ledger) UpdateTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		return err
	}

	//tm := am.Token(meta.Type)
	//
	//if tm != nil {
	//	tm = meta
	//	return l.UpdateAccountMeta(am, txns...)
	//}
	//tokens := am.Tokens
	for index, token := range am.Tokens {
		if token.Type == meta.Type {
			//am.Tokens = append(tokens[:index], tokens[index+1:]...)
			//am.Tokens = append(am.Tokens, meta)
			am.Tokens[index] = meta
			return l.UpdateAccountMeta(am, txns...)
		}
	}
	return ErrTokenNotFound
}

func (l *Ledger) AddOrUpdateTokenMeta(address types.Address, meta *types.TokenMeta, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range am.Tokens {
		if token.Type == meta.Type {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
			am.Tokens = append(am.Tokens, meta)
			return l.UpdateAccountMeta(am, txns...)
		}
	}

	am.Tokens = append(am.Tokens, meta)
	return l.UpdateAccountMeta(am, txns...)
}

func (l *Ledger) DeleteTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMetaConfirmed(address, txns...)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range tokens {
		if token.Type == tokenType {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}
	return l.UpdateAccountMeta(am, txns...)
}

func (l *Ledger) DeleteTokenMetaCache(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) error {
	am, err := l.GetAccountMetaCache(address, txns...)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range tokens {
		if token.Type == tokenType {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}
	return l.UpdateAccountMetaCache(am, txns...)
}

func (l *Ledger) HasTokenMeta(address types.Address, tokenType types.Hash, txns ...db.StoreTxn) (bool, error) {
	am, err := l.GetAccountMeta(address, txns...)
	if err != nil {
		if err == ErrAccountNotFound {
			return false, nil
		}
		return false, err
	}
	for _, t := range am.Tokens {
		if t.Type == tokenType {
			return true, nil
		}
	}
	return false, nil
}

func (l *Ledger) AddRepresentation(key types.Address, diff *types.Benefit, txns ...db.StoreTxn) error {
	spin := l.representCache.getLock(key.String())
	spin.Lock()
	defer spin.Unlock()

	value, err := l.GetRepresentation(key, txns...)
	if err != nil && err != ErrRepresentationNotFound {
		l.logger.Errorf("getRepresentation error: %s ,address: %s", err, key)
		return err
	}

	value.Balance = value.Balance.Add(diff.Balance)
	value.Vote = value.Vote.Add(diff.Vote)
	value.Network = value.Network.Add(diff.Network)
	value.Oracle = value.Oracle.Add(diff.Oracle)
	value.Storage = value.Storage.Add(diff.Storage)
	value.Total = value.Total.Add(diff.Total)

	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)
	return l.representCache.updateMemory(key, value, txn)
}

func (l *Ledger) SubRepresentation(key types.Address, diff *types.Benefit, txns ...db.StoreTxn) error {
	spin := l.representCache.getLock(key.String())
	spin.Lock()
	defer spin.Unlock()

	value, err := l.GetRepresentation(key, txns...)
	if err != nil {
		l.logger.Errorf("GetRepresentation error: %s ,address: %s", err, key)
		return err
	}
	value.Balance = value.Balance.Sub(diff.Balance)
	value.Vote = value.Vote.Sub(diff.Vote)
	value.Network = value.Network.Sub(diff.Network)
	value.Oracle = value.Oracle.Sub(diff.Oracle)
	value.Storage = value.Storage.Sub(diff.Storage)
	value.Total = value.Total.Sub(diff.Total)

	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)
	return l.representCache.updateMemory(key, value, txn)
}

func (l *Ledger) getRepresentation(key types.Address, txns ...db.StoreTxn) (*types.Benefit, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixRepresentation, key)
	if err != nil {
		return nil, err
	}

	value := new(types.Benefit)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			l.logger.Errorf("Unmarshal benefit error: %s ,address: %s, val: %s", err, key, string(v))
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrRepresentationNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetRepresentation(key types.Address, txns ...db.StoreTxn) (*types.Benefit, error) {
	if lr, ok := l.representCache.getFromMemory(key.String()); ok {
		return lr.(*types.Benefit), nil
	}
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixRepresentation, key)
	if err != nil {
		return nil, err
	}

	value := new(types.Benefit)
	err = txn.Get(k, func(v []byte, b byte) (err error) {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return &types.Benefit{
				Vote:    types.ZeroBalance,
				Network: types.ZeroBalance,
				Storage: types.ZeroBalance,
				Oracle:  types.ZeroBalance,
				Balance: types.ZeroBalance,
				Total:   types.ZeroBalance,
			}, ErrRepresentationNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetRepresentations(fn func(types.Address, *types.Benefit) error, txns ...db.StoreTxn) error {
	err := l.representCache.iterMemory(func(s string, i interface{}) error {
		address, err := types.HexToAddress(s)
		if err != nil {
			return err
		}
		benefit := i.(*types.Benefit)
		if err := fn(address, benefit); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	err = txn.Iterator(idPrefixRepresentation, func(key []byte, val []byte, b byte) error {
		address, err := types.BytesToAddress(key[1:])
		if err != nil {
			return err
		}
		if _, ok := l.representCache.getFromMemory(address.String()); !ok {
			benefit := new(types.Benefit)
			if err = benefit.Deserialize(val); err != nil {
				l.logger.Errorf("Unmarshal benefit error: %s ,val: %s", err, string(val))
				return err
			}
			if err := fn(address, benefit); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetRepresentationsCache(address types.Address, fn func(address types.Address, am *types.Benefit, amCache *types.Benefit) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	if !address.IsZero() {
		be, err := l.getRepresentation(address)
		if err != nil && err != ErrRepresentationNotFound {
			return err
		}
		if v, ok := l.representCache.getFromMemory(address.String()); ok {
			beMemory := v.(*types.Benefit)
			if err := fn(address, be, beMemory); err != nil {
				return err
			}
		} else {
			if err := fn(address, be, nil); err != nil {
				return err
			}
		}
		return nil
	} else {
		err := l.representCache.iterMemory(func(s string, i interface{}) error {
			addr, err := types.HexToAddress(s)
			if err != nil {
				return err
			}
			beMemory := i.(*types.Benefit)
			be, err := l.getRepresentation(addr)
			if err != nil && err != ErrRepresentationNotFound {
				return err
			}
			if err := fn(addr, be, beMemory); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		err = txn.Iterator(idPrefixRepresentation, func(key []byte, val []byte, b byte) error {
			addr, err := types.BytesToAddress(key[1:])
			if err != nil {
				l.logger.Error(err)
				return err
			}
			be := new(types.Benefit)
			_, err = be.UnmarshalMsg(val)
			if err != nil {
				return err
			}
			if _, ok := l.representCache.getFromMemory(addr.String()); !ok {
				if err := fn(addr, be, nil); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
		return nil
	}
}

func (l *Ledger) AddPending(key *types.PendingKey, value *types.PendingInfo, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPending, key)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrPendingExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetPending(key *types.PendingKey, txns ...db.StoreTxn) (*types.PendingInfo, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPending, key)
	if err != nil {
		return nil, err
	}

	value := new(types.PendingInfo)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPendingNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) GetPendings(fn func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := new(types.PendingKey)
		if err := pendingKey.Deserialize(key[1:]); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		pendingInfo := new(types.PendingInfo)
		if err := pendingInfo.Deserialize(val); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		if err := fn(pendingKey, pendingInfo); err != nil {
			l.logger.Error("process pending error %s", err)
			errStr = append(errStr, err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) SearchPending(address types.Address, fn func(key *types.PendingKey, value *types.PendingInfo) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Stream([]byte{idPrefixPending}, func(item *badger.Item) bool {
		key := &types.PendingKey{}
		if err := key.Deserialize(item.Key()[1:]); err == nil {
			return key.Address == address
		} else {
			l.logger.Error(util.ToString(key), err)
		}

		return false
	}, func(list *pb.KVList) error {
		for _, v := range list.Kv {
			pk := &types.PendingKey{}
			pi := &types.PendingInfo{}
			if err := pk.Deserialize(v.Key[1:]); err != nil {
				continue
			}
			if err := pi.Deserialize(v.Value); err != nil {
				continue
			}
			err := fn(pk, pi)
			if err != nil {
				l.logger.Error(err)
			}
		}
		return nil
	})
}

func (l *Ledger) DeletePending(key *types.PendingKey, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPending, key)
	if err != nil {
		return err
	}
	return txn.Delete(k)
}

func (l *Ledger) AddFrontier(frontier *types.Frontier, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixFrontier, frontier.HeaderBlock)
	if err != nil {
		return err
	}
	v := frontier.OpenBlock[:]

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrFrontierExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetFrontier(key types.Hash, txns ...db.StoreTxn) (*types.Frontier, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixFrontier, key)
	if err != nil {
		return nil, err
	}

	frontier := types.Frontier{HeaderBlock: key}
	err = txn.Get(k, func(v []byte, b byte) (err error) {
		copy(frontier.OpenBlock[:], v)
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrFrontierNotFound
		}
		return nil, err
	}
	return &frontier, nil
}

func (l *Ledger) GetFrontiers(txns ...db.StoreTxn) ([]*types.Frontier, error) {
	var frontiers []*types.Frontier
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixFrontier, func(key []byte, val []byte, b byte) error {
		var frontier types.Frontier
		copy(frontier.HeaderBlock[:], key[1:])
		copy(frontier.OpenBlock[:], val)
		frontiers = append(frontiers, &frontier)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Sort(types.Frontiers(frontiers))
	return frontiers, nil
}

func (l *Ledger) DeleteFrontier(key types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixFrontier, key)
	if err != nil {
		return err
	}
	return txn.Delete(k)
}

func (l *Ledger) CountFrontiers(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Count([]byte{idPrefixFrontier})
}

func (l *Ledger) GetChild(key types.Hash, txns ...db.StoreTxn) (types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixChild, key)
	if err != nil {
		return types.ZeroHash, err
	}

	value := new(types.Hash)
	err = txn.Get(k, func(val []byte, b byte) error {
		if _, err := value.UnmarshalMsg(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return types.ZeroHash, err
	}
	return *value, nil
}

func (l *Ledger) GetLinkBlock(key types.Hash, txns ...db.StoreTxn) (types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixLink, key)
	if err != nil {
		return types.ZeroHash, err
	}
	value := new(types.Hash)
	err = txn.Get(k, func(val []byte, b byte) (err error) {
		if _, err := value.UnmarshalMsg(val); err != nil {
			return errors.New("unmarshal hash error")
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return types.ZeroHash, ErrLinkNotFound
		}
		return types.ZeroHash, fmt.Errorf("get link error: %s", err)
	}
	return *value, nil
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

func (l *Ledger) Latest(account types.Address, token types.Hash, txns ...db.StoreTxn) types.Hash {
	zero := types.Hash{}
	am, err := l.GetAccountMeta(account, txns...)
	if err != nil {
		return zero
	}
	tm := am.Token(token)
	if tm != nil {
		return tm.Header
	}
	return zero
}

func (l *Ledger) Account(hash types.Hash, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	block, err := l.GetStateBlock(hash, txns...)
	if err != nil {
		return nil, err
	}
	addr := block.GetAddress()
	am, err := l.GetAccountMeta(addr, txns...)
	if err != nil {
		return nil, err
	}

	return am, nil
}

func (l *Ledger) Token(hash types.Hash, txns ...db.StoreTxn) (*types.TokenMeta, error) {
	block, err := l.GetStateBlockConfirmed(hash, txns...)
	if err != nil {
		return nil, err
	}
	token := block.GetToken()
	addr := block.GetAddress()
	am, err := l.GetAccountMetaConfirmed(addr, txns...)
	if err != nil {
		return nil, err
	}

	tm := am.Token(token)
	if tm != nil {
		return tm, nil
	}

	//TODO: hash to token name
	return nil, fmt.Errorf("can not find token %s", token)
}

func (l *Ledger) Pending(account types.Address, txns ...db.StoreTxn) ([]*types.PendingKey, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var cache []*types.PendingKey
	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := types.PendingKey{}
		_, err := pendingKey.UnmarshalMsg(key[1:])
		if err != nil {
			return err
		}
		if pendingKey.Address == account {
			cache = append(cache, &pendingKey)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return cache, nil
}

func (l *Ledger) Balance(account types.Address, txns ...db.StoreTxn) (map[types.Hash]types.Balance, error) {
	cache := make(map[types.Hash]types.Balance)
	am, err := l.GetAccountMeta(account, txns...)
	if err != nil {
		return cache, err
	}
	for _, tm := range am.Tokens {
		cache[tm.Type] = tm.Balance
	}

	if len(cache) == 0 {
		return nil, fmt.Errorf("can not find any token balance ")
	}

	return cache, nil
}

func (l *Ledger) TokenPending(account types.Address, token types.Hash, txns ...db.StoreTxn) ([]*types.PendingKey, error) {
	var cache []*types.PendingKey
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := types.PendingKey{}
		_, err := pendingKey.UnmarshalMsg(key[1:])
		if err != nil {
			return nil
		}
		if pendingKey.Address == account {
			var pending types.PendingInfo
			_, err := pending.UnmarshalMsg(val)
			if err != nil {
				return err
			}
			if pending.Type == token {
				cache = append(cache, &pendingKey)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return cache, nil
}

func (l *Ledger) TokenPendingInfo(account types.Address, token types.Hash, txns ...db.StoreTxn) ([]*types.PendingInfo, error) {
	var cache []*types.PendingInfo
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := types.PendingKey{}
		_, err := pendingKey.UnmarshalMsg(key[1:])
		if err != nil {
			return nil
		}
		if pendingKey.Address == account {
			var pendingInfo types.PendingInfo
			_, err := pendingInfo.UnmarshalMsg(val)
			if err != nil {
				return err
			}
			if pendingInfo.Type == token {
				cache = append(cache, &pendingInfo)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return cache, nil
}

func (l *Ledger) TokenBalance(account types.Address, token types.Hash, txns ...db.StoreTxn) (types.Balance, error) {
	am, err := l.GetAccountMeta(account, txns...)
	if err != nil {
		return types.ZeroBalance, err
	}
	tm := am.Token(token)
	if tm != nil {
		return tm.Balance, nil
	}

	return types.ZeroBalance, fmt.Errorf("can not find %s balance", token)
}

func (l *Ledger) Weight(account types.Address, txns ...db.StoreTxn) types.Balance {
	benefit, err := l.GetRepresentation(account, txns...)
	if err != nil {
		return types.ZeroBalance
	}
	return benefit.Total
}

func (l *Ledger) AddOrUpdatePerformance(value *types.PerformanceTime, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPerformance, value.Hash)
	if err != nil {
		return err
	}
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return txn.Set(k, v)
}

func (l *Ledger) PerformanceTimes(fn func(*types.PerformanceTime), txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPerformance, func(key []byte, val []byte, b byte) error {
		pt := new(types.PerformanceTime)
		err := json.Unmarshal(val, pt)
		if err != nil {
			return err
		}
		fn(pt)
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetPerformanceTime(key types.Hash, txns ...db.StoreTxn) (*types.PerformanceTime, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixPerformance, key)
	if err != nil {
		return nil, err
	}

	value := types.NewPerformanceTime()
	err = txn.Get(k, func(val []byte, b byte) (err error) {
		return json.Unmarshal(val, &value)
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPerformanceNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) IsPerformanceTimeExist(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	if _, err := l.GetPerformanceTime(key); err == nil {
		return true, nil
	} else {
		if err == ErrPerformanceNotFound {
			return false, nil
		} else {
			return false, err
		}
	}
}

func (l *Ledger) CalculateAmount(block *types.StateBlock, txns ...db.StoreTxn) (types.Balance, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var prev *types.StateBlock
	var err error
	switch block.GetType() {
	case types.Open:
		return block.TotalBalance(), err
	case types.Send:
		if prev, err = l.GetStateBlock(block.GetPrevious(), txn); err != nil {
			return types.ZeroBalance, err
		}
		return prev.TotalBalance().Sub(block.TotalBalance()), nil
	case types.Receive:
		if prev, err = l.GetStateBlock(block.GetPrevious(), txn); err != nil {
			return types.ZeroBalance, err
		}
		return block.TotalBalance().Sub(prev.TotalBalance()), nil
	case types.Change:
		return types.ZeroBalance, nil
	case types.ContractReward:
		prevHash := block.GetPrevious()
		if prevHash.IsZero() {
			return block.TotalBalance(), nil
		} else {
			if prev, err = l.GetStateBlock(prevHash, txn); err != nil {
				return types.ZeroBalance, err
			}
			return block.TotalBalance().Sub(prev.TotalBalance()), nil
		}
	case types.ContractSend:
		if common.IsGenesisBlock(block) {
			return block.GetBalance(), nil
		} else {
			if prev, err = l.GetStateBlock(block.GetPrevious(), txn); err != nil {
				return types.ZeroBalance, err
			}
			return prev.TotalBalance().Sub(block.TotalBalance()), nil
		}
	default:
		return types.ZeroBalance, errors.New("invalid block type")
	}
}

func (l *Ledger) GetGenesis(txns ...db.StoreTxn) ([]*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var blocks []*types.StateBlock
	err := txn.Iterator(idPrefixToken, func(key []byte, val []byte, b byte) error {
		t := new(types.TokenInfo)
		if err := json.Unmarshal(val, t); err != nil {
			return err
		}
		tm, err := l.GetTokenMeta(t.Owner, t.TokenId, txn)
		if err != nil {
			return err
		}
		block, err := l.GetStateBlock(tm.OpenBlock, txn)
		if err != nil {
			return err
		}
		blocks = append(blocks, block)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blocks, nil
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

func (l *Ledger) GetOnlineRepresentations(txns ...db.StoreTxn) ([]types.Address, error) {
	key := []byte{idPrefixOnlineReps}
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)
	var result []types.Address
	err := txn.Get(key, func(bytes []byte, b byte) error {
		if len(bytes) > 0 {
			return json.Unmarshal(bytes, &result)
		}
		return nil
	})

	return result, err
}

func (l *Ledger) SetOnlineRepresentations(addresses []*types.Address, txns ...db.StoreTxn) error {
	bytes, err := json.Marshal(addresses)
	if err != nil {
		return err
	}

	key := []byte{idPrefixOnlineReps}
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, bytes)
}

func (l *Ledger) AddMessageInfo(key types.Hash, value []byte, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixMessageInfo, key)
	if err != nil {
		return err
	}
	if err := txn.Set(k, value); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetMessageInfo(key types.Hash, txns ...db.StoreTxn) ([]byte, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixMessageInfo, key)
	if err != nil {
		return nil, err
	}
	var value []byte
	err = txn.Get(k, func(v []byte, b byte) error {
		value = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (l *Ledger) AddBlockCache(value *types.StateBlock, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCache, value.GetHash())
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetBlockCache(key types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCache, key)
	if err != nil {
		return nil, err
	}

	value := new(types.StateBlock)
	err = txn.Get(k, func(v []byte, b byte) error {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) HasBlockCache(key types.Hash, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCache, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *Ledger) DeleteBlockCache(key types.Hash, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCache, key)
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err != nil {
		return err
	}

	if err := txn.Delete(k); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) CountBlockCache(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Count([]byte{idPrefixBlockCache})
}

func (l *Ledger) GetBlockCaches(fn func(*types.StateBlock) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixBlockCache, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			l.logger.Errorf("deserialize block error: %s", err)
			errStr = append(errStr, err.Error())
			return nil
		}
		if err := fn(blk); err != nil {
			l.logger.Errorf("process block error: %s", err)
			errStr = append(errStr, err.Error())
		}
		return nil
	})

	if err != nil {
		return err
	}
	if len(errStr) != 0 {
		return errors.New(strings.Join(errStr, ", "))
	}
	return nil
}

func (l *Ledger) AddAccountMetaCache(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(v []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrAccountExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) GetAccountMetaCache(key types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, key)
	if err != nil {
		return nil, err
	}
	value := new(types.AccountMeta)
	err = txn.Get(k, func(v []byte, b byte) (err error) {
		if err := value.Deserialize(v); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return value, nil
}

func (l *Ledger) AddOrUpdateAccountMetaCache(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) UpdateAccountMetaCache(value *types.AccountMeta, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, value.Address)
	if err != nil {
		return err
	}
	v, err := value.Serialize()
	if err != nil {
		return err
	}

	err = txn.Get(k, func(vals []byte, b byte) error {
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrAccountNotFound
		}
		return err
	}
	return txn.Set(k, v)
}

func (l *Ledger) DeleteAccountMetaCache(key types.Address, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, key)
	if err != nil {
		return err
	}

	return txn.Delete(k)
}

func (l *Ledger) HasAccountMetaCache(key types.Address, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	k, err := getKeyOfParts(idPrefixBlockCacheAccount, key)
	if err != nil {
		return false, err
	}
	err = txn.Get(k, func(val []byte, b byte) error {
		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
