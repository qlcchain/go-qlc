/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

type CfgMigrate interface {
	Migration(cfg *Config) error
	StartVersion() int
	EndVersion() int
}

type CfgMigrations []CfgMigrate

func (m CfgMigrations) Len() int {
	return len(m)
}

func (m CfgMigrations) Less(i, j int) bool {
	if m[i].StartVersion() < m[j].StartVersion() {
		return true
	}

	if m[i].StartVersion() > m[j].StartVersion() {
		return false
	}

	return m[i].EndVersion() < m[j].EndVersion()
}

func (m CfgMigrations) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
