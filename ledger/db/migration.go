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

type Migrations []Migration

func (m Migrations) Len() int {
	return len(m)
}

func (m Migrations) Less(i, j int) bool {
	if m[i].StartVersion() < m[j].StartVersion() {
		return true
	}

	if m[i].StartVersion() > m[j].StartVersion() {
		return false
	}

	return m[i].EndVersion() < m[j].EndVersion()
}

func (m Migrations) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
