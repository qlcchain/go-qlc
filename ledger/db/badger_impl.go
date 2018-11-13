package db

import (
	"math/rand"
	"os"

	"github.com/dgraph-io/badger"
	badgerOpts "github.com/dgraph-io/badger/options"
	"github.com/qlcchain/go-qlc/common/types"
)

const (
	idPrefixBlock byte = iota
	idPrefixUncheckedBlockPrevious
	idPrefixUncheckedBlockLink
	idPrefixAccount
	idPrefixToken
	idPrefixFrontier
	idPrefixPending
	idPrefixRepresentation
)

const (
	badgerMaxOps = 10000
)

// BadgerStore represents a block lattice store backed by a badger database.
type BadgerStore struct {
	db *badger.DB
}

type BadgerStoreTxn struct {
	txn *badger.Txn
	db  *badger.DB
	ops uint64
}

// NewBadgerStore initializes/opens a badger database in the given directory.
func NewBadgerStore(dir string) (*BadgerStore, error) {
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	opts.ValueLogLoadingMode = badgerOpts.FileIO

	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := os.Mkdir(dir, 0700); err != nil {
			return nil, err
		}
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerStore{db: db}, nil
}

// Close closes the database
func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// Purge purges any old/deleted keys from the database.
func (s *BadgerStore) Purge() error {
	return s.db.RunValueLogGC(0.5)
}

func (s *BadgerStore) View(fn func(txn StoreTxn) error) error {
	//t := &BadgerStoreTxn{txn: s.db.NewTransaction(true), db: s.db}
	//defer t.txn.Discard()
	//if err := fn(t); err != nil {
	//	return err
	//}
	//return nil
	return s.db.View(func(txn *badger.Txn) error {
		return fn(&BadgerStoreTxn{txn: txn, db: s.db})
	})
}

func (s *BadgerStore) Update(fn func(txn StoreTxn) error) error {
	t := &BadgerStoreTxn{txn: s.db.NewTransaction(true), db: s.db}
	defer t.txn.Discard()

	if err := fn(t); err != nil {
		return err
	}
	return t.txn.Commit(nil)
}

func (t *BadgerStoreTxn) set(key []byte, val []byte) error {
	if err := t.txn.Set(key, val); err != nil {
		return err
	}

	t.ops++
	return nil
}

func (t *BadgerStoreTxn) setWithMeta(key []byte, val []byte, meta byte) error {
	if err := t.txn.SetWithMeta(key, val, meta); err != nil {
		return err
	}

	t.ops++
	return nil
}

func (t *BadgerStoreTxn) delete(key []byte) error {
	if err := t.txn.Delete(key); err != nil {
		return err
	}

	t.ops++
	return nil
}

// Empty reports whether the database is empty or not.
func (t *BadgerStoreTxn) Empty() (bool, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := [...]byte{idPrefixBlock}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		return false, nil
	}

	return true, nil
}

func (t *BadgerStoreTxn) Flush() error {
	if t.ops >= badgerMaxOps {
		if err := t.txn.Commit(nil); err != nil {
			return err
		}

		t.ops = 0
		t.txn = t.db.NewTransaction(true)
	}
	return nil
}

// ------------------- implement AccountMeta CURD -------------------

func (t *BadgerStoreTxn) getAccountMetaKey(address types.Address) (key [1 + types.AddressSize]byte) {
	key[0] = idPrefixAccount
	copy(key[1:], address[:])
	return
}

// AddAccountMeta adds the given AccountMeta to the database.
func (t *BadgerStoreTxn) AddAccountMeta(meta *types.AccountMeta) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := t.getAccountMetaKey(meta.Address)

	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return ErrAccountExists
	}
	return t.set(key[:], metaBytes)

}

// GetAccountMeta retrieves the AccountMeta with the given address from the database.
func (t *BadgerStoreTxn) GetAccountMeta(address types.Address) (*types.AccountMeta, error) {
	key := t.getAccountMetaKey(address)

	item, err := t.txn.Get(key[:])
	if err != nil {
		return nil, err
	}

	var meta types.AccountMeta
	err = item.Value(func(val []byte) error {
		metaBytes := val
		if _, err = meta.UnmarshalMsg(metaBytes); err != nil {
			return err
		} else {
			return nil
		}
	})

	if err != nil {
		return nil, err
	}

	return &meta, nil
}

// UpdateAccountMeta updates the accountmeta in the database.
func (t *BadgerStoreTxn) UpdateAccountMeta(meta *types.AccountMeta) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := t.getAccountMetaKey(meta.Address)
	return t.set(key[:], metaBytes)
}

// DeleteAccountMeta deletes the given address in the database.
func (t *BadgerStoreTxn) DeleteAccountMeta(address types.Address) error {
	key := t.getAccountMetaKey(address)
	return t.delete(key[:])
}

// HasAccountMeta reports whether the database contains a account with the address.
func (t *BadgerStoreTxn) HasAccountMeta(address types.Address) (bool, error) {
	key := t.getAccountMetaKey(address)

	if _, err := t.txn.Get(key[:]); err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ------------------- implement Token CURD ---------------------

func (t *BadgerStoreTxn) AddTokenMeta(address types.Address, meta *types.TokenMeta) error {
	accountmeta, err := t.GetAccountMeta(address)
	if err != nil {
		return err
	}

	for _, t := range accountmeta.Tokens {
		if t.Type == meta.Type {
			return ErrTokenExists
		}
	}

	accountmeta.Tokens = append(accountmeta.Tokens, meta)
	err = t.UpdateAccountMeta(accountmeta)
	if err != nil {
		return err
	}
	return nil
}

func (t *BadgerStoreTxn) GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error) {
	accountmeta, err := t.GetAccountMeta(address)
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

func (t *BadgerStoreTxn) UpdateTokenMeta(address types.Address, meta *types.TokenMeta) error {
	accountmeta, err := t.GetAccountMeta(address)
	if err != nil {
		return err
	}
	for _, token := range accountmeta.Tokens {
		if token.Type == meta.Type {
			if err = t.DelTokenMeta(address, token); err != nil {
				return err
			}
			if err := t.AddTokenMeta(address, meta); err != nil {
				return err
			} else {
				return nil
			}
		}
	}
	return ErrTokenNotFound
}

func (t *BadgerStoreTxn) DelTokenMeta(address types.Address, meta *types.TokenMeta) error {
	deleteTokenType := meta.Type
	accountmeta, err := t.GetAccountMeta(address)
	if err != nil {
		return err
	}
	tokens := accountmeta.Tokens
	for index, token := range tokens {
		if token.Type == deleteTokenType {
			accountmeta.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}

	err = t.UpdateAccountMeta(accountmeta)
	if err != nil {
		return err
	}
	return nil
}

// ------------------- implement Block CURD --------------------

func (t *BadgerStoreTxn) getBlockKey(hash types.Hash) (key [1 + types.HashSize]byte) {
	key[0] = idPrefixBlock
	copy(key[1:], hash[:])
	return
}

// AddBlock adds the given block to the database.
func (t *BadgerStoreTxn) AddBlock(blk types.Block) error {
	hash := blk.Hash()
	blockBytes, err := blk.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := t.getBlockKey(hash)
	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return ErrBlockExists
	}

	return t.setWithMeta(key[:], blockBytes, byte(blk.ID()))
}

// GetBlock retrieves the block with the given hash from the database.
func (t *BadgerStoreTxn) GetBlock(hash types.Hash) (types.Block, error) {
	key := t.getBlockKey(hash)
	item, err := t.txn.Get(key[:])
	if err != nil {
		return nil, err
	}

	blockType := item.UserMeta()
	blk, err := types.NewBlock(blockType)
	if err != nil {
		return nil, err
	}

	err = item.Value(func(val []byte) error {
		blockBytes := val
		if _, err = blk.UnmarshalMsg(blockBytes); err != nil {
			return err
		} else {
			return nil
		}
	})

	if err != nil {
		return nil, err
	}

	return blk, nil
}

func (t *BadgerStoreTxn) GetBlocks() ([]*types.Block, error) {
	var blocks []*types.Block

	it := t.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := [...]byte{idPrefixBlock}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		item := it.Item()

		block, err := types.NewBlock(byte(types.State))
		if err != nil {
			return nil, err
		}
		err = item.Value(func(val []byte) error {
			blockBytes := val
			if _, err := block.UnmarshalMsg(blockBytes); err != nil {
				return err
			} else {
				blocks = append(blocks, &block)
				return nil
			}
		})

		if err != nil {
			return nil, err
		}
	}

	return blocks, nil
}

// DeleteBlock deletes the given block in the database.
func (t *BadgerStoreTxn) DeleteBlock(hash types.Hash) error {
	key := t.getBlockKey(hash)
	return t.delete(key[:])
}

// HasBlock reports whether the database contains a block with the given hash.
func (t *BadgerStoreTxn) HasBlock(hash types.Hash) (bool, error) {
	key := t.getBlockKey(hash)

	if _, err := t.txn.Get(key[:]); err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CountBlocks returns the total amount of blocks in the database.
func (t *BadgerStoreTxn) CountBlocks() (uint64, error) {
	var count uint64
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := [...]byte{idPrefixBlock}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		count++
	}

	return count, nil
}

func (t *BadgerStoreTxn) GetRandomBlock() (types.Block, error) {
	count, err := t.CountBlocks()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrStoreEmpty
	}
	index := rand.Int63n(int64(count))
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := t.txn.NewIterator(opts)
	defer it.Close()

	var temp int64
	prefix := [...]byte{idPrefixBlock}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		if temp == index {
			item := it.Item()
			blockType := item.UserMeta()
			blk, err := types.NewBlock(blockType)
			if err != nil {
				return nil, err
			}
			err = item.Value(func(val []byte) error {
				blockBytes := val
				if _, err = blk.UnmarshalMsg(blockBytes); err != nil {
					return err
				} else {
					return nil
				}
			})
			if err != nil {
				return nil, err
			}
			return blk, nil
		}
		temp++

	}
	return nil, ErrBlockNotFound
}

// ------------------- implement Representation CURD --------------------

func (t *BadgerStoreTxn) setRepresentation(address types.Address, amount types.Balance) error {
	var key [1 + types.AddressSize]byte
	key[0] = idPrefixRepresentation
	copy(key[1:], address[:])

	amountBytes, err := amount.MarshalText()
	if err != nil {
		return err
	}

	return t.set(key[:], amountBytes)
}

func (t *BadgerStoreTxn) AddRepresentation(address types.Address, amount types.Balance) error {
	oldAmount, err := t.GetRepresentation(address)
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return t.setRepresentation(address, oldAmount.Add(amount))
}

func (t *BadgerStoreTxn) SubRepresentation(address types.Address, amount types.Balance) error {
	oldAmount, err := t.GetRepresentation(address)
	if err != nil {
		return err
	}
	return t.setRepresentation(address, oldAmount.Sub(amount))
}

func (t *BadgerStoreTxn) GetRepresentation(address types.Address) (types.Balance, error) {
	var key [1 + types.AddressSize]byte
	key[0] = idPrefixRepresentation
	copy(key[1:], address[:])

	item, err := t.txn.Get(key[:])
	if err != nil {
		return types.ZeroBalance, err
	}

	var amount types.Balance
	err = item.Value(func(val []byte) error {
		if err := amount.UnmarshalText(val); err != nil {
			return err
		} else {
			return nil
		}
	})

	return amount, nil
}

// ------------------- implement UncheckedBlock CURD --------------------

func (t *BadgerStoreTxn) uncheckedKindToPrefix(kind types.UncheckedKind) byte {
	switch kind {
	case types.UncheckedKindPrevious:
		return idPrefixUncheckedBlockPrevious
	case types.UncheckedKindLink:
		return idPrefixUncheckedBlockLink
	default:
		panic("bad unchecked block kind")
	}
}

func (t *BadgerStoreTxn) getUncheckedBlockKey(hash types.Hash, kind types.UncheckedKind) (key [1 + types.HashSize]byte) {
	key[0] = t.uncheckedKindToPrefix(kind)
	copy(key[1:], hash[:])
	return
}

func (t *BadgerStoreTxn) AddUncheckedBlock(parentHash types.Hash, blk types.Block, kind types.UncheckedKind) error {
	blockBytes, err := blk.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := t.getUncheckedBlockKey(parentHash, kind)
	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return ErrBlockExists
	}

	return t.txn.SetWithMeta(key[:], blockBytes, byte(blk.ID()))
}

func (t *BadgerStoreTxn) DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) error {
	key := t.getUncheckedBlockKey(parentHash, kind)
	return t.delete(key[:])
}

func (t *BadgerStoreTxn) GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) (types.Block, error) {
	key := t.getUncheckedBlockKey(parentHash, kind)

	item, err := t.txn.Get(key[:])
	if err != nil {
		return nil, err
	}

	blockType := item.UserMeta()

	blk, err := types.NewBlock(blockType)
	if err != nil {
		return nil, err
	}

	err = item.Value(func(val []byte) error {
		blockBytes := val
		if _, err := blk.UnmarshalMsg(blockBytes); err != nil {
			return err
		} else {
			return nil
		}
	})

	if err != nil {
		return nil, err
	}

	return blk, nil
}

func (t *BadgerStoreTxn) HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind) (bool, error) {
	key := t.getUncheckedBlockKey(hash, kind)
	if _, err := t.txn.Get(key[:]); err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (t *BadgerStoreTxn) WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc) error {
	panic("implement me")
}

func (t *BadgerStoreTxn) countUncheckedBlocks(kind types.UncheckedKind) uint64 {
	var count uint64
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := [...]byte{t.uncheckedKindToPrefix(kind)}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		count++
	}

	return count
}

func (t *BadgerStoreTxn) CountUncheckedBlocks() (uint64, error) {
	return t.countUncheckedBlocks(types.UncheckedKindLink) +
		t.countUncheckedBlocks(types.UncheckedKindPrevious), nil
}

// ------------------- implement Pending CURD --------------------

func (t *BadgerStoreTxn) getPendingKey(destination types.Address, hash types.Hash) [1 + types.PendingKeySize]byte {
	var key [1 + types.PendingKeySize]byte
	key[0] = idPrefixPending
	copy(key[1:], destination[:])
	copy(key[1+types.AddressSize:], hash[:])
	return key
}

func (t *BadgerStoreTxn) AddPending(destination types.Address, hash types.Hash, pending *types.PendingInfo) error {
	pendingBytes, err := pending.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := t.getPendingKey(destination, hash)

	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return ErrPendingExists
	}

	return t.set(key[:], pendingBytes)
}

func (t *BadgerStoreTxn) GetPending(destination types.Address, hash types.Hash) (*types.PendingInfo, error) {
	key := t.getPendingKey(destination, hash)

	item, err := t.txn.Get(key[:])
	if err != nil {
		return nil, err
	}

	var pending types.PendingInfo
	err = item.Value(func(val []byte) error {
		infoBytes := val
		if _, err := pending.UnmarshalMsg(infoBytes); err != nil {
			return err
		} else {
			return nil
		}
	})

	if err != nil {
		return nil, err
	}
	return &pending, nil
}

func (t *BadgerStoreTxn) DeletePending(destination types.Address, hash types.Hash) error {
	key := t.getPendingKey(destination, hash)
	return t.delete(key[:])
}

// ------------------- implement Frontier CURD --------------------

func (t *BadgerStoreTxn) getFrontierKey(hash types.Hash) [1 + types.HashSize]byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], hash[:])
	return key
}

func (t *BadgerStoreTxn) AddFrontier(frontier *types.Frontier) error {
	key := t.getFrontierKey(frontier.Hash)

	// never overwrite implicitly
	if _, err := t.txn.Get(key[:]); err != nil && err != badger.ErrKeyNotFound {
		return err
	} else if err == nil {
		return ErrFrontierExists
	}

	return t.set(key[:], frontier.Address[:])
}

func (t *BadgerStoreTxn) GetFrontier(hash types.Hash) (*types.Frontier, error) {
	key := t.getFrontierKey(hash)
	item, err := t.txn.Get(key[:])

	if err != nil {
		return nil, err
	}

	frontier := types.Frontier{Hash: hash}
	err = item.Value(func(val []byte) error {
		copy(frontier.Address[:], val)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &frontier, nil
}

func (t *BadgerStoreTxn) GetFrontiers() ([]*types.Frontier, error) {
	var frontiers []*types.Frontier
	it := t.txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := [...]byte{idPrefixFrontier}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		item := it.Item()

		var frontier types.Frontier
		err := item.Value(func(val []byte) error {
			copy(frontier.Address[:], val)
			copy(frontier.Hash[:], item.Key()[1:])
			return nil
		})

		if err != nil {
			return nil, err
		}
		frontiers = append(frontiers, &frontier)
	}

	return frontiers, nil
}

func (t *BadgerStoreTxn) DeleteFrontier(hash types.Hash) error {
	key := t.getFrontierKey(hash)
	return t.delete(key[:])
}

func (t *BadgerStoreTxn) CountFrontiers() (uint64, error) {
	var count uint64
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false

	it := t.txn.NewIterator(opts)
	defer it.Close()

	prefix := [...]byte{idPrefixFrontier}
	for it.Seek(prefix[:]); it.ValidForPrefix(prefix[:]); it.Next() {
		count++
	}
	return count, nil
}
