package ledger

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
)

type BlockStore interface {
	AddStateBlock(block *types.StateBlock) error
	GetStateBlock(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error)
	HasStateBlock(hash types.Hash) (bool, error)
	GetRandomStateBlock() (*types.StateBlock, error)
	//
	UpdateStateBlock(block *types.StateBlock, c storage.Cache) error
	DeleteStateBlock(key types.Hash, c storage.Cache) error
	GetStateBlockConfirmed(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error)
	GetStateBlocksConfirmed(fn func(*types.StateBlock) error) error
	HasStateBlockConfirmed(hash types.Hash) (bool, error)
	CountStateBlocks() (uint64, error)

	AddBlockCache(blk *types.StateBlock, batch ...storage.Batch) error
	DeleteBlockCache(hash types.Hash, batch ...storage.Batch) error
	GetBlockCache(key types.Hash) (*types.StateBlock, error)
	GetBlockCaches(fn func(*types.StateBlock) error) error
	HasBlockCache(key types.Hash) (bool, error)
	CountBlocksCache() (uint64, error)

	GetBlockChild(hash types.Hash, c ...storage.Cache) (types.Hash, error)
	GetBlockLink(key types.Hash, c ...storage.Cache) (types.Hash, error)
}

func (l *Ledger) GetStateBlock(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error) {
	if b, err := l.GetBlockCache(hash); err == nil {
		return b, nil
	}

	if b, err := l.GetStateBlockConfirmed(hash, c...); err == nil {
		return b, nil
	} else {
		return nil, err
	}
}

func (l *Ledger) HasStateBlock(hash types.Hash) (bool, error) {
	if b, _ := l.HasBlockCache(hash); b {
		return true, nil
	}
	return l.HasStateBlockConfirmed(hash)
}

func (l *Ledger) GetRandomStateBlock() (*types.StateBlock, error) {
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
		prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock)
		err = l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
			if temp == index {
				var b = new(types.StateBlock)
				if err = b.Deserialize(val); err != nil {
					return err
				}
				if b.IsPrivate() {
					pl, err := l.GetBlockPrivatePayload(b.GetHash())
					if err == nil {
						b.SetPrivatePayload(pl)
					}
				}
				if !config.IsGenesisBlock(b) {
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

// Block Confirmed

func (l *Ledger) AddStateBlock(block *types.StateBlock) error {
	err := l.cache.BatchUpdate(func(c *Cache) error {
		if err := l.UpdateStateBlock(block, c); err != nil {
			l.logger.Error(err)
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	l.logger.Debug("publish addRelation,", block.GetHash())
	l.BlockConfirmed(block)
	l.EB.Publish(topic.EventAddRelation, block)
	return nil
}

func (l *Ledger) UpdateStateBlock(block *types.StateBlock, c storage.Cache) error {
	val, _ := block.Serialize()
	l.logger.Warnf("block length: %d [%s]", len(val), block.Type.String())
	if err := l.setStateBlock(block, c); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) setStateBlock(block *types.StateBlock, c storage.Cache) error {
	err := block.CheckPrivateRecvRsp()
	if err != nil {
		return err
	}
	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlock, block.GetHash())
	if err != nil {
		return err
	}
	//if b, err := l.HasBlockCache(block.GetHash()); b && err == nil {
	//	if err := l.DeleteBlockCache(block.GetHash()); err != nil {
	//		return fmt.Errorf("delete block cache error: %s", err)
	//	}
	//}
	if err := l.setBlockChild(block, c); err != nil {
		return fmt.Errorf("add block child error: %s", err)
	}
	if err := l.setBlockLink(block, c); err != nil {
		return fmt.Errorf("add block link error: %s", err)
	}
	return c.Put(k, block.Clone())
}

func (l *Ledger) DeleteStateBlock(key types.Hash, c storage.Cache) error {
	blk := new(types.StateBlock)
	blk, err := l.GetStateBlockConfirmed(key)
	if err != nil {
		return fmt.Errorf("get stateblock error %s", err)
	}

	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlock, key)
	if err != nil {
		return err
	}

	if err := c.Delete(k); err != nil {
		return err
	}

	if err := l.deleteBlockChild(blk, c); err != nil {
		return fmt.Errorf("delete child error: %s", err)
	}
	if err := l.deleteBlockLink(blk, c); err != nil {
		return fmt.Errorf("delete link error: %s", err)
	}

	l.logger.Info("publish deleteRelation,", key.String())
	l.EB.Publish(topic.EventDeleteRelation, key)
	return nil
}

func (l *Ledger) GetStateBlockConfirmed(hash types.Hash, c ...storage.Cache) (*types.StateBlock, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlock, hash)
	if err != nil {
		return nil, err
	}

	if r, err := l.getFromCache(k, c...); r != nil {
		return r.(*types.StateBlock).Clone(), nil
	} else {
		if err == ErrKeyDeleted {
			return nil, ErrBlockNotFound
		}
	}

	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	meta := new(types.StateBlock)
	if err := meta.Deserialize(v); err != nil {
		return nil, err
	}
	if meta.IsPrivate() {
		pl, err := l.GetBlockPrivatePayload(hash)
		if err == nil {
			meta.SetPrivatePayload(pl)
		}
	}
	return meta, nil
}

func (l *Ledger) GetStateBlocksConfirmed(fn func(*types.StateBlock) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixBlock)
	return l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			l.logger.Errorf("deserialize block error: %s", err)
			return nil
		}
		if blk.IsPrivate() {
			pl, err := l.GetBlockPrivatePayload(blk.GetHash())
			if err == nil {
				blk.SetPrivatePayload(pl)
			}
		}
		if err := fn(blk); err != nil {
			l.logger.Errorf("process block error: %s", err)
			return err
		}
		return nil
	})
}

func (l *Ledger) HasStateBlockConfirmed(hash types.Hash) (bool, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlock, hash)
	if err != nil {
		return false, err
	}

	if r, err := l.getFromCache(k); r != nil {
		return true, nil
	} else {
		if err == ErrKeyDeleted {
			return false, nil
		}
	}

	return l.store.Has(k)
}

func (l *Ledger) CountStateBlocks() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixBlock)})
}

// Block Child / Link
func (l *Ledger) GetBlockChild(hash types.Hash, c ...storage.Cache) (types.Hash, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixChild, hash)
	if err != nil {
		return types.ZeroHash, err
	}

	r, err := l.getFromCache(k, c...)
	if r != nil {
		h := r.(*types.Hash)
		return *h, nil
	} else {
		if err == ErrKeyDeleted {
			return types.ZeroHash, errors.New("block child not found")
		}
	}

	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return types.ZeroHash, errors.New("block child not found")
		}
		return types.ZeroHash, err
	}
	var meta types.Hash
	if err := meta.Deserialize(v); err != nil {
		return types.ZeroHash, errors.New("unmarshal child hash error")
	}
	return meta, nil
}

func (l *Ledger) GetBlockLink(hash types.Hash, c ...storage.Cache) (types.Hash, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixLink, hash)
	if err != nil {
		return types.ZeroHash, err
	}

	r, err := l.getFromCache(k, c...)
	if r != nil {
		h := r.(*types.Hash)
		return *h, nil
	} else {
		if err == ErrKeyDeleted {
			return types.ZeroHash, errors.New("block link not found")
		}
	}

	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return types.ZeroHash, errors.New("block link not found")
		}
		return types.ZeroHash, fmt.Errorf("get link error: %s", err)
	}
	meta := new(types.Hash)
	if err := meta.Deserialize(v); err != nil {
		return types.ZeroHash, errors.New("unmarshal link hash error")
	}
	return *meta, nil
}

func (l *Ledger) setBlockChild(cBlock *types.StateBlock, c storage.Cache) error {
	pHash := cBlock.Parent()
	cHash := cBlock.GetHash()
	if !config.IsGenesisBlock(cBlock) && pHash != types.ZeroHash && !cBlock.IsOpen() {
		// is parent block existed
		if b, _ := l.HasBlockCache(pHash); !b {
			if exist, _ := l.HasStateBlockConfirmed(pHash); !exist {
				return fmt.Errorf("%s can not find parent %s", cHash.String(), pHash.String())
			}
		}
		// is parent have used
		k, err := storage.GetKeyOfParts(storage.KeyPrefixChild, pHash)
		if err != nil {
			return err
		}
		if _, err := l.GetBlockChild(pHash, c); err == nil {
			return fmt.Errorf("%s already have child ", pHash.String())
		}
		return c.Put(k, &cHash)
	}
	return nil
}

func (l *Ledger) setBlockLink(block *types.StateBlock, c storage.Cache) error {
	if block.GetType() == types.Open || block.GetType() == types.Receive || block.GetType() == types.ContractReward {
		linkHash := block.GetLink()
		// is link block existed
		//if b, _ := l.HasBlockCache(linkHash); !b {
		//	if exist, _ := l.HasStateBlockConfirmed(linkHash); !exist {
		//		return fmt.Errorf("%s can not find link %s", block.GetHash().String(), linkHash.String())
		//	}
		//}

		h := block.GetHash()
		k, err := storage.GetKeyOfParts(storage.KeyPrefixLink, linkHash)
		if err != nil {
			return err
		}
		return c.Put(k, &h)
	}
	return nil
}

func (l *Ledger) deleteBlockChild(blk *types.StateBlock, c storage.Cache) error {
	pHash := blk.Parent()
	if !pHash.IsZero() {
		k, err := storage.GetKeyOfParts(storage.KeyPrefixChild, pHash)
		if err != nil {
			return err
		}
		if err := c.Delete(k); err != nil {
			return err
		}
	}
	return nil
}

func (l *Ledger) deleteBlockLink(blk *types.StateBlock, c storage.Cache) error {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixLink, blk.GetLink())
	if err != nil {
		return err
	}
	if err := c.Delete(k); err != nil {
		return err
	}
	sc, err := blk.ConvertToSchema()
	if err != nil {
		return err
	}
	l.deletedSchema = append(l.deletedSchema, sc...)
	return nil
}

// Block UnConfirmed
func (l *Ledger) AddBlockCache(blk *types.StateBlock, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCache, blk.GetHash())
	if err != nil {
		return err
	}
	v, err := blk.Serialize()
	if err != nil {
		return err
	}
	if _, err := b.Get(k); err == nil {
		return ErrBlockExists
	}
	if err := b.Put(k, v); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) GetBlockCache(hash types.Hash) (*types.StateBlock, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCache, hash)
	if err != nil {
		return nil, err
	}
	//if r, err := l.cache.Get(k); err == nil {
	//	return r.(*types.StateBlock), nil
	//}
	v, err := l.store.Get(k)
	if err != nil {
		if err == storage.KeyNotFound {
			return nil, ErrBlockNotFound
		}
		return nil, err
	}
	meta := new(types.StateBlock)
	if err := meta.Deserialize(v); err != nil {
		return nil, err
	}
	return meta, nil
}

func (l *Ledger) HasBlockCache(hash types.Hash) (bool, error) {
	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCache, hash)
	if err != nil {
		return false, err
	}
	//if b := l.cache.Has(k); b {
	//	return true
	//}
	return l.store.Has(k)
}

func (l *Ledger) DeleteBlockCache(hash types.Hash, batch ...storage.Batch) error {
	b, flag := l.getBatch(true, batch...)
	defer l.releaseBatch(b, flag)

	k, err := storage.GetKeyOfParts(storage.KeyPrefixBlockCache, hash)
	if err != nil {
		return err
	}

	return b.Delete(k)
}

func (l *Ledger) CountBlocksCache() (uint64, error) {
	return l.store.Count([]byte{byte(storage.KeyPrefixBlockCache)})
}

func (l *Ledger) GetBlockCaches(fn func(*types.StateBlock) error) error {
	prefix, _ := storage.GetKeyOfParts(storage.KeyPrefixBlockCache)
	return l.store.Iterator(prefix, nil, func(key []byte, val []byte) error {
		blk := new(types.StateBlock)
		if err := blk.Deserialize(val); err != nil {
			l.logger.Errorf("deserialize block error: %s", err)
			return nil
		}
		if err := fn(blk); err != nil {
			l.logger.Errorf("process block error: %s", err)
			return err
		}
		return nil
	})
}
