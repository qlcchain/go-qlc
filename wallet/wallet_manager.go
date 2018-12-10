/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import "github.com/google/uuid"

type walletManager interface {
	WalletIds() ([]uuid.UUID, error)
	NewWallet() (uuid.UUID, error)
	CurrentId() (uuid.UUID, error)
	RemoveWallet(id uuid.UUID) error
}
