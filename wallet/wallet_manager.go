/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import "github.com/qlcchain/go-qlc/common/types"

type walletManager interface {
	WalletIds() ([]types.Address, error)
	NewWalletBySeed(seed string) (types.Address, error)
	NewWallet() (types.Address, error)
	CurrentId() (types.Address, error)
	RemoveWallet(id types.Address) error
	IsWalletExist(address types.Address) (bool, error)
}
