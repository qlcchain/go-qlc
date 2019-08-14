/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import "encoding/json"

type MigrationV4ToV5 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV4ToV5() *MigrationV4ToV5 {
	return &MigrationV4ToV5{startVersion: 4, endVersion: 5}
}

func (m *MigrationV4ToV5) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg4 ConfigV4
	err := json.Unmarshal(data, &cfg4)
	if err != nil {
		return data, version, err
	}

	cfg5, err := DefaultConfigV5(cfg4.DataDir)
	if err != nil {
		return data, version, err
	}
	cfg5.ConfigV4 = cfg4
	//enable pov by default
	cfg5.PoV.PovEnabled = true
	cfg5.Version = configVersion

	bytes, err := json.Marshal(cfg5)
	return bytes, m.endVersion, err
}

func (m *MigrationV4ToV5) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV4ToV5) EndVersion() int {
	return m.endVersion
}
