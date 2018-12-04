package ledger

import (
	"errors"
	"math/rand"
	"os"
	"path"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
)

type Ledger struct {
	store db.Store
	txn   db.StoreTxn
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
	dir := getBadgerDir()
	store, err := db.NewBadgerStore(dir)
	if err != nil {
		return nil, err
	}
	ledger := Ledger{store: store}
	return &ledger, nil
}

//TODO: refine path
func getBadgerDir() string {
	root, _ := os.Getwd()
	s := root[:strings.LastIndex(root, "go-qlc")]
	return path.Join(path.Join(path.Join(s, "go-qlc"), "ledger"), "db")
}

func (l *Ledger) Close() error {
	return l.store.Close()
}
func (l *Ledger) Update(fn func() error) error {
	return l.store.Update(func(txn db.StoreTxn) error {
		l.txn = txn
		return fn()
	})
}
func (l *Ledger) View(fn func() error) error {
	return l.store.View(func(txn db.StoreTxn) error {
		l.txn = txn
		return fn()
	})
}

// Empty reports whether the database is empty or not.
func (l *Ledger) Empty() (bool, error) {
	r := true
	err := l.txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		r = false
		return nil
	})
	if err != nil {
		return r, err
	}
	return r, nil
}
func (l *Ledger) Flush() error {
	return nil
}

// -------------------  Block  --------------------

func (l *Ledger) getBlockKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixBlock
	copy(key[1:], hash[:])
	return key[:]
}
func (l *Ledger) AddBlock(blk types.Block) error {
	hash := blk.GetHash()
	log.Info("adding block,", hash)
	blockBytes, err := blk.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := l.getBlockKey(hash)

	//never overwrite implicitly
	err = l.txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return l.txn.SetWithMeta(key, blockBytes, byte(blk.GetType()))
}
func (l *Ledger) GetBlock(hash types.Hash) (types.Block, error) {
	key := l.getBlockKey(hash)
	var blk types.Block
	err := l.txn.Get(key, func(val []byte, b byte) (err error) {
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
func (l *Ledger) GetBlocks() ([]types.Block, error) {
	var blocks []types.Block
	//var blk types.Block
	err := l.txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) (err error) {
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
func (l *Ledger) DeleteBlock(hash types.Hash) error {
	key := l.getBlockKey(hash)
	return l.txn.Delete(key)

}
func (l *Ledger) HasBlock(hash types.Hash) (bool, error) {
	key := l.getBlockKey(hash)
	err := l.txn.Get(key, func(val []byte, b byte) error {
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
func (l *Ledger) CountBlocks() (uint64, error) {
	var count uint64

	err := l.txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (l *Ledger) GetRandomBlock() (types.Block, error) {
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
	err = l.txn.Iterator(idPrefixBlock, func(key []byte, val []byte, b byte) error {
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
func (l *Ledger) AddUncheckedBlock(parentHash types.Hash, blk types.Block, kind types.UncheckedKind) error {

	blockBytes, err := blk.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := l.getUncheckedBlockKey(parentHash, kind)

	//never overwrite implicitly
	err = l.txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrUncheckedBlockExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	return l.txn.SetWithMeta(key, blockBytes, byte(blk.GetType()))
}
func (l *Ledger) GetUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) (types.Block, error) {
	key := l.getUncheckedBlockKey(parentHash, kind)
	var blk types.Block
	err := l.txn.Get(key, func(val []byte, b byte) (err error) {
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
func (l *Ledger) DeleteUncheckedBlock(parentHash types.Hash, kind types.UncheckedKind) error {
	key := l.getUncheckedBlockKey(parentHash, kind)
	return l.txn.Delete(key)
}
func (l *Ledger) HasUncheckedBlock(hash types.Hash, kind types.UncheckedKind) (bool, error) {
	key := l.getUncheckedBlockKey(hash, kind)
	err := l.txn.Get(key, func(val []byte, b byte) error {
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
func (l *Ledger) WalkUncheckedBlocks(visit types.UncheckedBlockWalkFunc) error {
	return nil
}
func (l *Ledger) CountUncheckedBlocks() (uint64, error) {
	var count uint64

	err := l.txn.Iterator(idPrefixUncheckedBlockLink, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	err = l.txn.Iterator(idPrefixUncheckedBlockPrevious, func(key []byte, val []byte, b byte) error {
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
func (l *Ledger) AddAccountMeta(meta *types.AccountMeta) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}

	key := l.getAccountMetaKey(meta.Address)

	// never overwrite implicitly
	err = l.txn.Get(key, func(vals []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrAccountExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return l.txn.Set(key, metaBytes)
}
func (l *Ledger) GetAccountMeta(address types.Address) (*types.AccountMeta, error) {
	key := l.getAccountMetaKey(address)
	var meta types.AccountMeta
	err := l.txn.Get(key, func(val []byte, b byte) (err error) {
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
func (l *Ledger) UpdateAccountMeta(meta *types.AccountMeta) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaKey(meta.Address)

	err = l.txn.Get(key, func(vals []byte, b byte) error {
		return nil
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return ErrAccountNotFound
		}
		return err
	}
	return l.txn.Set(key, metaBytes)
}
func (l *Ledger) AddOrUpdateAccountMeta(meta *types.AccountMeta) error {
	metaBytes, err := meta.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getAccountMetaKey(meta.Address)
	return l.txn.Set(key, metaBytes)
}
func (l *Ledger) DeleteAccountMeta(address types.Address) error {
	key := l.getAccountMetaKey(address)
	return l.txn.Delete(key)
}
func (l *Ledger) HasAccountMeta(address types.Address) (bool, error) {
	key := l.getAccountMetaKey(address)
	err := l.txn.Get(key, func(val []byte, b byte) error {
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

func (l *Ledger) AddTokenMeta(address types.Address, meta *types.TokenMeta) error {
	accountmeta, err := l.GetAccountMeta(address)
	if err != nil {
		return err
	}
	for _, t := range accountmeta.Tokens {
		if t.Type == meta.Type {
			return ErrTokenExists
		}
	}

	accountmeta.Tokens = append(accountmeta.Tokens, meta)
	return l.UpdateAccountMeta(accountmeta)
}
func (l *Ledger) GetTokenMeta(address types.Address, tokenType types.Hash) (*types.TokenMeta, error) {
	accountmeta, err := l.GetAccountMeta(address)
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
func (l *Ledger) UpdateTokenMeta(address types.Address, meta *types.TokenMeta) error {
	accountmeta, err := l.GetAccountMeta(address)
	if err != nil {
		return err
	}
	tokens := accountmeta.Tokens
	for index, token := range accountmeta.Tokens {
		if token.Type == meta.Type {
			accountmeta.Tokens = append(tokens[:index], tokens[index+1:]...)
			accountmeta.Tokens = append(accountmeta.Tokens, meta)
			return l.UpdateAccountMeta(accountmeta)
		}
	}
	return ErrTokenNotFound
}
func (l *Ledger) AddOrUpdateTokenMeta(address types.Address, meta *types.TokenMeta) error {
	accountmeta, err := l.GetAccountMeta(address)
	if err != nil {
		return err
	}
	tokens := accountmeta.Tokens
	for index, token := range accountmeta.Tokens {
		if token.Type == meta.Type {
			accountmeta.Tokens = append(tokens[:index], tokens[index+1:]...)
			accountmeta.Tokens = append(accountmeta.Tokens, meta)
			return l.UpdateAccountMeta(accountmeta)
		}
	}

	accountmeta.Tokens = append(accountmeta.Tokens, meta)
	return l.UpdateAccountMeta(accountmeta)
}
func (l *Ledger) DeleteTokenMeta(address types.Address, tokenType types.Hash) error {
	accountmeta, err := l.GetAccountMeta(address)
	if err != nil {
		return err
	}
	tokens := accountmeta.Tokens
	for index, token := range tokens {
		if token.Type == tokenType {
			accountmeta.Tokens = append(tokens[:index], tokens[index+1:]...)
		}
	}
	return l.UpdateAccountMeta(accountmeta)
}
func (l *Ledger) HasTokenMeta(address types.Address, tokenType types.Hash) (bool, error) {
	accountmeta, err := l.GetAccountMeta(address)
	if err != nil {
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
func (l *Ledger) AddRepresentationWeight(address types.Address, amount types.Balance) error {
	oldAmount, err := l.GetRepresentation(address)
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	amount = oldAmount.Add(amount)
	key := l.getRepresentationKey(address)
	amountBytes, err := amount.MarshalText()
	if err != nil {
		return err
	}
	return l.txn.Set(key, amountBytes)

}
func (l *Ledger) SubRepresentationWeight(address types.Address, amount types.Balance) error {
	oldAmount, err := l.GetRepresentation(address)
	if err != nil {
		return err
	}
	amount = oldAmount.Sub(amount)
	key := l.getRepresentationKey(address)
	amountBytes, err := amount.MarshalText()
	if err != nil {
		return err
	}
	return l.txn.Set(key, amountBytes)
}
func (l *Ledger) GetRepresentation(address types.Address) (types.Balance, error) {
	key := l.getRepresentationKey(address)
	var amount types.Balance
	err := l.txn.Get(key, func(val []byte, b byte) (err error) {
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
func (l *Ledger) AddPending(destination types.Address, hash types.Hash, pending *types.PendingInfo) error {
	pendingBytes, err := pending.MarshalMsg(nil)
	if err != nil {
		return err
	}
	key := l.getPendingKey(destination, hash)

	//never overwrite implicitly
	err = l.txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrPendingExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return l.txn.Set(key, pendingBytes)
}
func (l *Ledger) GetPending(destination types.Address, hash types.Hash) (*types.PendingInfo, error) {
	key := l.getPendingKey(destination, hash)
	var pending types.PendingInfo
	err := l.txn.Get(key[:], func(val []byte, b byte) (err error) {
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
func (l *Ledger) DeletePending(destination types.Address, hash types.Hash) error {
	key := l.getPendingKey(destination, hash)
	return l.txn.Delete(key)
}

// ------------------- frontier  --------------------

func (t *Ledger) getFrontierKey(hash types.Hash) []byte {
	var key [1 + types.HashSize]byte
	key[0] = idPrefixFrontier
	copy(key[1:], hash[:])
	return key[:]
}
func (l *Ledger) AddFrontier(frontier *types.Frontier) error {
	key := l.getFrontierKey(frontier.Hash)

	// never overwrite implicitly
	err := l.txn.Get(key, func(bytes []byte, b byte) error {
		return nil
	})
	if err == nil {
		return ErrFrontierExists
	} else if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	return l.txn.Set(key, frontier.Address[:])
}
func (l *Ledger) GetFrontier(hash types.Hash) (*types.Frontier, error) {
	key := l.getFrontierKey(hash)
	frontier := types.Frontier{Hash: hash}
	err := l.txn.Get(key, func(val []byte, b byte) (err error) {
		copy(frontier.Address[:], val)
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
func (l *Ledger) GetFrontiers() ([]*types.Frontier, error) {
	var frontiers []*types.Frontier

	err := l.txn.Iterator(idPrefixFrontier, func(key []byte, val []byte, b byte) error {
		var frontier types.Frontier
		copy(frontier.Hash[:], key[1:])
		copy(frontier.Address[:], val)
		frontiers = append(frontiers, &frontier)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return frontiers, nil
}
func (l *Ledger) DeleteFrontier(hash types.Hash) error {
	key := l.getFrontierKey(hash)
	return l.txn.Delete(key)
}
func (l *Ledger) CountFrontiers() (uint64, error) {
	var count uint64

	err := l.txn.Iterator(idPrefixFrontier, func(key []byte, val []byte, b byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
