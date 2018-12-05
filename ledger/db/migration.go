/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package db

type Migration interface {
	Migrate(txn StoreTxn) error
	StartVersion() int
	EndVersion() int
}
