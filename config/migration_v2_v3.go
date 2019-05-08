/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import "encoding/json"

type MigrationV2ToV3 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV2ToV3() *MigrationV2ToV3 {
	return &MigrationV2ToV3{startVersion: 2, endVersion: 3}
}

func (m *MigrationV2ToV3) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg2 ConfigV2
	err := json.Unmarshal(data, &cfg2)
	if err != nil {
		return data, version, err
	}
	if len(cfg2.RPC.HttpVirtualHosts) == 0 {
		cfg2.RPC.HttpVirtualHosts = []string{"*"}
	}

	cfg3, err := DefaultConfigV3(cfg2.DataDir)
	if err != nil {
		return data, version, err
	}
	cfg3.ConfigV2 = cfg2
	cfg3.Version = 3
	cfg3.RPC.PublicModules = append(cfg3.RPC.PublicModules, "pledge")

	bytes, err := json.Marshal(cfg3)
	return bytes, m.endVersion, err
}

func (m *MigrationV2ToV3) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV2ToV3) EndVersion() int {
	return m.endVersion
}
