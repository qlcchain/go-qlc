/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

type walletManager interface {
	WalletIds() ([]WalletId, error)
	NewWallet() (WalletId, error)
	CurrentId() (WalletId, error)
	RemoveWallet(id WalletId) error
}
