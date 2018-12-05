package ledger

import (
	"errors"
	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
	"io"
	"math/rand"
	"strconv"
	"sync"
)

type Ledger struct {
	io.Closer
	db db.Store
}

//LedgerSession
type LedgerSession struct {
	db.Store
	io.Closer
	txn   db.StoreTxn
	reuse bool
	mode  bool
}

var log = common.NewLogger("ledger")

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
	ledger *Ledger
	once   sync.Once
)

func NewLedger() *Ledger {
	once.Do(func() {
		dir := util.QlcDir("ledger")
		store, err := db.NewBadgerStore(dir)
		if err != nil {
			log.Fatal(err.Error())
		}
		ledger = &Ledger{db: store}
	})
	return ledger
}

func (l *Ledger) Close() error {
	return l.db.Close()
}

func (l *Ledger) NewLedgerSession(reuse bool) *LedgerSession {
	return &LedgerSession{Store: l.db, reuse: reuse}
}

func (ls *LedgerSession) Close() error {
	if ls.txn != nil {
		ls.txn.Discard()
		log.Debugf("close txn session %p", ls.txn)
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
		if blk, err = types.NewBlock(b); err != nil {
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
func (ls *LedgerSession) GetBlocks() ([]types.Block, error) {
	var blocks []types.Block
	//var blk types.Block
	txn := ls.getTxn(true)
	defer ls.releaseTxn()

	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) (err error) {
		blk, err := types.NewBlock(b)
		if err != nil {
			return
		}
		_, err = blk.UnmarshalMsg(val)
		blocks = append(blocks, blk)
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
			blk, err = types.NewBlock(b)
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
		if blk, err = types.NewBlock(b); err != nil {
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
func (ls *LedgerSession) getPendingKey(destination types.Address, hash types.Hash) []byte {
	var key [1 + types.PendingKeySize]byte
	key[0] = idPrefixPending
	copy(key[1:], destination[:])
	copy(key[1+types.AddressSize:], hash[:])
	return key[:]
}

// ------------------- pending  --------------------

func (ls *LedgerSession) AddPending(destination types.Address, hash types.Hash, pending *types.PendingInfo) error {
	pendingBytes, err := pending.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := ls.getPendingKey(destination, hash)
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
func (ls *LedgerSession) GetPending(destination types.Address, hash types.Hash) (*types.PendingInfo, error) {
	key := ls.getPendingKey(destination, hash)
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
func (ls *LedgerSession) DeletePending(destination types.Address, hash types.Hash) error {
	key := ls.getPendingKey(destination, hash)
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
	log.Debugf("txn: %p, flag: %s", ls.txn, strconv.FormatBool(update))
	return ls.txn
}
func (ls *LedgerSession) releaseTxn() {
	if !ls.reuse {
		log.Debug("commit")
		err := ls.txn.Commit(nil)
		if err != nil {
			log.Fatal(err)
		}
		ls.Close()
	}
}

func (ls *LedgerSession) BatchUpdate(fn func() error) error {
	if !ls.reuse {
		return errors.New("batch update should enable reuse transaction")
	}
	txn := ls.NewTransaction(true)
	defer txn.Discard()

	if err := fn(); err != nil {
		return err
	}
	return txn.Commit(nil)
}
