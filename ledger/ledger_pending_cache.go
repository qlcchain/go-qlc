package ledger

import "github.com/qlcchain/go-qlc/common/types"

func (c *rCache) AddAccountPending(address types.Address, token types.Hash, amount types.Balance) error {
	th, _ := types.HashBytes(address[:], token[:])
	return c.accountPending.Set(th, amount)
}

func (c *rCache) UpdateAccountPending(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo, add bool) error {
	th, _ := types.HashBytes(pendingKey.Address[:], pendingInfo.Type[:])
	v, err := c.accountPending.Get(th)
	if err == nil {
		amount := v.(types.Balance)
		if add {
			amount = amount.Add(pendingInfo.Amount)
		} else {
			amount = amount.Sub(pendingInfo.Amount)
		}
		return c.accountPending.Set(th, amount)
	}
	return nil
}

func (c *rCache) GetAccountPending(address types.Address, token types.Hash) (types.Balance, error) {
	th, _ := types.HashBytes(address[:], token[:])
	v, err := c.accountPending.Get(th)
	if err != nil {
		return types.ZeroBalance, err
	}
	return v.(types.Balance), nil
}

func (c *rCache) DeleteAccountPending(address types.Address, token types.Hash) {
	th, _ := types.HashBytes(address[:], token[:])
	c.accountPending.Remove(th)
}
