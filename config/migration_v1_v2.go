/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import "encoding/json"

type MigrationV1ToV2 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV1ToV2() *MigrationV1ToV2 {
	return &MigrationV1ToV2{startVersion: 1, endVersion: 2}
}

func (m *MigrationV1ToV2) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg1 ConfigV1
	err := json.Unmarshal(data, &cfg1)
	if err != nil {
		return data, version, err
	}

	cfg2, err := DefaultConfigV2(cfg1.DataDir)
	if err != nil {
		return data, version, err
	}

	cfg2.AutoGenerateReceive = cfg1.AutoGenerateReceive

	bytes, err := json.Marshal(cfg2)
	return bytes, m.endVersion, err
}

func (m *MigrationV1ToV2) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV1ToV2) EndVersion() int {
	return m.endVersion
}
