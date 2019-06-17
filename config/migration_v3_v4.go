/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import "encoding/json"

type MigrationV3ToV4 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV3ToV4() *MigrationV3ToV4 {
	return &MigrationV3ToV4{startVersion: 3, endVersion: 4}
}

func (m *MigrationV3ToV4) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg3 ConfigV3
	err := json.Unmarshal(data, &cfg3)
	if err != nil {
		return data, version, err
	}

	cfg4, err := DefaultConfigV4(cfg3.DataDir)
	if err != nil {
		return data, version, err
	}
	cfg4.ConfigV3 = cfg3
	cfg4.Version = 4
	cfg4.RPC.PublicModules = append(cfg3.RPC.PublicModules, "pov", "miner")

	bytes, err := json.Marshal(cfg4)
	return bytes, m.endVersion, err
}

func (m *MigrationV3ToV4) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV3ToV4) EndVersion() int {
	return m.endVersion
}
