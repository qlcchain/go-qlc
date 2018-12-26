package ledger

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
)

type Ledger struct {
	io.Closer
	db  db.Store
	dir string
}

var logger = log.NewLogger("ledger")

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
)

const (
	idPrefixBlock byte = iota
	idPrefixUncheckedBlockPrevious
	idPrefixUncheckedBlockLink
	idPrefixAccount
	//idPrefixToken
	idPrefixFrontier
	idPrefixPending
	idPrefixRepresentation
)

var (
	cache = make(map[string]*Ledger)
	lock  = sync.RWMutex{}
)

func NewLedger(dir string) *Ledger {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := cache[dir]; !ok {
		store, err := db.NewBadgerStore(dir)
		if err != nil {
			logger.Fatal(err.Error())
		}

		cache[dir] = &Ledger{db: store, dir: dir}
	}
	return cache[dir]
}

//CloseLedger force release all ledger instance
func CloseLedger() {
	for k, v := range cache {
		if v != nil {
			v.Close()
			logger.Debugf("release ledger from %s", k)
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

func (l *Ledger) getBlockKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixBlock
	copy(key[1:], hash[:])
	return key[:]
}

// -------------------  Block  --------------------

func (l *Ledger) block2badger(blk types.Block) ([]byte, []byte, error) {
	hash := blk.GetHash()
	blockBytes, err := blk.MarshalMsg(nil)

	key := l.getBlockKey(hash)

	return key[:], blockBytes, err
}

func (l *Ledger) AddBlock(blk types.Block, txns ...db.StoreTxn) error {
	key, val, err := l.block2badger(blk)
	if err != nil {
		return err
	}

	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	//never overwrite implicitly
	err = txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return txn.SetWithMeta(key, val, byte(blk.GetType()))
}

func (l *Ledger) GetBlock(hash types.Hash, txns ...db.StoreTxn) (types.Block, error) {
	key := l.getBlockKey(hash)
	var blk types.Block
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if blk, err = types.NewBlock(types.BlockType(b)); err != nil {
			return err
		}
		if _, err = blk.UnmarshalMsg(val); err != nil {
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

func (l *Ledger) GetBlocks(txns ...db.StoreTxn) ([]*types.Block, error) {
	var blocks []*types.Block
	//var blk types.Block
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) (err error) {
		blk, err := types.NewBlock(types.BlockType(b))
		if err != nil {
			return
		}
		_, err = blk.UnmarshalMsg(val)
		blocks = append(blocks, &blk)
		return
	})

	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (l *Ledger) DeleteBlock(hash types.Hash, txns ...db.StoreTxn) error {
	key := l.getBlockKey(hash)
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	return txn.Delete(key)
}

func (l *Ledger) HasBlock(hash types.Hash, txns ...db.StoreTxn) (bool, error) {
	key := l.getBlockKey(hash)
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

func (l *Ledger) CountBlocks(txns ...db.StoreTxn) (uint64, error) {
	var count uint64
	txn, flag := l.getTxn(true, txns...)
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

func (l *Ledger) GetRandomBlock(txns ...db.StoreTxn) (types.Block, error) {
	c, err := l.CountBlocks()
	if err != nil {
		return nil, err
	}
	if c == 0 {
		return nil, ErrStoreEmpty
	}
	index := rand.Int63n(int64(c))
	var blk types.Block
	var temp int64
	txn, flag := l.getTxn(true, txns...)
	defer l.releaseTxn(txn, flag)

	err = txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		if temp == index {
			blk, err = types.NewBlock(types.BlockType(b))
			if err != nil {
				return err
			}
			_, err = blk.UnmarshalMsg(val)
			if err != nil {
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

// ------------------- unchecked Block  --------------------

func (l *Ledger) getUncheckedBlockKey(hash types.Hash, kind types.UncheckedKind) []byte {
	var key [1 + types.HashSize]byte
	key[0] = l.uncheckedKindToPrefix(kind)
	copy(key[1:], hash[:])
	return key[:]
}

func (l *Ledger) AddUncheckedBlock(parentHash types.Hash, blk types.Block, kind types.UncheckedKind, txns ...db.StoreTxn) error {
	blockBytes, err := blk.MarshalMsg(nil)
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

	return txn.SetWithMeta(key, blockBytes, byte(blk.GetType()))
}

func (l *Ledger) GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txns ...db.StoreTxn) (types.Block, error) {
	key := l.getUncheckedBlockKey(parentHash, kind)

	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	var blk types.Block
	err := txn.Get(key, func(val []byte, b byte) (err error) {
		if blk, err = types.NewBlock(types.BlockType(b)); err != nil {
			return err
		}
		if _, err = blk.UnmarshalMsg(val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrUncheckedBlockNotFound
		}
		return nil, err
	}
	return blk, nil
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

func (l *Ledger) WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc, txns ...db.StoreTxn) error {
	return nil
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

// ------------------- AccountMeta  --------------------

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

// ------------------- TokenMeta  --------------------

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

// ------------------- representation  --------------------

func (l *Ledger) AddRepresentation(address types.Address, amount types.Balance, txns ...db.StoreTxn) error {
	oldAmount, err := l.GetRepresentation(address, txns...)
	if err != nil && err != badger.ErrKeyNotFound {
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
		return types.ZeroBalance, err
	}
	return amount, nil
}

func (l *Ledger) getPendingKey(pendingKey types.PendingKey) []byte {
	key := []byte{idPrefixPending}
	_, _ = pendingKey.MarshalMsg(key[1:])
	return key[:]
}

// ------------------- pending  --------------------

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

func (l *Ledger) getFrontierKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], hash[:])
	return key[:]
}

// ------------------- frontier  --------------------

func (l *Ledger) AddFrontier(frontier *types.Frontier, txns ...db.StoreTxn) error {
	key := l.getFrontierKey(frontier.HeaderBlock)
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
	key := l.getFrontierKey(hash)
	frontier := types.Frontier{HeaderBlock: hash}
	txn, flag := l.getTxn(true, txns...)
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
	txn, flag := l.getTxn(true, txns...)
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
	return frontiers, nil
}

func (l *Ledger) DeleteFrontier(hash types.Hash, txns ...db.StoreTxn) error {
	key := l.getFrontierKey(hash)
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
			logger.Error(err)
		}
		txn.Discard()
	}
}

// BatchUpdate MUST pass the same txn
func (l *Ledger) BatchUpdate(fn func(txn db.StoreTxn) error) error {
	txn := l.db.NewTransaction(true)
	logger.Debugf("BatchUpdate NewTransaction %p", txn)
	defer func() {
		logger.Debugf("BatchUpdate Discard %p", txn)
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
	block, err := l.GetBlock(hash, txns...)
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
	block, err := l.GetBlock(hash, txns...)
	if err != nil {
		return nil, err
	}
	if b, ok := block.(*types.StateBlock); ok {
		token := b.GetToken()
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
	return nil, fmt.Errorf("invalid block by hash(%s)", hash)
}

func (l *Ledger) Pending(account types.Address, txns ...db.StoreTxn) ([]*types.PendingKey, error) {
	var cache []*types.PendingKey
	txn, flag := l.getTxn(false, txns...)
	defer l.releaseTxn(txn, flag)

	err := txn.Iterator(idPrefixPending, func(key []byte, val []byte, b byte) error {
		pendingKey := types.PendingKey{}
		_, err := pendingKey.UnmarshalMsg(key)
		if err != nil {
			return nil
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
		_, err := pendingKey.UnmarshalMsg(key)
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

const (
	receive byte = iota
	send
	open
	change
	unknown
)

// TODO: implement
func (l *Ledger) Rollback(hash types.Hash) error {
	return l.BatchUpdate(func(txn db.StoreTxn) error {
		return l.processRollback(hash, true, txn)
	})
}

func (l *Ledger) processRollback(hash types.Hash, isRoot bool, txn db.StoreTxn) error {
	tm, err := l.Token(hash, txn)
	if err != nil {
		return err
	}

	blockHead, err := l.getStateBlock(tm.Header, txn)
	if err != nil {
		return err
	}

	blockCur := blockHead
	for {
		hashCur := blockCur.GetHash()
		blockType, err := l.getBlockKind(hashCur, txn)
		if err != nil {
			return err
		}

		blockPre := new(types.StateBlock)
		if blockType != open {
			blockPre, err = l.getStateBlock(blockCur.Previous, txn)
			if err != nil {
				return err
			}
		}

		switch blockType {
		case open:
			logger.Info("---delete open block, ", hashCur)
			if err := l.DeleteBlock(hashCur, txn); err != nil {
				return err
			}
			if err := l.rollBackTokenDel(tm, txn); err != nil {
				return err
			}
			if err := l.rollBackFrontier(types.Hash{}, blockCur.GetHash(), txn); err != nil {
				return err
			}
			if err := l.rollBackRep(blockCur.GetRepresentative(), blockCur.GetBalance(), false, txn); err != nil {
				return err
			}

			if hashCur != hash || isRoot {
				if err := l.processRollback(blockCur.GetLink(), false, txn); err != nil {
					return err
				}
			}
		case send:
			logger.Info("---delete send block, ", hashCur)
			if err := l.DeleteBlock(hashCur, txn); err != nil {
				return err
			}
			if err := l.rollBackToken(tm, blockCur.GetPrevious(), tm.RepBlock, blockPre.GetBalance(), txn); err != nil {
				return err
			}
			if err := l.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
				return err
			}
			if err := l.rollBackRep(blockCur.GetRepresentative(), blockPre.GetBalance().Sub(blockCur.GetBalance()), true, txn); err != nil {
				return err
			}

			if hashCur != hash || isRoot {
				linkblock, err := l.getLinkBlock(blockCur, txn)
				if err != nil {
					return err
				}
				if err := l.processRollback(linkblock.GetHash(), false, txn); err != nil {
					return err
				}
			}
		case receive:
			logger.Info("---delete receive block, ", hashCur)
			if err := l.DeleteBlock(hashCur, txn); err != nil {
				return err
			}
			if err := l.rollBackToken(tm, blockCur.GetPrevious(), tm.RepBlock, blockPre.GetBalance(), txn); err != nil {
				return err
			}
			if err := l.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
				return err
			}
			if err := l.rollBackRep(blockCur.GetRepresentative(), blockCur.GetBalance().Sub(blockPre.GetBalance()), false, txn); err != nil {
				return err
			}

			if hashCur != hash || isRoot {
				if err := l.processRollback(blockCur.GetLink(), false, txn); err != nil {
					return err
				}
			}
		case change:
			logger.Info("---delete change block, ", hashCur)
			if err := l.DeleteBlock(hashCur, txn); err != nil {
				return err
			}
			if err := l.rollBackToken(tm, blockCur.GetPrevious(), blockCur.GetPrevious(), tm.Balance, txn); err != nil {
				return err
			}
			if err := l.rollBackFrontier(blockPre.GetHash(), blockCur.GetHash(), txn); err != nil {
				return err
			}
			if err := l.rollBackRepChange(blockPre.GetRepresentative(), blockCur.GetRepresentative(), blockCur.Balance, txn); err != nil {
				return err
			}
		}

		if hashCur == hash {
			break
		}

		blockCur, err = l.getStateBlock(blockCur.GetPrevious(), txn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) getBlockKind(hash types.Hash, txn db.StoreTxn) (byte, error) {
	block, err := l.getStateBlock(hash, txn)
	if err != nil {
		return unknown, err
	}

	pre := block.GetPrevious()
	if pre.IsZero() {
		return open, nil
	} else {
		blockPre, err := l.getStateBlock(pre, txn)
		if err != nil {
			return unknown, err
		}
		if blockPre.GetBalance().Compare(block.GetBalance()) == types.BalanceCompSmaller {
			return receive, nil
		} else {
			link := block.GetLink()
			if link.IsZero() {
				return change, nil
			}
			return send, nil
		}
	}
}

func (l *Ledger) getStateBlock(hash types.Hash, txn db.StoreTxn) (*types.StateBlock, error) {
	b, err := l.GetBlock(hash, txn)
	if err != nil {
		return nil, err
	}
	bs, ok := b.(*types.StateBlock)
	if !ok {
		return nil, errors.New("invalid block")
	}
	return bs, nil
}

func (l *Ledger) getLinkBlock(block *types.StateBlock, txn db.StoreTxn) (*types.StateBlock, error) {
	tm, err := l.GetTokenMeta(types.Address(block.GetLink()), block.GetToken(), txn)
	if err != nil {
		return nil, err
	}
	blockLink, err := l.getStateBlock(tm.Header, txn)
	if err != nil {
		return nil, err
	}
	for {
		if blockLink.GetLink() == block.GetHash() {
			break
		}
		blockLink, err = l.getStateBlock(blockLink.GetPrevious(), txn)
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
	logger.Info("delete frontier, ", frontier)
	if err := l.DeleteFrontier(cur, txn); err != nil {
		return err
	}
	if !pre.IsZero() {
		frontier.HeaderBlock = pre
		logger.Info("add frontier, ", frontier)
		if err := l.AddFrontier(frontier, txn); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) rollBackToken(tm *types.TokenMeta, header types.Hash, rep types.Hash, balance types.Balance, txn db.StoreTxn) error {
	tm.Balance = balance
	tm.Header = header
	tm.RepBlock = rep
	tm.BlockCount = tm.BlockCount - 1
	tm.Modified = time.Now().Unix()
	logger.Info("update token, ", tm.BelongTo, tm.Type)
	if err := l.UpdateTokenMeta(tm.BelongTo, tm, txn); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) rollBackTokenDel(tm *types.TokenMeta, txn db.StoreTxn) error {
	address := tm.BelongTo
	logger.Info("delete token, ", address, tm.Type)
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
		logger.Infof("add rep %s to %s", balance, address)
		if err := l.AddRepresentation(address, balance, txn); err != nil {
			return err
		}
	} else {
		logger.Infof("sub rep %s from %s", balance, address)
		if err := l.SubRepresentation(address, balance, txn); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) rollBackRepChange(pre types.Address, cur types.Address, balance types.Balance, txn db.StoreTxn) error {
	logger.Infof("add rep %s to %s", balance, pre)
	if err := l.AddRepresentation(pre, balance, txn); err != nil {
		return err
	}
	logger.Infof("sub rep %s from %s", balance, cur)
	if err := l.SubRepresentation(cur, balance, txn); err != nil {
		return err
	}
	return nil
}
