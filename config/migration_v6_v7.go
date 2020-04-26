/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"encoding/json"
)

type MigrationV6ToV7 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV6ToV7() *MigrationV6ToV7 {
	return &MigrationV6ToV7{startVersion: 6, endVersion: 7}
}

func (m *MigrationV6ToV7) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg6 ConfigV6
	err := json.Unmarshal(data, &cfg6)
	if err != nil {
		return data, version, err
	}

	cfg7, err := DefaultConfigV7(cfg6.DataDir)
	if err != nil {
		return data, version, err
	}
	cfg7.ConfigV6 = cfg6
	cfg7.Version = configVersion
	cfg7.P2P.ListeningIp = "127.0.0.1"

	bytes, _ := json.Marshal(cfg7)
	return bytes, m.endVersion, err
}

func (m *MigrationV6ToV7) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV6ToV7) EndVersion() int {
	return m.endVersion
}
