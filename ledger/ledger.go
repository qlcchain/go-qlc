package ledger

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/log"
	"io"
	"math/rand"
	"sync"
)

type Ledger struct {
	io.Closer
	db  db.Store
	dir string
}

//LedgerSession
type LedgerSession struct {
	db.Store
	io.Closer

	txn   db.StoreTxn
	reuse bool
	mode  bool
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
	lock.RLock()
	defer lock.RUnlock()
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
	lock.RLock()
	defer lock.RUnlock()
	for k, v := range cache {
		if v != nil {
			v.Close()
			logger.Debugf("release ledger from %s", k)
		}
		delete(cache, k)
	}
}

func (l *Ledger) Close() error {
	lock.RLock()
	defer lock.RUnlock()
	err := l.db.Close()
	delete(cache, l.dir)
	return err
}

func (l *Ledger) NewLedgerSession(reuse bool) *LedgerSession {
	return &LedgerSession{Store: l.db, reuse: reuse}
}

func (ls *LedgerSession) Close() error {
	if ls.txn != nil {
		ls.txn.Discard()
		logger.Debugf("close txn session %p", ls.txn)
		ls.txn = nil
	}
	return nil
}

// Empty reports whether the database is empty or not.
func (ls *LedgerSession) Empty() (bool, error) {
	r := true
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		r = false
		return nil
	})
	if err != nil {
		return r, err
	}
	return r, nil
}

func (ls *LedgerSession) getBlockKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixBlock
	copy(key[1:], hash[:])
	return key[:]
}

// -------------------  Block  --------------------

func (ls *LedgerSession) block2badger(blk types.Block) ([]byte, []byte, error) {
	hash := blk.GetHash()
	blockBytes, err := blk.MarshalMsg(nil)

	key := ls.getBlockKey(hash)

	return key[:], blockBytes, err
}

func (ls *LedgerSession) AddBlock(blk types.Block) error {
	key, val, err := ls.block2badger(blk)
	if err != nil {
		return err
	}
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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

func (ls *LedgerSession) GetBlock(hash types.Hash) (types.Block, error) {
	key := ls.getBlockKey(hash)
	var blk types.Block
	txn := ls.getTxn(false)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) GetBlocks() ([]*types.Block, error) {
	var blocks []*types.Block
	//var blk types.Block
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) DeleteBlock(hash types.Hash) error {
	key := ls.getBlockKey(hash)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()
	return txn.Delete(key)

}
func (ls *LedgerSession) HasBlock(hash types.Hash) (bool, error) {
	key := ls.getBlockKey(hash)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) CountBlocks() (uint64, error) {
	var count uint64
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (ls *LedgerSession) GetRandomBlock() (types.Block, error) {
	c, err := ls.CountBlocks()
	if err != nil {
		return nil, err
	}
	if c == 0 {
		return nil, ErrStoreEmpty
	}
	index := rand.Int63n(int64(c))
	var blk types.Block
	var temp int64
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) uncheckedKindToPrefix(kind types.UncheckedKind) byte {
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

func (ls *LedgerSession) getUncheckedBlockKey(hash types.Hash, kind types.UncheckedKind) []byte {
	var key [1 + types.HashSize]byte
	key[0] = ls.uncheckedKindToPrefix(kind)
	copy(key[1:], hash[:])
	return key[:]
}
func (ls *LedgerSession) AddUncheckedBlock(parentHash types.Hash, blk types.Block, kind types.UncheckedKind) error {
	blockBytes, err := blk.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := ls.getUncheckedBlockKey(parentHash, kind)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) (types.Block, error) {
	key := ls.getUncheckedBlockKey(parentHash, kind)

	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) error {
	key := ls.getUncheckedBlockKey(parentHash, kind)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	return txn.Delete(key)
}
func (ls *LedgerSession) HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind) (bool, error) {
	key := ls.getUncheckedBlockKey(hash, kind)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc) error {
	return nil
}
func (ls *LedgerSession) CountUncheckedBlocks() (uint64, error) {
	var count uint64
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) getAccountMetaKey(address types.Address) []byte {
	var key [1 + types.AddressSize]byte
	key[0] = idPrefixAccount
	copy(key[1:], address[:])
	return key[:]
}

// ------------------- AccountMeta  --------------------

func (ls *LedgerSession) AddAccountMeta(meta *types.AccountMeta) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := ls.getAccountMetaKey(meta.Address)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) GetAccountMeta(address types.Address) (*types.AccountMeta, error) {
	key := ls.getAccountMetaKey(address)
	var meta types.AccountMeta

	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) UpdateAccountMeta(meta *types.AccountMeta) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := ls.getAccountMetaKey(meta.Address)

	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) AddOrUpdateAccountMeta(meta *types.AccountMeta) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := ls.getAccountMetaKey(meta.Address)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	return txn.Set(key, metaBytes)
}
func (ls *LedgerSession) DeleteAccountMeta(address types.Address) error {
	key := ls.getAccountMetaKey(address)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	return txn.Delete(key)
}
func (ls *LedgerSession) HasAccountMeta(address types.Address) (bool, error) {
	key := ls.getAccountMetaKey(address)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) AddTokenMeta(address types.Address, meta *types.TokenMeta) error {
	am, err := ls.GetAccountMeta(address)
	if err != nil {
		return err
	}
	for _, t := range am.Tokens {
		if t.Type == meta.Type {
			return ErrTokenExists
		}
	}

	am.Tokens = append(am.Tokens, meta)
	return ls.UpdateAccountMeta(am)
}

// ------------------- TokenMeta  --------------------

func (ls *LedgerSession) GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error) {
	am, err := ls.GetAccountMeta(address)
	if err != nil {
		return nil, err
	}
	for _, token := range am.Tokens {
		if token.Type == tokenType {
			return token, nil
		}
	}
	return nil, ErrTokenNotFound
}
func (ls *LedgerSession) UpdateTokenMeta(address types.Address, meta *types.TokenMeta) error {
	am, err := ls.GetAccountMeta(address)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range am.Tokens {
		if token.Type == meta.Type {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
			am.Tokens = append(am.Tokens, meta)
			return ls.UpdateAccountMeta(am)
		}
	}
	return ErrTokenNotFound
}
func (ls *LedgerSession) AddOrUpdateTokenMeta(address types.Address, meta *types.TokenMeta) error {
	am, err := ls.GetAccountMeta(address)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range am.Tokens {
		if token.Type == meta.Type {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
			am.Tokens = append(am.Tokens, meta)
			return ls.UpdateAccountMeta(am)
		}
	}

	am.Tokens = append(am.Tokens, meta)
	return ls.UpdateAccountMeta(am)
}
func (ls *LedgerSession) DeleteTokenMeta(address types.Address, tokenType types.Hash) error {
	am, err := ls.GetAccountMeta(address)
	if err != nil {
		return err
	}
	tokens := am.Tokens
	for index, token := range tokens {
		if token.Type == tokenType {
			am.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}
	return ls.UpdateAccountMeta(am)
}
func (ls *LedgerSession) HasTokenMeta(address types.Address, tokenType types.Hash) (bool, error) {
	am, err := ls.GetAccountMeta(address)
	if err != nil {
		return false, err
	}
	for _, t := range am.Tokens {
		if t.Type == tokenType {
			return true, nil
		}
	}
	return false, nil
}
func (ls *LedgerSession) getRepresentationKey(address types.Address) []byte {
	var key [1 + types.AddressSize]byte
	key[0] = idPrefixRepresentation
	copy(key[1:], address[:])
	return key[:]
}

// ------------------- representation  --------------------

func (ls *LedgerSession) AddRepresentation(address types.Address, amount types.Balance) error {
	oldAmount, err := ls.GetRepresentation(address)
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	amount = oldAmount.Add(amount)
	key := ls.getRepresentationKey(address)
	amountBytes, err := amount.MarshalText()
	if err != nil {
		return err
	}
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	return txn.Set(key, amountBytes)

}
func (ls *LedgerSession) SubRepresentation(address types.Address, amount types.Balance) error {
	oldAmount, err := ls.GetRepresentation(address)
	if err != nil {
		return err
	}
	amount = oldAmount.Sub(amount)
	key := ls.getRepresentationKey(address)
	amountBytes, err := amount.MarshalText()
	if err != nil {
		return err
	}
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	return txn.Set(key, amountBytes)
}
func (ls *LedgerSession) GetRepresentation(address types.Address) (types.Balance, error) {
	key := ls.getRepresentationKey(address)
	var amount types.Balance
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) getPendingKey(pendingKey types.PendingKey) []byte {
	key := []byte{idPrefixPending}
	_, _ = pendingKey.MarshalMsg(key[1:])
	return key[:]
}

// ------------------- pending  --------------------

func (ls *LedgerSession) AddPending(pendingKey types.PendingKey, pending *types.PendingInfo) error {
	pendingBytes, err := pending.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := ls.getPendingKey(pendingKey)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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

func (ls *LedgerSession) GetPending(pendingKey types.PendingKey) (*types.PendingInfo, error) {
	key := ls.getPendingKey(pendingKey)
	var pending types.PendingInfo
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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

func (ls *LedgerSession) DeletePending(pendingKey types.PendingKey) error {
	key := ls.getPendingKey(pendingKey)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	return txn.Delete(key)
}

func (ls *LedgerSession) getFrontierKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], hash[:])
	return key[:]
}

// ------------------- frontier  --------------------

func (ls *LedgerSession) AddFrontier(frontier *types.Frontier) error {
	key := ls.getFrontierKey(frontier.HeaderBlock)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) GetFrontier(hash types.Hash) (*types.Frontier, error) {
	key := ls.getFrontierKey(hash)
	frontier := types.Frontier{HeaderBlock: hash}
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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

func (ls *LedgerSession) GetFrontiers() ([]*types.Frontier, error) {
	var frontiers []*types.Frontier
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) DeleteFrontier(hash types.Hash) error {
	key := ls.getFrontierKey(hash)
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	return txn.Delete(key)
}
func (ls *LedgerSession) CountFrontiers() (uint64, error) {
	var count uint64
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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
func (ls *LedgerSession) getTxn(update bool) db.StoreTxn {
	if ls.reuse {
		if update != ls.mode || ls.txn == nil {
			ls.Close()
			ls.txn = ls.NewTransaction(update)
		}
		ls.mode = update
	} else {
		ls.Close()
		ls.txn = ls.NewTransaction(update)
	}
	//logger.Debugf("get txn: %p, mode: %s", ls.txn, strconv.FormatBool(update))
	return ls.txn
}

// releaseTxn commit change and close txn
func (ls *LedgerSession) releaseTxn() {
	if !ls.reuse {
		err := ls.txn.Commit(nil)
		if err != nil {
			logger.Fatal(err)
		}
		//logger.Debugf("releaseTxn and commit %p", ls.txn)
		ls.Close()
	}
}

func (ls *LedgerSession) BatchUpdate(fn func() error) error {
	if !ls.reuse {
		return errors.New("batch update should enable reuse transaction")
	}
	ls.txn = ls.NewTransaction(true)
	logger.Debugf("BatchUpdate NewTransaction %p", ls.txn)
	defer func() {
		logger.Debugf("BatchUpdate Discard %p", ls.txn)
		ls.txn.Discard()
	}()

	if err := fn(); err != nil {
		return err
	}
	return ls.txn.Commit(nil)
}

func (ls *LedgerSession) Latest(account types.Address, token types.Hash) types.Hash {
	zero := types.Hash{}
	am, err := ls.GetAccountMeta(account)
	if err != nil {
		return zero
	}

	for _, t := range am.Tokens {
		if t.Type == token {
			return t.Header
		}
	}
	return zero
}

func (ls *LedgerSession) Account(hash types.Hash) (*types.AccountMeta, error) {
	block, err := ls.GetBlock(hash)
	if err != nil {
		return nil, err
	}
	addr := block.GetAddress()
	am, err := ls.GetAccountMeta(addr)
	if err != nil {
		return nil, err
	}

	return am, nil
}

func (ls *LedgerSession) Token(hash types.Hash) (*types.TokenMeta, error) {
	block, err := ls.GetBlock(hash)
	if err != nil {
		return nil, err
	}
	if b, ok := block.(*types.StateBlock); ok {
		token := b.GetToken()
		addr := block.GetAddress()
		am, err := ls.GetAccountMeta(addr)
		if err != nil {
			return nil, err
		}

		for _, tm := range am.Tokens {
			if tm.Type == token {
				return tm, nil
			}
		}
		//TODO: hash to token name
		return nil, fmt.Errorf("can not find token %s", token)
	}
	return nil, fmt.Errorf("invalid block by hash(%s)", hash)
}

func (ls *LedgerSession) Pending(account types.Address) ([]*types.PendingKey, error) {
	var cache []*types.PendingKey
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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

func (ls *LedgerSession) Balance(account types.Address) (map[types.Hash]types.Balance, error) {
	cache := make(map[types.Hash]types.Balance)
	am, err := ls.GetAccountMeta(account)
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

func (ls *LedgerSession) TokenPending(account types.Address, token types.Hash) ([]*types.PendingKey, error) {
	var cache []*types.PendingKey
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

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

func (ls *LedgerSession) TokenBalance(account types.Address, token types.Hash) (types.Balance, error) {
	am, err := ls.GetAccountMeta(account)
	if err != nil {
		return types.ZeroBalance, err
	}
	for _, tm := range am.Tokens {
		if tm.Type == token {
			return tm.Balance, nil
		}
	}

	return types.ZeroBalance, fmt.Errorf("can not find %s balance", token)
}

func (ls *LedgerSession) Weight(account types.Address) types.Balance {
	balance, err := ls.GetRepresentation(account)
	if err != nil {
		return types.ZeroBalance
	}
	return balance
}

// TODO: implement
func (ls *LedgerSession) Rollback(hash types.Hash) error {
	return ls.BatchUpdate(func() error {
		//tm, err := ls.Token(hash)
		//if err != nil {
		//	return err
		//}
		//block, err := ls.GetBlock(tm.Header)
		//if err != nil {
		//	return nil
		//}

		return nil
	})
}
