package ledger

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi"
	"go.uber.org/zap"
)

type Ledger struct {
	io.Closer
	db     db.Store
	dir    string
	logger *zap.SugaredLogger
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
	ErrPendingExists          = errors.New("pending transaction already exists")
	ErrPendingNotFound        = errors.New("pending transaction not found")
	ErrFrontierExists         = errors.New("frontier already exists")
	ErrFrontierNotFound       = errors.New("frontier not found")
	ErrRepresentationNotFound = errors.New("representation not found")
	ErrPerformanceNotFound    = errors.New("performance not found")
	ErrPosteriorExists        = errors.New("posterior already exists")
	ErrPosteriorNotFound      = errors.New("posterior not found")
	ErrVersionNotFound        = errors.New("version not found")
	ErrStorageExists          = errors.New("storage already exists")
	ErrStorageNotFound        = errors.New("storage not found")
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
	idPrefixPosterior
	idPrefixVersion
	idPrefixStorage
	idPrefixToken
	idPrefixSender
	idPrefixReceiver
	idPrefixMessage
	idPrefixMessageInfo
	idPrefixOnlineReps
)

var (
	cache = make(map[string]*Ledger)
	lock  = sync.RWMutex{}
)

const version = 1

func NewLedger(dir string) *Ledger {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[dir]; !ok {
		store, err := db.NewBadgerStore(dir)
		if err != nil {
			fmt.Println(err.Error())
		}
		l := &Ledger{db: store, dir: dir}
		l.logger = log.NewLogger("ledger")

		if err := l.upgrade(); err != nil {
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
	err := l.db.Close()
	lock.Lock()
	delete(cache, l.dir)
	lock.Unlock()
	return err
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
		ms := []db.Migration{new(BlockPosterior)}
		return txn.Upgrade(ms)
	})
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

func getKeyOfHash(hash types.Hash, t byte) []byte {
	var key [1 + types.HashSize]byte
	key[0] = t
	copy(key[1:], hash[:])
	return key[:]
}

func getKeyOfBytes(bytes []byte, t byte) []byte {
	var key []byte
	key = append(key, t)
	key = append(key, bytes...)
	return key
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
		return err
	}
	if err := txn.Set(key, blockBytes); err != nil {
		return err
	}
	if err := addPosterior(blk, txn); err != nil {
		return err
	}
	if err := addToken(blk, txn); err != nil {
		return err
	}
	if err := addSMSDataForBlock(blk, txn); err != nil {
		return err
	}
	l.releaseTxn(txn, flag)
	return nil
}

func addPosterior(blk *types.StateBlock, txn db.StoreTxn) error {
	previous := blk.GetPrevious()
	hash := blk.GetHash()
	if !previous.IsZero() {
		bKey := getKeyOfHash(previous, idPrefixBlock)
		err := txn.Get(bKey, func(val []byte, b byte) error {
			return nil
		})
		if err == nil {
			pKey := getKeyOfHash(previous, idPrefixPosterior)
			err := txn.Get(pKey, func(val []byte, b byte) error {
				return nil
			})
			if err == nil {
				return ErrPosteriorExists
			} else if err != nil && err != badger.ErrKeyNotFound {
				return err
			}
			val := make([]byte, types.HashSize)
			err = hash.MarshalBinaryTo(val)
			if err != nil {
				return err
			}
			if err := txn.Set(pKey, val); err != nil {
				return err
			}
		}
	}
	return nil
}

func addToken(blk *types.StateBlock, txn db.StoreTxn) error {
	if blk.GetType() == types.ContractReward {
		token, err := cabi.ParseTokenInfo(blk.GetData())
		if err != nil {
			return err
		}
		tokenKey := getKeyOfHash(token.TokenId, idPrefixToken)
		err = txn.Get(tokenKey, func(bytes []byte, b byte) error {
			return nil
		})
		if err == nil {
			return errors.New("token already exist")
		} else if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		val, err := json.Marshal(token)
		if err != nil {
			return err
		}
		return txn.Set(tokenKey, val)
	}
	return nil
}

func (l *Ledger) GetStateBlock(hash types.Hash, txns ...db.StoreTxn) (*types.StateBlock, error) {
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

	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			return err
		}
		if err := fn(blk); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) DeleteStateBlock(hash types.Hash, txns ...db.StoreTxn) error {
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
	if err := deletePosterior(blk, txn); err != nil {
		return err
	}
	if err := deleteToken(blk, txn); err != nil {
		return err
	}
	if err := deleteSmsDataForBlock(blk, txn); err != nil {
		return err
	}

	l.releaseTxn(txn, flag)
	return nil
}

func deletePosterior(blk *types.StateBlock, txn db.StoreTxn) error {
	pKey := getKeyOfHash(blk.GetPrevious(), idPrefixPosterior)
	if err := txn.Delete(pKey); err != nil {
		return err
	}
	return nil
}

func deleteToken(blk *types.StateBlock, txn db.StoreTxn) error {
	if blk.GetType() == types.ContractReward {
		token, err := cabi.ParseTokenInfo(blk.GetData())
		if err != nil {
			return err
		}
		tokenKey := getKeyOfHash(token.TokenId, idPrefixToken)
		if err := txn.Delete(tokenKey); err != nil {
			return err
		}
	}
	return nil
}

func deleteSenderOrReceiver(number string, t byte, hash types.Hash, txn db.StoreTxn) error {
	if number != "" {
		key := getKeyOfBytes(util.String2Bytes(number), t)
		err := txn.Get(key, func(val []byte, b byte) error {
			hs := new([]types.Hash)
			if err := json.Unmarshal(val, hs); err != nil {
				return err
			}
			hashes := *hs
			if len(hashes) == 1 {
				return txn.Delete(key)
			}
			var hashes2 []types.Hash
			for index, h := range hashes {
				if h == hash {
					hashes2 = append(hashes[:index], hashes[index+1:]...)
					break
				}
			}
			val2, err := json.Marshal(hashes2)
			if err != nil {
				return err
			}
			return txn.Set(key, val2)
		})
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
	}
	return nil
}

func (l *Ledger) HasStateBlock(hash types.Hash, txns ...db.StoreTxn) (bool, error) {
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
	var count uint64
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})

	if err != nil {
		return 0, err
	}
	return count, nil
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
	index := rand.Int63n(int64(c))
	blk := new(types.StateBlock)
	var temp int64
	err = txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		if temp == index {
			if err = blk.Deserialize(val); err != nil {
				return err
			}
		}
		temp++
		return nil
	})

	if err != nil {
		return nil, err
	}
	return blk, nil

}

func (l *Ledger) AddSmartContractBlock(blk types.SmartContractBlock, txns ...db.StoreTxn) error {
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

	err := txn.Iterator(idPrefixSmartContractBlock, func(key []byte, val []byte, b byte) error {
		blk := new(types.SmartContractBlock)
		if err := blk.Deserialize(val); err != nil {
			return err
		}
		if err := fn(blk); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
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
	var count uint64
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixSmartContractBlock, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})

	if err != nil {
		return 0, err
	}
	return count, nil
}

func (l *Ledger) uncheckedKindToPrefix(kind types.UncheckedKind) byte {
	switch kind {
	case types.UncheckedKindPrevious:
		return idPrefixUncheckedBlockPrevious
	case types.UncheckedKindLink:
		return idPrefixUncheckedBlockLink
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

	prefix := l.uncheckedKindToPrefix(kind)
	err := txn.Iterator(prefix, func(key []byte, val []byte, b byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			return err
		}
		if err := visit(blk, kind); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (l *Ledger) WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc, txns ...db.StoreTxn) error {
	if err := l.walkUncheckedBlocks(types.UncheckedKindPrevious, visit, txns...); err != nil {
		return err
	}

	return l.walkUncheckedBlocks(types.UncheckedKindLink, visit, txns...)
}

func (l *Ledger) CountUncheckedBlocks(txns ...db.StoreTxn) (uint64, error) {
	var count uint64
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixUncheckedBlockLink, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	err = txn.Iterator(idPrefixUncheckedBlockPrevious, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (l *Ledger) getAccountMetaKey(address types.Address) []byte {
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

	key := l.getAccountMetaKey(meta.Address)
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
	key := l.getAccountMetaKey(address)
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
	var count uint64
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixAccount, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})

	if err != nil {
		return 0, err
	}
	return count, nil
}

func (l *Ledger) UpdateAccountMeta(meta *types.AccountMeta, txns ...db.StoreTxn) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaKey(meta.Address)

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
	key := l.getAccountMetaKey(meta.Address)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, metaBytes)
}

func (l *Ledger) DeleteAccountMeta(address types.Address, txns ...db.StoreTxn) error {
	key := l.getAccountMetaKey(address)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Delete(key)
}

func (l *Ledger) HasAccountMeta(address types.Address, txns ...db.StoreTxn) (bool, error) {
	key := l.getAccountMetaKey(address)
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
			return l.UpdateAccountMeta(am)
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

func (l *Ledger) getRepresentationKey(address types.Address) []byte {
	var key [1 + types.AddressSize]byte
	key[0] = idPrefixRepresentation
	copy(key[1:], address[:])
	return key[:]
}

func (l *Ledger) AddRepresentation(address types.Address, amount types.Balance, txns ...db.StoreTxn) error {
	oldAmount, err := l.GetRepresentation(address, txns...)
	if err != nil && err != ErrRepresentationNotFound {
		return err
	}
	amount = oldAmount.Add(amount)
	key := l.getRepresentationKey(address)
	amountBytes, err := amount.MarshalText()
	if err != nil {
		return err
	}
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, amountBytes)
}

func (l *Ledger) SubRepresentation(address types.Address, amount types.Balance, txns ...db.StoreTxn) error {
	oldAmount, err := l.GetRepresentation(address, txns...)
	if err != nil {
		return err
	}
	amount = oldAmount.Sub(amount)
	key := l.getRepresentationKey(address)
	amountBytes, err := amount.MarshalText()
	if err != nil {
		return err
	}
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Set(key, amountBytes)
}

func (l *Ledger) GetRepresentation(address types.Address, txns ...db.StoreTxn) (types.Balance, error) {
	key := l.getRepresentationKey(address)
	var amount types.Balance
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if err = amount.UnmarshalText(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return types.ZeroBalance, ErrRepresentationNotFound
		}
		return types.ZeroBalance, err
	}
	return amount, nil
}

func (l *Ledger) GetRepresentations(fn func(types.Address, types.Balance) error, txns ...db.StoreTxn) error {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	err := txn.Iterator(idPrefixRepresentation, func(key []byte, val []byte, b byte) error {
		var amount types.Balance
		var address types.Address
		address, err := types.BytesToAddress(key[1:])
		if err != nil {
			return err
		}
		if err := amount.UnmarshalText(val); err != nil {
			return err
		}
		if err := fn(address, amount); err != nil {
			return err
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

func (l *Ledger) AddPending(pendingKey types.PendingKey, pending *types.PendingInfo, txns ...db.StoreTxn) error {
	pendingBytes, err := pending.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getPendingKey(pendingKey)
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

func (l *Ledger) DeletePending(pendingKey types.PendingKey, txns ...db.StoreTxn) error {
	key := l.getPendingKey(pendingKey)
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
	var count uint64
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixFrontier, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (l *Ledger) GetPosterior(hash types.Hash, txns ...db.StoreTxn) (types.Hash, error) {
	key := getKeyOfHash(hash, idPrefixPosterior)
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)
	h := new(types.Hash)
	err := txn.Get(key, func(val []byte, b byte) error {
		if err := h.UnmarshalBinary(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return types.ZeroHash, ErrPosteriorNotFound
		}
		return types.ZeroHash, err
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
		txn := l.db.NewTransaction(update)
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
	txn := l.db.NewTransaction(true)
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
	balance, err := l.GetRepresentation(account, txns...)
	if err != nil {
		return types.ZeroBalance
	}
	return balance
}

// TODO: implement
func (l *Ledger) Rollback(hash types.Hash) error {
	return l.BatchUpdate(func(txn db.StoreTxn) error {
		err := l.processRollback(hash, nil, true, txn)
		if err != nil {
			l.logger.Error(err)
		}
		return err
	})
}

func (l *Ledger) processRollback(hash types.Hash, blockLink *types.StateBlock, isRoot bool, txn db.StoreTxn) error {
	tm, err := l.Token(hash, txn)
	if err != nil {
		return err
	}

	blockHead, err := l.GetStateBlock(tm.Header, txn)
	if err != nil {
		return err
	}

	blockCur := blockHead
	for {
		hashCur := blockCur.GetHash()
		//blockType, err := l.JudgeBlockKind(hashCur, txn)
		blockType := blockCur.GetType()
		if err != nil {
			return err
		}

		blockPre := new(types.StateBlock)
		if blockType != types.Open {
			blockPre, err = l.GetStateBlock(blockCur.Previous, txn)
			if err != nil {
				return err
			}
		}

		switch blockType {
		case types.Open:
			l.logger.Debug("---delete open block, ", hashCur)
			if err := l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), open(%s)", err, hashCur)
			}
			if err := l.rollBackTokenDel(tm, txn); err != nil {
				return fmt.Errorf("rollback token fail(%s), open(%s)", err, hashCur)
			}
			if err := l.rollBackFrontier(types.Hash{}, blockCur.GetHash(), txn); err != nil {
				return fmt.Errorf("rollback frontier fail(%s), open(%s)", err, hashCur)
			}
			if err := l.rollBackRep(blockCur.GetRepresentative(), blockCur.GetBalance(), false, txn); err != nil {
				return fmt.Errorf("rollback representative fail(%s), open(%s)", err, hashCur)
			}
			if err := l.rollBackPendingAdd(blockCur, blockLink, tm.Balance, blockCur.GetToken(), txn); err != nil {
				return fmt.Errorf("rollback pending fail(%s), open(%s)", err, hashCur)
			}

			if hashCur != hash || isRoot {
				if err := l.processRollback(blockCur.GetLink(), blockCur, false, txn); err != nil {
					return err
				}
			}
		case types.Send:
			l.logger.Debug("---delete send block, ", hashCur)
			if err := l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), send(%s)", err, hashCur)
			}
			if err := l.rollBackToken(tm, blockCur.GetPrevious(), tm.Representative, blockPre.GetBalance(), txn); err != nil {
				return fmt.Errorf("rollback token fail(%s), send(%s)", err, hashCur)
			}
			if err := l.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
				return fmt.Errorf("rollback frontier fail(%s), send(%s)", err, hashCur)
			}
			if err := l.rollBackRep(blockCur.GetRepresentative(), blockPre.GetBalance().Sub(blockCur.GetBalance()), true, txn); err != nil {
				return fmt.Errorf("rollback representative fail(%s), send(%s)", err, hashCur)
			}
			if err := l.rollBackPendingDel(types.Address(blockCur.Link), blockCur.GetHash(), txn); err != nil {
				return fmt.Errorf("rollback pending fail(%s), send(%s)", err, hashCur)
			}

			if hashCur != hash || isRoot {
				linkblock, err := l.getLinkBlock(blockCur, txn)
				if err != nil {
					return err
				}
				if linkblock != nil {
					if err := l.processRollback(linkblock.GetHash(), blockCur, false, txn); err != nil {
						return err
					}
				}
			}
		case types.Receive:
			l.logger.Debug("---delete receive block, ", hashCur)
			if err := l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), receive(%s)", err, hashCur)
			}
			if err := l.rollBackToken(tm, blockCur.GetPrevious(), tm.Representative, blockPre.GetBalance(), txn); err != nil {
				return fmt.Errorf("rollback token fail(%s), receive(%s)", err, hashCur)
			}
			if err := l.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
				return fmt.Errorf("rollback frontier fail(%s), receive(%s)", err, hashCur)
			}
			if err := l.rollBackRep(blockCur.GetRepresentative(), blockCur.GetBalance().Sub(blockPre.GetBalance()), false, txn); err != nil {
				return fmt.Errorf("rollback representative fail(%s), receive(%s)", err, hashCur)
			}
			if err := l.rollBackPendingAdd(blockCur, blockLink, blockCur.GetBalance().Sub(blockPre.GetBalance()), blockCur.GetToken(), txn); err != nil {
				return fmt.Errorf("rollback pending fail(%s), receive(%s)", err, hashCur)
			}

			if hashCur != hash || isRoot {
				if err := l.processRollback(blockCur.GetLink(), blockCur, false, txn); err != nil {
					return err
				}
			}
		case types.Change:
			l.logger.Debug("---delete change block, ", hashCur)
			if err := l.DeleteStateBlock(hashCur, txn); err != nil {
				return fmt.Errorf("delete state block fail(%s), change(%s)", err, hashCur)
			}
			if err := l.rollBackToken(tm, blockCur.GetPrevious(), blockCur.GetRepresentative(), tm.Balance, txn); err != nil {
				return fmt.Errorf("rollback token fail(%s), change(%s)", err, hashCur)
			}
			if err := l.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
				return fmt.Errorf("rollback frontier fail(%s), change(%s)", err, hashCur)
			}
			if err := l.rollBackRepChange(blockPre.GetRepresentative(), blockCur.GetRepresentative(), blockCur.Balance, txn); err != nil {
				return fmt.Errorf("rollback representative fail(%s), change(%s)", err, hashCur)
			}
		}

		if hashCur == hash {
			break
		}

		blockCur, err = l.GetStateBlock(blockCur.GetPrevious(), txn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) getLinkBlock(block *types.StateBlock, txn db.StoreTxn) (*types.StateBlock, error) {
	tm, err := l.GetTokenMeta(types.Address(block.GetLink()), block.GetToken(), txn)
	if err != nil {
		if err == ErrAccountNotFound || err == ErrTokenNotFound {
			return nil, nil
		}
		return nil, err
	}
	blockLink, err := l.GetStateBlock(tm.Header, txn)
	if err != nil {
		return nil, err
	}
	for {
		if blockLink.GetLink() == block.GetHash() {
			break
		}
		blockLink, err = l.GetStateBlock(blockLink.GetPrevious(), txn)
		if err != nil {
			return nil, err
		}
	}
	return blockLink, nil
}

func (l *Ledger) rollBackFrontier(pre types.Hash, cur types.Hash, txn db.StoreTxn) error {
	frontier, err := l.GetFrontier(cur, txn)
	if err != nil {
		return err
	}
	l.logger.Debug("delete frontier, ", frontier)
	if err := l.DeleteFrontier(cur, txn); err != nil {
		return err
	}
	if !pre.IsZero() {
		frontier.HeaderBlock = pre
		l.logger.Debug("add frontier, ", frontier)
		if err := l.AddFrontier(frontier, txn); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) rollBackToken(tm *types.TokenMeta, header types.Hash, rep types.Address, balance types.Balance, txn db.StoreTxn) error {
	tm.Balance = balance
	tm.Header = header
	tm.Representative = rep
	tm.BlockCount = tm.BlockCount - 1
	tm.Modified = time.Now().Unix()
	l.logger.Debug("update token, ", tm.BelongTo, tm.Type)
	if err := l.UpdateTokenMeta(tm.BelongTo, tm, txn); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) rollBackTokenDel(tm *types.TokenMeta, txn db.StoreTxn) error {
	address := tm.BelongTo
	l.logger.Debug("delete token, ", address, tm.Type)
	if err := l.DeleteTokenMeta(address, tm.Type, txn); err != nil {
		return err
	}
	ac, err := l.GetAccountMeta(address, txn)
	if err != nil {
		return err
	}
	if len(ac.Tokens) == 0 {
		if err := l.DeleteAccountMeta(address, txn); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) rollBackRep(address types.Address, balance types.Balance, isSend bool, txn db.StoreTxn) error {
	if isSend {
		l.logger.Debugf("add rep %s to %s", balance, address)
		if err := l.AddRepresentation(address, balance, txn); err != nil {
			return err
		}
	} else {
		l.logger.Debugf("sub rep %s from %s", balance, address)
		if err := l.SubRepresentation(address, balance, txn); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) rollBackRepChange(pre types.Address, cur types.Address, balance types.Balance, txn db.StoreTxn) error {
	l.logger.Debugf("add rep %s to %s", balance, pre)
	if err := l.AddRepresentation(pre, balance, txn); err != nil {
		return err
	}
	l.logger.Debugf("sub rep %s from %s", balance, cur)
	if err := l.SubRepresentation(cur, balance, txn); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) rollBackPendingAdd(blockCur *types.StateBlock, blockLink *types.StateBlock, amount types.Balance, token types.Hash, txn db.StoreTxn) error {
	if blockLink == nil {
		link, err := l.GetStateBlock(blockCur.GetLink(), txn)
		if err != nil {
			return err
		}
		blockLink = link
	}
	pendingkey := types.PendingKey{
		Address: blockCur.GetAddress(),
		Hash:    blockLink.GetHash(),
	}
	pendinginfo := types.PendingInfo{
		Source: blockLink.GetAddress(),
		Amount: amount,
		Type:   token,
	}
	l.logger.Debug("add pending, ", pendingkey, pendinginfo)
	if err := l.AddPending(pendingkey, &pendinginfo, txn); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) rollBackPendingDel(address types.Address, hash types.Hash, txn db.StoreTxn) error {
	pendingkey := types.PendingKey{
		Address: address,
		Hash:    hash,
	}
	l.logger.Debug("delete pending ,", pendingkey)
	if err := l.DeletePending(pendingkey, txn); err != nil {
		return err
	}
	return nil
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
		return block.GetBalance(), err
	case types.Send:
		if prev, err = l.GetStateBlock(block.Previous); err != nil {
			return types.ZeroBalance, err
		}
		return prev.Balance.Sub(block.GetBalance()), nil
	case types.Receive:
		if prev, err = l.GetStateBlock(block.Previous); err != nil {
			return types.ZeroBalance, err
		}
		return block.GetBalance().Sub(prev.Balance), nil
	case types.Change:
		return types.ZeroBalance, nil
	case types.ContractReward:
		return block.GetBalance(), nil
	case types.ContractSend:
		return types.ZeroBalance, nil
	default:
		return types.ZeroBalance, errors.New("invalid block type")
	}
}

func (l *Ledger) ListTokens(txns ...db.StoreTxn) ([]*types.TokenInfo, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var tokens []*types.TokenInfo
	err := txn.Iterator(idPrefixToken, func(key []byte, val []byte, b byte) error {
		token := new(types.TokenInfo)
		if err := json.Unmarshal(val, token); err != nil {
			fmt.Println(err)
			return err
		}
		tokens = append(tokens, token)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (l *Ledger) GetTokenById(tokenId types.Hash, txns ...db.StoreTxn) (*types.TokenInfo, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	key := getKeyOfHash(tokenId, idPrefixToken)
	token := new(types.TokenInfo)
	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if err := json.Unmarshal(val, token); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	return token, nil
}

func (l *Ledger) GetTokenByName(tokenName string, txns ...db.StoreTxn) (*types.TokenInfo, error) {
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var token *types.TokenInfo
	err := txn.Iterator(idPrefixToken, func(key []byte, val []byte, b byte) error {
		t := new(types.TokenInfo)
		if err := json.Unmarshal(val, t); err != nil {
			return err
		}
		if t.TokenName == tokenName {
			token = t
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
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
	////cache to db
	//_ = s.setWork(hash, work)
}

func (l *Ledger) GenerateSendBlock(from, to types.Address, token types.Hash, amount types.Balance, sender, receiver, message string, prk ed25519.PrivateKey) (*types.StateBlock, error) {
	tm, err := l.GetTokenMeta(from, token)
	if err != nil {
		return nil, err
	}
	//balance, err := l.TokenBalance(source, token)
	//if err != nil {
	//	return nil, err
	//}
	mHash, err := l.addMessageInfo(message)
	if err != nil {
		return nil, err
	}

	if tm.Balance.Compare(amount) != types.BalanceCompSmaller {
		sb := types.StateBlock{
			Type:           types.Send,
			Address:        from,
			Token:          token,
			Link:           to.ToHash(),
			Balance:        tm.Balance.Sub(amount),
			Previous:       tm.Header,
			Representative: tm.Representative,
			Sender:         sender,
			Receiver:       receiver,
			Message:        mHash,
		}
		acc := types.NewAccount(prk)
		sb.Signature = acc.Sign(sb.GetHash())
		sb.Work = l.generateWork(sb.Root())
		return &sb, nil
	} else {
		return nil, fmt.Errorf("not enought balance(%s) of %s", tm.Balance, amount)
	}
}

func (l *Ledger) GenerateReceiveBlock(sendBlock *types.StateBlock, prk ed25519.PrivateKey) (*types.StateBlock, error) {
	hash := sendBlock.GetHash()
	if !sendBlock.GetType().Equal(types.Send) {
		return nil, fmt.Errorf("(%s) is not send block", hash.String())
	}
	if exist, err := l.HasStateBlock(hash); !exist || err != nil {
		return nil, fmt.Errorf("send block(%s) does not exist", hash.String())
	}
	acc := types.NewAccount(prk)
	rxAccount := types.Address(sendBlock.Link)
	info, err := l.GetPending(types.PendingKey{Address: rxAccount, Hash: hash})
	if err != nil {
		return nil, err
	}
	has, err := l.HasAccountMeta(rxAccount)
	if err != nil {
		return nil, err
	}
	if has {
		rxAm, err := l.GetAccountMeta(rxAccount)
		if err != nil {
			return nil, err
		}
		rxTm := rxAm.Token(sendBlock.GetToken())
		sb := types.StateBlock{
			Type:           types.Receive,
			Address:        rxAccount,
			Balance:        rxTm.Balance.Add(info.Amount),
			Previous:       rxTm.Header,
			Link:           hash,
			Representative: rxTm.Representative,
			Token:          rxTm.Type,
			Extra:          types.ZeroHash,
		}
		sb.Signature = acc.Sign(sb.GetHash())
		sb.Work = l.generateWork(sb.Root())
		return &sb, nil
	} else {
		//genesis, err := mock.GetTokenById(mock.GetChainTokenType())
		//if err != nil {
		//	return nil, err
		//}
		sb := &types.StateBlock{
			Type:           types.Open,
			Address:        rxAccount,
			Balance:        info.Amount,
			Previous:       types.ZeroHash,
			Link:           hash,
			Representative: sendBlock.GetRepresentative(), //Representative: genesis.Owner,
			Token:          sendBlock.GetToken(),
			Extra:          types.ZeroHash,
		}
		sb.Signature = acc.Sign(sb.GetHash())
		sb.Work = l.generateWork(sb.Root())
		return sb, nil
	}
}

func (l *Ledger) GenerateChangeBlock(account types.Address, representative types.Address, prk ed25519.PrivateKey) (*types.StateBlock, error) {
	if b, err := l.HasAccountMeta(account); err != nil || !b {
		return nil, fmt.Errorf("account[%s] is not exist", account.String())
	}

	if _, err := l.GetAccountMeta(representative); err != nil {
		return nil, fmt.Errorf("invalid representative[%s]", representative.String())
	}

	//get latest chain token block
	hash := l.Latest(account, common.QLCChainToken)
	if hash.IsZero() {
		return nil, fmt.Errorf("account [%s] does not have the main chain account", account.String())
	}

	block, err := l.GetStateBlock(hash)
	if err != nil {
		return nil, err
	}

	tm, err := l.GetTokenMeta(account, common.QLCChainToken)
	sb := types.StateBlock{
		Type:           types.Change,
		Address:        account,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Link:           types.ZeroHash,
		Representative: representative,
		Token:          block.Token,
		Extra:          types.ZeroHash,
	}
	acc := types.NewAccount(prk)
	sb.Signature = acc.Sign(sb.GetHash())
	sb.Work = l.generateWork(sb.Root())
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