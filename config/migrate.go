/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

type CfgMigrate interface {
	Migration(cfg []byte, version int) ([]byte, int, error)
	StartVersion() int
	EndVersion() int
}
