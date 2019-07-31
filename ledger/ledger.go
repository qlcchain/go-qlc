package ledger

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/sync/hashmap"
	"github.com/qlcchain/go-qlc/common/sync/spinlock"
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
	representation *hashmap.HashMap
	representLock  *hashmap.HashMap
	cacheRound     *int64
	cacheOrder     *int64
	closed         chan bool
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

const version = 6

func NewLedger(dir string) *Ledger {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[dir]; !ok {
		store, err := db.NewBadgerStore(dir)
		if err != nil {
			fmt.Println(err.Error())
		}
		l := &Ledger{
			Store:          store,
			dir:            dir,
			EB:             event.GetEventBus(dir),
			representation: &hashmap.HashMap{},
			representLock:  &hashmap.HashMap{},
			closed:         make(chan bool, 1),
		}
		l.logger = log.NewLogger("ledger")

		if err := l.upgrade(); err != nil {
			l.logger.Error(err)
		}
		if err := l.init(); err != nil {
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
		l.closed <- true
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
		ms := []db.Migration{new(MigrationV1ToV2), new(MigrationV2ToV3), new(MigrationV3ToV4),
			new(MigrationV4ToV5), new(MigrationV5ToV6)}

		err = txn.Upgrade(ms)
		if err != nil {
			l.logger.Error(err)
		}
		return err
	})
}

func (l *Ledger) init() error {
	err := l.setCacheToDB(nil)
	if err != nil {
		l.logger.Error(err)
		return err
	}
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for {
			select {
			case <-l.closed:
				return
			case <-ticker.C:
				l.processCache()
			}
		}
	}()
	l.cacheRound = new(int64)
	l.cacheOrder = new(int64)
	return nil
}

func (l *Ledger) processCache() {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[l.dir]; ok {
		cacheRound := atomic.LoadInt64(l.cacheRound)
		atomic.StoreInt64(l.cacheOrder, 0)
		atomic.AddInt64(l.cacheRound, 1)
		l.setCacheToDB(&cacheRound)
	}
}

func (l *Ledger) setCacheToDB(cacheRound *int64) error {
	txn := l.Store.NewTransaction(true)
	defer txn.Commit(nil)

	var roundTemp int64
	beCacheArray := make(map[types.Address]*types.Benefit)
	beOrderArray := make(map[types.Address]int64)
	err := txn.Iterator(idPrefixRepresentationCache, func(cacheKey []byte, cacheVal []byte, b byte) error {
		addrCache, err := types.BytesToAddress(cacheKey[1 : 1+types.AddressSize])
		if err != nil {
			l.logger.Error(err)
			return err
		}
		var beCache types.Benefit
		if _, err := beCache.UnmarshalMsg(cacheVal); err != nil {
			l.logger.Error(err)
			return err
		}
		round := int64(util.BE_BytesToUint64(cacheKey[1+types.AddressSize : 1+types.AddressSize+8]))
		order := int64(util.BE_BytesToUint64(cacheKey[1+types.AddressSize+8:]))

		if cacheRound == nil {
			if round > roundTemp {
				roundTemp = round
				beOrderArray[addrCache] = order
				beCacheArray[addrCache] = &beCache
			} else {
				if _, ok := beCacheArray[addrCache]; ok {
					if order > beOrderArray[addrCache] {
						beOrderArray[addrCache] = order
						beCacheArray[addrCache] = &beCache
					}
				} else {
					beOrderArray[addrCache] = order
					beCacheArray[addrCache] = &beCache
				}
			}
		} else {
			if round == *cacheRound {
				if _, ok := beCacheArray[addrCache]; ok {
					if order > beOrderArray[addrCache] {
						beOrderArray[addrCache] = order
						beCacheArray[addrCache] = &beCache
					}
				} else {
					beOrderArray[addrCache] = order
					beCacheArray[addrCache] = &beCache
				}
				if err := txn.Delete(cacheKey); err != nil {
					l.logger.Error(err)
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		l.logger.Error(err)
		return err
	}
	for key, val := range beCacheArray {
		if err := l.updateRepresentation(key, val, txn); err != nil {
			l.logger.Error(err)
			return err
		}
	}

	if cacheRound == nil {
		if err := txn.Drop([]byte{idPrefixRepresentationCache}); err != nil {
			return err
		}
	}
	return nil
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

func getKeyOfHash(hash types.Hash, t byte) []byte {
	var key [1 + types.HashSize]byte
	key[0] = t
	copy(key[1:], hash[:])
	return key[:]
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
		default:
			return nil, errors.New("Key contains of invalid part.")
		}

		buffer = append(buffer, src...)
	}

	return buffer, nil
}

func (l *Ledger) AddStateBlock(blk *types.StateBlock, txns ...db.StoreTxn) error {
	key := getKeyOfHash(blk.GetHash(), idPrefixBlock)
	txn, flag := l.getTxn(true, txns...)

	//never overwrite implicitly
	err := txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	blockBytes, err := blk.Serialize()
	if err != nil {
		return fmt.Errorf("serialize block error: %s", err)
	}
	if err := txn.Set(key, blockBytes); err != nil {
		return err
	}
	if b, err := l.HasBlockCache(blk.GetHash()); b && err == nil {
		if err := l.DeleteBlockCache(blk.GetHash(), txn); err != nil {
			return fmt.Errorf("delete block cache error: %s", err)
		}
	}
	if err := addChild(blk, txn); err != nil {
		return fmt.Errorf("add block child error: %s", err)
	}
	if err := addLink(blk, txn); err != nil {
		return fmt.Errorf("add block link error: %s", err)
	}
	l.releaseTxn(txn, flag)
	l.logger.Debug("publish addRelation,", blk.GetHash())
	l.EB.Publish(common.EventAddRelation, blk)
	return nil
}

func addChild(cBlock *types.StateBlock, txn db.StoreTxn) error {
	pHash := cBlock.Parent()
	cHash := cBlock.GetHash()
	if !common.IsGenesisBlock(cBlock) && pHash != types.ZeroHash && !cBlock.IsOpen() {
		// is parent block existed
		err := txn.Get(getKeyOfHash(pHash, idPrefixBlockCache), func(val []byte, b byte) error {
			return nil
		})
		if err != nil {
			err := txn.Get(getKeyOfHash(pHash, idPrefixBlock), func(val []byte, b byte) error {
				return nil
			})
			if err != nil {
				return fmt.Errorf("%s can not find parent %s", cHash.String(), pHash.String())
			}
		}

		// is parent have used
		pKey := getKeyOfHash(pHash, idPrefixChild)
		err = txn.Get(pKey, func(val []byte, b byte) error {
			return nil
		})
		if err == nil {
			return fmt.Errorf("%s already have child ", pHash.String())
		}

		// add new relationship
		val, err := cHash.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := txn.Set(pKey, val); err != nil {
			return err
		}
	}
	return nil
}

func addLink(block *types.StateBlock, txn db.StoreTxn) error {
	if block.GetType() == types.Open || block.GetType() == types.Receive || block.GetType() == types.ContractReward {
		key := getKeyOfHash(block.GetLink(), idPrefixLink)
		h := block.GetHash()
		val, err := h.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := txn.Set(key, val); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) GetStateBlock(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	if blkCache, err := l.GetBlockCache(hash); err == nil {
		return blkCache, nil
	}
	key := getKeyOfHash(hash, idPrefixBlock)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	blk := new(types.StateBlock)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := blk.Deserialize(val); err != nil {
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
	return blk, nil
}

func (l *Ledger) GetStateBlockConfirmed(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	key := getKeyOfHash(hash, idPrefixBlock)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	blk := new(types.StateBlock)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := blk.Deserialize(val); err != nil {
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
	return blk, nil
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

func (l *Ledger) DeleteStateBlock(hash types.Hash, txns ...db.StoreTxn) error {
	if b, err := l.HasBlockCache(hash); b && err == nil {
		if err = l.DeleteBlockCache(hash); err != nil {
			return fmt.Errorf("delete block cache fail(%s), hash(%s)", err, hash)
		}
		return nil
	}

	key := getKeyOfHash(hash, idPrefixBlock)
	txn, flag := l.getTxn(true, txns...)

	blk := new(types.StateBlock)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := blk.Deserialize(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil && err != ErrBlockNotFound {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}
	if err := l.deleteChild(blk, txn); err != nil {
		return fmt.Errorf("delete child error: %s", err)
	}
	if err := l.deleteLink(blk, txn); err != nil {
		return fmt.Errorf("delete link error: %s", err)
	}

	l.releaseTxn(txn, flag)
	l.logger.Info("publish deleteRelation,", hash.String())
	l.EB.Publish(common.EventDeleteRelation, hash)
	return nil
}

func (l *Ledger) deleteChild(blk *types.StateBlock, txn db.StoreTxn) error {
	pHash := blk.Parent()
	if !pHash.IsZero() {
		pKey := getKeyOfHash(pHash, idPrefixChild)
		if err := txn.Delete(pKey); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) deleteLink(blk *types.StateBlock, txn db.StoreTxn) error {
	pKey := getKeyOfHash(blk.GetLink(), idPrefixLink)
	if err := txn.Delete(pKey); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) HasStateBlock(hash types.Hash, txns ...db.StoreTxn) (bool, error) {
	if exit, err := l.HasBlockCache(hash); err == nil && exit {
		return exit, nil
	}
	key := getKeyOfHash(hash, idPrefixBlock)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) error {
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

func (l *Ledger) HasStateBlockConfirmed(hash types.Hash, txns ...db.StoreTxn) (bool, error) {
	key := getKeyOfHash(hash, idPrefixBlock)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) error {
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

func (l *Ledger) AddSmartContractBlock(blk *types.SmartContractBlock, txns ...db.StoreTxn) error {
	key := getKeyOfHash(blk.GetHash(), idPrefixSmartContractBlock)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	//never overwrite implicitly
	err := txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	blockBytes, err := blk.Serialize()
	if err != nil {
		return err
	}
	return txn.Set(key, blockBytes)
}

func (l *Ledger) GetSmartContractBlock(hash types.Hash, txns ...db.StoreTxn) (*types.SmartContractBlock, error) {
	key := getKeyOfHash(hash, idPrefixSmartContractBlock)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	blk := new(types.SmartContractBlock)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := blk.Deserialize(val); err != nil {
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
	return blk, nil
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

func (l *Ledger) HasSmartContractBlock(hash types.Hash, txns ...db.StoreTxn) (bool, error) {
	key := getKeyOfHash(hash, idPrefixSmartContractBlock)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) error {
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

func (l *Ledger) getUncheckedBlockKey(hash types.Hash, kind types.UncheckedKind) []byte {
	var key [1 + types.HashSize]byte
	key[0] = l.uncheckedKindToPrefix(kind)
	copy(key[1:], hash[:])
	return key[:]
}

func (l *Ledger) AddUncheckedBlock(parentHash types.Hash, blk *types.StateBlock, kind types.UncheckedKind, sync types.SynchronizedKind, txns ...db.StoreTxn) error {
	blockBytes, err := blk.Serialize()
	if err != nil {
		return err
	}

	key := l.getUncheckedBlockKey(parentHash, kind)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	//never overwrite implicitly
	err = txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrUncheckedBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	return txn.SetWithMeta(key, blockBytes, byte(sync))
}

func (l *Ledger) GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (*types.StateBlock, types.SynchronizedKind, error) {
	key := l.getUncheckedBlockKey(parentHash, kind)

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	blk := new(types.StateBlock)
	var sync types.SynchronizedKind
	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if err = blk.Deserialize(val); err != nil {
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
	return blk, sync, nil
}

func (l *Ledger) DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) error {
	key := l.getUncheckedBlockKey(parentHash, kind)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Delete(key)
}

func (l *Ledger) HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (bool, error) {
	key := l.getUncheckedBlockKey(hash, kind)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) error {
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
	var count uint64
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

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

func getAccountMetaKey(address types.Address) []byte {
	var key [1 + types.AddressSize]byte
	key[0] = idPrefixAccount
	copy(key[1:], address[:])
	return key[:]
}

func (l *Ledger) AddAccountMeta(meta *types.AccountMeta, txns ...db.StoreTxn) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := getAccountMetaKey(meta.Address)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	// never overwrite implicitly
	err = txn.Get(key, func(vals []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrAccountExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(key, metaBytes)
}

func (l *Ledger) GetAccountMeta(address types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	am, err := l.GetAccountMetaCache(address)
	if am != nil && err == nil {
		return am, nil
	}
	var meta types.AccountMeta
	key := getAccountMetaKey(address)

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err = txn.Get(key, func(val []byte, b byte) (err error) {
		if _, err = meta.UnmarshalMsg(val); err != nil {
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
	return &meta, nil
}

func (l *Ledger) GetAccountMetaConfirmed(address types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	var meta types.AccountMeta
	key := getAccountMetaKey(address)

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if _, err = meta.UnmarshalMsg(val); err != nil {
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
	return &meta, nil
}

func (l *Ledger) GetAccountMetas(fn func(am *types.AccountMeta) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixAccount, func(key []byte, val []byte, b byte) error {
		am := new(types.AccountMeta)
		_, err := am.UnmarshalMsg(val)
		if err != nil {
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

func (l *Ledger) UpdateAccountMeta(meta *types.AccountMeta, txns ...db.StoreTxn) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := getAccountMetaKey(meta.Address)

	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err = txn.Get(key, func(vals []byte, b byte) error {
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrAccountNotFound
		}
		return err
	}
	return txn.Set(key, metaBytes)
}

func (l *Ledger) AddOrUpdateAccountMeta(meta *types.AccountMeta, txns ...db.StoreTxn) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := getAccountMetaKey(meta.Address)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, metaBytes)
}

func (l *Ledger) DeleteAccountMeta(address types.Address, txns ...db.StoreTxn) error {
	key := getAccountMetaKey(address)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Delete(key)
}

func (l *Ledger) HasAccountMeta(address types.Address, txns ...db.StoreTxn) (bool, error) {
	key := getAccountMetaKey(address)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) error {
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
	am, err := l.GetAccountMeta(address, txns...)
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

func (l *Ledger) AddRepresentation(address types.Address, diff *types.Benefit, txns ...db.StoreTxn) error {
	i, _ := l.representLock.GetOrInsert(address.String(), &spinlock.SpinLock{})
	spin, _ := i.(*spinlock.SpinLock)
	spin.Lock()
	defer spin.Unlock()

	benefit, err := l.GetRepresentation(address, txns...)
	if err != nil && err != ErrRepresentationNotFound {
		l.logger.Errorf("GetRepresentation error: %s ,address: %s", err, address)
		return err
	}

	benefit.Balance = benefit.Balance.Add(diff.Balance)
	benefit.Vote = benefit.Vote.Add(diff.Vote)
	benefit.Network = benefit.Network.Add(diff.Network)
	benefit.Oracle = benefit.Oracle.Add(diff.Oracle)
	benefit.Storage = benefit.Storage.Add(diff.Storage)
	benefit.Total = benefit.Total.Add(diff.Total)

	return l.updateRepresentationCache(address, benefit, txns...)
}

func (l *Ledger) SubRepresentation(address types.Address, diff *types.Benefit, txns ...db.StoreTxn) error {
	i, _ := l.representLock.GetOrInsert(address.String(), &spinlock.SpinLock{})
	spin, _ := i.(*spinlock.SpinLock)
	spin.Lock()
	defer spin.Unlock()

	benefit, err := l.GetRepresentation(address, txns...)
	if err != nil {
		l.logger.Errorf("GetRepresentation error: %s ,address: %s", err, address)
		return err
	}
	benefit.Balance = benefit.Balance.Sub(diff.Balance)
	benefit.Vote = benefit.Vote.Sub(diff.Vote)
	benefit.Network = benefit.Network.Sub(diff.Network)
	benefit.Oracle = benefit.Oracle.Sub(diff.Oracle)
	benefit.Storage = benefit.Storage.Sub(diff.Storage)
	benefit.Total = benefit.Total.Sub(diff.Total)

	return l.updateRepresentationCache(address, benefit, txns...)
}

func (l *Ledger) updateRepresentation(address types.Address, benefit *types.Benefit, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixRepresentation, address)
	if err != nil {
		return err
	}
	val, err := benefit.MarshalMsg(nil)
	if err != nil {
		l.logger.Errorf("MarshalMsg benefit error: %s ,address: %s, val: %s", err, address, benefit)
		return err
	}
	return txn.Set(key, val)
}

func (l *Ledger) getRepresentation(address types.Address, txns ...db.StoreTxn) (*types.Benefit, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixRepresentation, address)
	if err != nil {
		return nil, err
	}
	benefit := new(types.Benefit)
	err = txn.Get(key, func(val []byte, b byte) (err error) {
		if _, err = benefit.UnmarshalMsg(val); err != nil {
			l.logger.Errorf("Unmarshal benefit error: %s ,address: %s, val: %s", err, address, string(val))
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
	return benefit, nil
}

func (l *Ledger) updateRepresentationCache(address types.Address, benefit *types.Benefit, txns ...db.StoreTxn) error {
	l.representation.Set(address.String(), benefit)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixRepresentationCache, address, *l.cacheRound, atomic.AddInt64(l.cacheOrder, 1))
	if err != nil {
		return err
	}
	val, err := benefit.MarshalMsg(nil)
	if err != nil {
		l.logger.Errorf("MarshalMsg benefit error: %s ,address: %s, val: %s", err, address, benefit)
		return err
	}

	if err := txn.Set(key, val); err != nil {
		l.logger.Error(err)
		return err
	}
	return nil
}

func (l *Ledger) GetRepresentation(address types.Address, txns ...db.StoreTxn) (*types.Benefit, error) {
	if lr, ok := l.representation.Get(address.String()); ok {
		return lr.(*types.Benefit), nil
	}
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key, err := getKeyOfParts(idPrefixRepresentation, address)
	if err != nil {
		return nil, err
	}
	benefit := new(types.Benefit)
	err = txn.Get(key, func(val []byte, b byte) (err error) {
		if _, err = benefit.UnmarshalMsg(val); err != nil {
			l.logger.Errorf("Unmarshal benefit error: %s ,address: %s, val: %s", err, address, string(val))
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
	return benefit, nil
}

func (l *Ledger) GetRepresentations(fn func(types.Address, *types.Benefit) error, txns ...db.StoreTxn) error {
	for kv := range l.representation.Iter() {
		aStr := kv.Key.(string)
		address, err := types.HexToAddress(aStr)
		if err != nil {
			return err
		}
		benefit := kv.Value.(*types.Benefit)
		if err := fn(address, benefit); err != nil {
			return err
		}
	}

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	err := txn.Iterator(idPrefixRepresentation, func(key []byte, val []byte, b byte) error {
		address, err := types.BytesToAddress(key[1:])
		if err != nil {
			return err
		}
		if _, ok := l.representation.Get(address.String()); !ok {
			benefit := new(types.Benefit)
			if _, err = benefit.UnmarshalMsg(val); err != nil {
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
		be, _ := l.getRepresentation(address)
		if v, ok := l.representation.Get(address.String()); ok {
			beCache := v.(*types.Benefit)
			if err := fn(address, be, beCache); err != nil {
				return err
			}
		} else {
			if err := fn(address, be, nil); err != nil {
				return err
			}
		}
		return nil
	}

	for kv := range l.representation.Iter() {
		s := kv.Key.(string)
		addr, err := types.HexToAddress(s)
		if err != nil {
			return err
		}
		beCache := kv.Value.(*types.Benefit)
		be, _ := l.getRepresentation(addr)
		if err := fn(addr, be, beCache); err != nil {
			return err
		}
	}

	err := txn.Iterator(idPrefixRepresentation, func(key []byte, val []byte, b byte) error {
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
		if _, ok := l.representation.Get(addr.String()); !ok {
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

func (l *Ledger) getPendingKey(pendingKey types.PendingKey) []byte {
	//key := []byte{idPrefixPending}
	//_, _ = pendingKey.MarshalMsg(key[1:])
	//return key[:]

	msg, _ := pendingKey.MarshalMsg(nil)
	key := make([]byte, 1+len(msg))
	key[0] = idPrefixPending
	copy(key[1:], msg)
	return key[:]

}

func (l *Ledger) AddPending(pendingKey *types.PendingKey, pending *types.PendingInfo, txns ...db.StoreTxn) error {
	pendingBytes, err := pending.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getPendingKey(*pendingKey)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	//never overwrite implicitly
	err = txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrPendingExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(key, pendingBytes)
}

func (l *Ledger) GetPending(pendingKey types.PendingKey, txns ...db.StoreTxn) (*types.PendingInfo, error) {
	key := l.getPendingKey(pendingKey)
	var pending types.PendingInfo
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key[:], func(val []byte, b byte) (err error) {
		if _, err = pending.UnmarshalMsg(val); err != nil {
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
	return &pending, nil

}

func (l *Ledger) GetPendings(fn func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	errStr := make([]string, 0)
	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := new(types.PendingKey)
		if _, err := pendingKey.UnmarshalMsg(key[1:]); err != nil {
			errStr = append(errStr, err.Error())
			return nil
		}
		pendingInfo := new(types.PendingInfo)
		if _, err := pendingInfo.UnmarshalMsg(val); err != nil {
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
		if _, err := key.UnmarshalMsg(item.Key()[1:]); err == nil {
			return key.Address == address
		} else {
			l.logger.Error(util.ToString(key), err)
		}
		return false
	}, func(list *pb.KVList) error {
		for _, v := range list.Kv {
			pk := &types.PendingKey{}
			pi := &types.PendingInfo{}

			if _, err := pk.UnmarshalMsg(v.Key[1:]); err != nil {
				continue
			}
			if _, err := pi.UnmarshalMsg(v.Value); err != nil {
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

func (l *Ledger) DeletePending(pendingKey *types.PendingKey, txns ...db.StoreTxn) error {
	key := l.getPendingKey(*pendingKey)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Delete(key)
}

func (l *Ledger) AddFrontier(frontier *types.Frontier, txns ...db.StoreTxn) error {
	key := getKeyOfHash(frontier.HeaderBlock, idPrefixFrontier)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	// never overwrite implicitly
	err := txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrFrontierExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(key, frontier.OpenBlock[:])
}

func (l *Ledger) GetFrontier(hash types.Hash, txns ...db.StoreTxn) (*types.Frontier, error) {
	key := getKeyOfHash(hash, idPrefixFrontier)
	frontier := types.Frontier{HeaderBlock: hash}
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) (err error) {
		copy(frontier.OpenBlock[:], val)
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

func (l *Ledger) DeleteFrontier(hash types.Hash, txns ...db.StoreTxn) error {
	key := getKeyOfHash(hash, idPrefixFrontier)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Delete(key)
}

func (l *Ledger) CountFrontiers(txns ...db.StoreTxn) (uint64, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Count([]byte{idPrefixFrontier})
}

func (l *Ledger) GetChild(hash types.Hash, txns ...db.StoreTxn) (types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key := getKeyOfHash(hash, idPrefixChild)
	var h types.Hash
	err := txn.Get(key, func(val []byte, b byte) error {
		if _, err := h.UnmarshalMsg(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return types.ZeroHash, err
	}
	return h, nil
}

func (l *Ledger) GetLinkBlock(hash types.Hash, txns ...db.StoreTxn) (types.Hash, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key := getKeyOfHash(hash, idPrefixLink)
	h := new(types.Hash)
	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if _, err := h.UnmarshalMsg(val); err != nil {
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
	return *h, nil
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
	block, err := l.GetStateBlock(hash, txns...)
	if err != nil {
		return nil, err
	}
	token := block.GetToken()
	addr := block.GetAddress()
	am, err := l.GetAccountMeta(addr, txns...)
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
	var cache []*types.PendingKey
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

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

func (l *Ledger) AddOrUpdatePerformance(p *types.PerformanceTime, txns ...db.StoreTxn) error {
	key := l.getPerformanceKey(p.Hash)
	bytes, err := json.Marshal(p)
	if err != nil {
		return err
	}
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, bytes)
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

func (l *Ledger) GetPerformanceTime(hash types.Hash, txns ...db.StoreTxn) (*types.PerformanceTime, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key := l.getPerformanceKey(hash)
	pt := types.NewPerformanceTime()
	err := txn.Get(key, func(val []byte, b byte) (err error) {
		return json.Unmarshal(val, &pt)
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrPerformanceNotFound
		}
		return nil, err
	}
	return pt, nil
}

func (l *Ledger) getPerformanceKey(hash types.Hash) []byte {
	var key []byte
	key = append(key, idPrefixPerformance)
	key = append(key, hash[:]...)
	return key
}

func (l *Ledger) IsPerformanceTimeExist(hash types.Hash, txns ...db.StoreTxn) (bool, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	if _, err := l.GetPerformanceTime(hash); err == nil {
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

//func (l *Ledger) ListTokens(txns ...db.StoreTxn) ([]*types.TokenInfo, error) {
//	txn, flag := l.getTxn(false, txns...)
//	defer l.releaseTxn(txn, flag)
//
//	var tokens []*types.TokenInfo
//	err := txn.Iterator(idPrefixToken, func(key []byte, val []byte, b byte) error {
//		token := new(types.TokenInfo)
//		if err := json.Unmarshal(val, token); err != nil {
//			fmt.Println(err)
//			return err
//		}
//		tokens = append(tokens, token)
//		return nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	return tokens, nil
//}
//
//func (l *Ledger) GetTokenById(tokenId types.Hash, txns ...db.StoreTxn) (*types.TokenInfo, error) {
//	txn, flag := l.getTxn(false, txns...)
//	defer l.releaseTxn(txn, flag)
//
//	key := getKeyOfHash(tokenId, idPrefixToken)
//	token := new(types.TokenInfo)
//	err := txn.Get(key, func(val []byte, b byte) (err error) {
//		if err := json.Unmarshal(val, token); err != nil {
//			return err
//		}
//		return nil
//	})
//	if err != nil {
//		if err == badger.ErrKeyNotFound {
//			return nil, ErrTokenInfoNotFound
//		}
//		return nil, err
//	}
//	return token, nil
//}
//
//func (l *Ledger) GetTokenByName(tokenName string, txns ...db.StoreTxn) (*types.TokenInfo, error) {
//	txn, flag := l.getTxn(false, txns...)
//	defer l.releaseTxn(txn, flag)
//
//	var token *types.TokenInfo
//	err := txn.Iterator(idPrefixToken, func(key []byte, val []byte, b byte) error {
//		t := new(types.TokenInfo)
//		if err := json.Unmarshal(val, t); err != nil {
//			return err
//		}
//		if t.TokenName == tokenName {
//			token = t
//		}
//		return nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	if token == nil {
//		return nil, ErrTokenInfoNotFound
//	}
//	return token, nil
//}

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
	info, err := l.GetPending(types.PendingKey{Address: rxAccount, Hash: hash})
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

	if prk != nil {
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

func (l *Ledger) AddMessageInfo(mHash types.Hash, message []byte, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	key := getKeyOfHash(mHash, idPrefixMessageInfo)
	if err := txn.Set(key, message); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetMessageInfo(mHash types.Hash, txns ...db.StoreTxn) ([]byte, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key := getKeyOfHash(mHash, idPrefixMessageInfo)
	var m []byte
	err := txn.Get(key, func(val []byte, b byte) error {
		m = val
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (l *Ledger) AddBlockCache(blk *types.StateBlock, txns ...db.StoreTxn) error {
	key := getKeyOfHash(blk.GetHash(), idPrefixBlockCache)
	txn, flag := l.getTxn(true, txns...)

	//never overwrite implicitly
	err := txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}

	blockBytes, err := blk.Serialize()
	if err != nil {
		return fmt.Errorf("serialize block error: %s", err)
	}
	if err := txn.Set(key, blockBytes); err != nil {
		return err
	}
	//if err := addChild(blk, txn); err != nil {
	//	return fmt.Errorf("add block child error: %s", err)
	//}
	//if err := addLink(blk, txn); err != nil {
	//	return fmt.Errorf("add block link error: %s", err)
	//}
	l.releaseTxn(txn, flag)
	return nil
}

func (l *Ledger) GetBlockCache(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
	key := getKeyOfHash(hash, idPrefixBlockCache)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	blk := new(types.StateBlock)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := blk.Deserialize(val); err != nil {
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
	return blk, nil
}

func (l *Ledger) HasBlockCache(hash types.Hash, txns ...db.StoreTxn) (bool, error) {
	key := getKeyOfHash(hash, idPrefixBlockCache)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) error {
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

func (l *Ledger) DeleteBlockCache(hash types.Hash, txns ...db.StoreTxn) error {
	key := getKeyOfHash(hash, idPrefixBlockCache)
	txn, flag := l.getTxn(true, txns...)

	blk := new(types.StateBlock)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := blk.Deserialize(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil && err != ErrBlockNotFound {
		return err
	}

	if err := txn.Delete(key); err != nil {
		return err
	}
	//if err := l.deleteChild(blk, txn); err != nil {
	//	return fmt.Errorf("delete child error: %s", err)
	//}
	//if err := l.deleteLink(blk, txn); err != nil {
	//	return fmt.Errorf("delete link error: %s", err)
	//}
	l.releaseTxn(txn, flag)
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

func (l *Ledger) getAccountMetaCacheKey(address types.Address) []byte {
	var key [1 + types.AddressSize]byte
	key[0] = idPrefixBlockCacheAccount
	copy(key[1:], address[:])
	return key[:]
}

func (l *Ledger) AddAccountMetaCache(meta *types.AccountMeta, txns ...db.StoreTxn) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := l.getAccountMetaCacheKey(meta.Address)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	// never overwrite implicitly
	err = txn.Get(key, func(vals []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrAccountExists
	} else if err != badger.ErrKeyNotFound {
		return err
	}
	return txn.Set(key, metaBytes)
}

func (l *Ledger) GetAccountMetaCache(address types.Address, txns ...db.StoreTxn) (*types.AccountMeta, error) {
	key := l.getAccountMetaCacheKey(address)
	var meta types.AccountMeta

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if _, err = meta.UnmarshalMsg(val); err != nil {
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
	return &meta, nil
}

func (l *Ledger) AddOrUpdateAccountMetaCache(meta *types.AccountMeta, txns ...db.StoreTxn) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaCacheKey(meta.Address)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, metaBytes)
}

func (l *Ledger) UpdateAccountMetaCache(meta *types.AccountMeta, txns ...db.StoreTxn) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaCacheKey(meta.Address)

	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err = txn.Get(key, func(vals []byte, b byte) error {
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrAccountNotFound
		}
		return err
	}
	return txn.Set(key, metaBytes)
}

func (l *Ledger) DeleteAccountMetaCache(address types.Address, txns ...db.StoreTxn) error {
	key := l.getAccountMetaCacheKey(address)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Delete(key)
}

func (l *Ledger) HasAccountMetaCache(address types.Address, txns ...db.StoreTxn) (bool, error) {
	key := l.getAccountMetaCacheKey(address)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) error {
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
