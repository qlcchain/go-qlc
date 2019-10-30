package ledger

import "github.com/qlcchain/go-qlc/common/types"

func (c *Cache) UpdateBlockUnConfirmed(block *types.StateBlock) error {
	am, err := c.GetAccountMetaUnConfirmed(block.GetAddress())
	if err == nil {
		tm := am.Token(block.GetToken())
		if tm != nil {
			c.unConfirmedBlock.Remove(tm.Header)
		}
	}
	return c.unConfirmedBlock.Set(block.GetHash(), block)
}

func (c *Cache) GetBlockUnConfirmed(hash types.Hash) (*types.StateBlock, error) {
	v, err := c.unConfirmedBlock.Get(hash)
	if err != nil {
		return nil, err
	}
	return v.(*types.StateBlock), nil
}

func (c *Cache) DeleteBlockUnConfirmed(hash types.Hash) {
	c.unConfirmedBlock.Remove(hash)
}

func (c *Cache) UpdateBlockConfirmed(block *types.StateBlock) error {
	am, err := c.GetAccountMetaConfirmed(block.GetAddress())
	if err == nil {
		tm := am.Token(block.GetToken())
		if tm != nil {
			c.confirmedBlock.Remove(tm.Header)
		}
	}
	return c.confirmedBlock.Set(block.GetHash(), block)
}

func (c *Cache) GetBlockConfirmed(hash types.Hash) (*types.StateBlock, error) {
	v, err := c.confirmedBlock.Get(hash)
	if err != nil {
		return nil, err
	}
	return v.(*types.StateBlock), nil
}

func (c *Cache) DeleteBlockConfirmed(hash types.Hash) {
	c.confirmedBlock.Remove(hash)
}
