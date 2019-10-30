package ledger

import "github.com/qlcchain/go-qlc/common/types"

func (c *Cache) UpdateAccountMetaUnConfirmed(am *types.AccountMeta) error {
	return c.unConfirmedAccount.Set(am.Address, *am)
}

func (c *Cache) GetAccountMetaUnConfirmed(address types.Address) (*types.AccountMeta, error) {
	v, err := c.unConfirmedAccount.Get(address)
	if err != nil {
		return nil, err
	}
	am := v.(types.AccountMeta)
	return &am, nil
}

func (c *Cache) DeleteAccountMetaUnConfirmed(address types.Address) {
	c.unConfirmedAccount.Remove(address)
}

func (c *Cache) UpdateAccountMetaConfirmed(am *types.AccountMeta) error {
	return c.confirmedAccount.Set(am.Address, *am)
}

func (c *Cache) GetAccountMetaConfirmed(address types.Address) (*types.AccountMeta, error) {
	v, err := c.confirmedAccount.Get(address)
	if err != nil {
		return nil, err
	}
	am := v.(types.AccountMeta)
	return &am, nil
}

func (c *Cache) DeleteAccountMetaConfirmed(address types.Address) {
	c.confirmedAccount.Remove(address)
}
