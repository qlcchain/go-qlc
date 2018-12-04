package ledger

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type Ledger struct {
	store db.Store
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

func NewLedger() (*Ledger, error) {
	dir := util.QlcDir("ledger")
	store, err := db.NewBadgerStore(dir)
	if err != nil {
		return nil, err
	}
	ledger := Ledger{store: store}
	return &ledger, nil
}

func (l *Ledger) Close() error {
	return l.store.Close()
}
func (l *Ledger) UpdateDataInTransaction(fn func(txn db.StoreTxn) error) error {
	return l.store.Update(func(txn db.StoreTxn) error {
		return fn(txn)
	})
}
func (l *Ledger) GetDataInTransaction(fn func(txn db.StoreTxn) error) error {
	return l.store.View(func(txn db.StoreTxn) error {
		return fn(txn)
	})
}
func (l *Ledger) Empty(txn db.StoreTxn) (bool, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}

	r := true
	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		r = false
		return nil
	})
	if err != nil {
		return r, err
	}
	return r, nil
}
func (l *Ledger) Flush(txn db.StoreTxn) error {
	return nil
}

// -------------------  Block  --------------------

func (l *Ledger) getBlockKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixBlock
	copy(key[1:], hash[:])
	return key[:]
}
func (l *Ledger) AddBlock(blk types.Block, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}

	hash := blk.GetHash()
	log.Info("adding block,", hash)
	blockBytes, err := blk.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := l.getBlockKey(hash)

	//never overwrite implicitly
	err = txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	if err = txn.SetWithMeta(key, blockBytes, byte(blk.GetType())); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) GetBlock(hash types.Hash, txn db.StoreTxn) (types.Block, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	key := l.getBlockKey(hash)
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
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	return blk, nil
}
func (l *Ledger) GetBlocks(txn db.StoreTxn) ([]types.Block, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	var blocks []types.Block
	//var blk types.Block
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
func (l *Ledger) DeleteBlock(hash types.Hash, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	key := l.getBlockKey(hash)
	if err := txn.Delete(key); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) HasBlock(hash types.Hash, txn db.StoreTxn) (bool, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	key := l.getBlockKey(hash)
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
func (l *Ledger) CountBlocks(txn db.StoreTxn) (uint64, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	var count uint64

	err := txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (l *Ledger) GetRandomBlock(txn db.StoreTxn) (types.Block, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	c, err := l.CountBlocks(txn)
	if err != nil {
		return nil, err
	}
	if c == 0 {
		return nil, ErrStoreEmpty
	}
	index := rand.Int63n(int64(c))
	var blk types.Block
	var temp int64
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

// ------------------- unchecked Block  --------------------

func (t *Ledger) uncheckedKindToPrefix(kind types.UncheckedKind) byte {
	switch kind {
	case types.UncheckedKindPrevious:
		return idPrefixUncheckedBlockPrevious
	case types.UncheckedKindLink:
		return idPrefixUncheckedBlockLink
	default:
		panic("bad unchecked block kind")
	}
}
func (t *Ledger) getUncheckedBlockKey(hash types.Hash, kind types.UncheckedKind) []byte {
	var key [1 + types.HashSize]byte
	key[0] = t.uncheckedKindToPrefix(kind)
	copy(key[1:], hash[:])
	return key[:]
}
func (l *Ledger) AddUncheckedBlock(parentHash types.Hash, blk types.Block, kind types.UncheckedKind, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	blockBytes, err := blk.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := l.getUncheckedBlockKey(parentHash, kind)

	//never overwrite implicitly
	err = txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrUncheckedBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	if err = txn.SetWithMeta(key, blockBytes, byte(blk.GetType())); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txn db.StoreTxn) (types.Block, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	key := l.getUncheckedBlockKey(parentHash, kind)
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
func (l *Ledger) DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	key := l.getUncheckedBlockKey(parentHash, kind)
	if err := txn.Delete(key); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind, txn db.StoreTxn) (bool, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	key := l.getUncheckedBlockKey(hash, kind)
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
func (l *Ledger) WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc, txn db.StoreTxn) error {
	return nil
}
func (l *Ledger) CountUncheckedBlocks(txn db.StoreTxn) (uint64, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	var count uint64

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

// ------------------- AccountMeta  --------------------

func (l *Ledger) getAccountMetaKey(address types.Address) []byte {
	var key [1 + types.AddressSize]byte
	key[0] = idPrefixAccount
	copy(key[1:], address[:])
	return key[:]
}
func (l *Ledger) AddAccountMeta(meta *types.AccountMeta, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := l.getAccountMetaKey(meta.Address)

	// never overwrite implicitly
	err = txn.Get(key, func(vals []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrAccountExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	if err := txn.Set(key, metaBytes); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) GetAccountMeta(address types.Address, txn db.StoreTxn) (*types.AccountMeta, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	key := l.getAccountMetaKey(address)
	var meta types.AccountMeta
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
func (l *Ledger) UpdateAccountMeta(meta *types.AccountMeta, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaKey(meta.Address)

	err = txn.Get(key, func(vals []byte, b byte) error {
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrAccountNotFound
		}
		return err
	}
	if err := txn.Set(key, metaBytes); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) AddOrUpdateAccountMeta(meta *types.AccountMeta, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaKey(meta.Address)
	if err := txn.Set(key, metaBytes); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) DeleteAccountMeta(address types.Address, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	key := l.getAccountMetaKey(address)
	if err := txn.Delete(key); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) HasAccountMeta(address types.Address, txn db.StoreTxn) (bool, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	key := l.getAccountMetaKey(address)
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

// ------------------- TokenMeta  --------------------

func (l *Ledger) AddTokenMeta(address types.Address, tokenmeta *types.TokenMeta, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	accountmeta, err := l.GetAccountMeta(address, txn)
	if err != nil {
		return err
	}
	for _, t := range accountmeta.Tokens {
		if t.Type == tokenmeta.Type {
			return ErrTokenExists
		}
	}
	accountmeta.Tokens = append(accountmeta.Tokens, tokenmeta)

	accountmetabytes, err := accountmeta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaKey(accountmeta.Address)
	if err := txn.Set(key, accountmetabytes); err != nil {
		return err
	}

	return txn.Commit(nil)
}
func (l *Ledger) GetTokenMeta(address types.Address, tokenType types.Hash, txn db.StoreTxn) (*types.TokenMeta, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	accountmeta, err := l.GetAccountMeta(address, txn)
	if err != nil {
		return nil, err
	}
	for _, token := range accountmeta.Tokens {
		if token.Type == tokenType {
			return token, nil
		}
	}
	return nil, ErrTokenNotFound
}
func (l *Ledger) UpdateTokenMeta(address types.Address, tokenmeta *types.TokenMeta, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	accountmeta, err := l.GetAccountMeta(address, txn)
	if err != nil {
		return err
	}
	tokens := accountmeta.Tokens
	for index, token := range accountmeta.Tokens {
		if token.Type == tokenmeta.Type {
			accountmeta.Tokens = append(tokens[:index], tokens[index+1:]...)
			accountmeta.Tokens = append(accountmeta.Tokens, tokenmeta)

			accountmetabytes, err := accountmeta.MarshalMsg(nil)
			if err != nil {
				return err
			}
			key := l.getAccountMetaKey(accountmeta.Address)
			if err := txn.Set(key, accountmetabytes); err != nil {
				return err
			}

			return txn.Commit(nil)
		}
	}
	return ErrTokenNotFound
}
func (l *Ledger) AddOrUpdateTokenMeta(address types.Address, tokenmeta *types.TokenMeta, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	accountmeta, err := l.GetAccountMeta(address, txn)
	if err != nil {
		return err
	}
	tokens := accountmeta.Tokens
	for index, token := range accountmeta.Tokens {
		if token.Type == tokenmeta.Type {
			accountmeta.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}
	accountmeta.Tokens = append(accountmeta.Tokens, tokenmeta)

	accountmetabytes, err := accountmeta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaKey(accountmeta.Address)
	if err := txn.Set(key, accountmetabytes); err != nil {
		return err
	}

	return txn.Commit(nil)
}
func (l *Ledger) DeleteTokenMeta(address types.Address, tokenType types.Hash, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	accountmeta, err := l.GetAccountMeta(address, txn)
	if err != nil {
		if err == ErrAccountNotFound {
			return nil
		}
		return err
	}
	tokens := accountmeta.Tokens
	for index, token := range tokens {
		if token.Type == tokenType {
			accountmeta.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}

	accountmetabytes, err := accountmeta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaKey(accountmeta.Address)
	if err := txn.Set(key, accountmetabytes); err != nil {
		return err
	}

	return txn.Commit(nil)
}
func (l *Ledger) HasTokenMeta(address types.Address, tokenType types.Hash, txn db.StoreTxn) (bool, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	accountmeta, err := l.GetAccountMeta(address, txn)
	if err != nil {
		if err == ErrAccountNotFound {
			return false, nil
		}
		return false, err
	}
	for _, t := range accountmeta.Tokens {
		if t.Type == tokenType {
			return true, nil
		}
	}
	return false, nil

}

// ------------------- representation  --------------------

func (t *Ledger) getRepresentationKey(address types.Address) []byte {
	var key [1 + types.AddressSize]byte
	key[0] = idPrefixRepresentation
	copy(key[1:], address[:])
	return key[:]
}
func (l *Ledger) AddRepresentationWeight(address types.Address, amount types.Balance, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	oldAmount, err := l.GetRepresentation(address, txn)
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	amount = oldAmount.Add(amount)
	key := l.getRepresentationKey(address)
	amountBytes, err := amount.MarshalText()
	if err != nil {
		return err
	}
	if err := txn.Set(key, amountBytes); err != nil {
		fmt.Println(err)
		return err
	}
	return txn.Commit(nil)

}
func (l *Ledger) SubRepresentationWeight(address types.Address, amount types.Balance, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	oldAmount, err := l.GetRepresentation(address, txn)
	if err != nil {
		return err
	}
	amount = oldAmount.Sub(amount)
	key := l.getRepresentationKey(address)
	amountBytes, err := amount.MarshalText()
	if err != nil {
		return err
	}
	if err := txn.Set(key, amountBytes); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) GetRepresentation(address types.Address, txn db.StoreTxn) (types.Balance, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	key := l.getRepresentationKey(address)
	var amount types.Balance
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

// ------------------- pending  --------------------

func (t *Ledger) getPendingKey(destination types.Address, hash types.Hash) []byte {
	var key [1 + types.PendingKeySize]byte
	key[0] = idPrefixPending
	copy(key[1:], destination[:])
	copy(key[1+types.AddressSize:], hash[:])
	return key[:]
}
func (l *Ledger) AddPending(destination types.Address, hash types.Hash, pending *types.PendingInfo, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}

	pendingBytes, err := pending.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getPendingKey(destination, hash)

	//never overwrite implicitly
	err = txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrPendingExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	if err := txn.Set(key, pendingBytes); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) GetPending(destination types.Address, hash types.Hash, txn db.StoreTxn) (*types.PendingInfo, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	key := l.getPendingKey(destination, hash)
	var pending types.PendingInfo
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
func (l *Ledger) DeletePending(destination types.Address, hash types.Hash, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	key := l.getPendingKey(destination, hash)
	if err := txn.Delete(key); err != nil {
		return err
	}
	return txn.Commit(nil)
}

// ------------------- frontier  --------------------

func (t *Ledger) getFrontierKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], hash[:])
	return key[:]
}
func (l *Ledger) AddFrontier(frontier *types.Frontier, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	key := l.getFrontierKey(frontier.HeaderBlock)

	// never overwrite implicitly
	err := txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrFrontierExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	if err := txn.Set(key, frontier.OpenBlock[:]); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) GetFrontier(hash types.Hash, txn db.StoreTxn) (*types.Frontier, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	key := l.getFrontierKey(hash)
	frontier := types.Frontier{HeaderBlock: hash}
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
func (l *Ledger) GetFrontiers(txn db.StoreTxn) ([]*types.Frontier, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	var frontiers types.Frontiers

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
	sort.Sort(frontiers)
	return frontiers, nil
}
func (l *Ledger) DeleteFrontier(hash types.Hash, txn db.StoreTxn) error {
	if txn == nil {
		txn = l.store.NewTransaction(true)
		defer txn.Discard()
	}
	key := l.getFrontierKey(hash)
	if err := txn.Delete(key); err != nil {
		return err
	}
	return txn.Commit(nil)
}
func (l *Ledger) CountFrontiers(txn db.StoreTxn) (uint64, error) {
	if txn == nil {
		txn = l.store.NewTransaction(false)
		defer txn.Discard()
	}
	var count uint64

	err := txn.Iterator(idPrefixFrontier, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
