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

type MigrationV5ToV6 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV5ToV6() *MigrationV5ToV6 {
	return &MigrationV5ToV6{startVersion: 5, endVersion: 6}
}

func (m *MigrationV5ToV6) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg5 ConfigV5
	err := json.Unmarshal(data, &cfg5)
	if err != nil {
		return data, version, err
	}

	cfg6, err := DefaultConfigV6(cfg5.DataDir)
	if err != nil {
		return data, version, err
	}
	cfg6.ConfigV5 = cfg5
	cfg6.Version = configVersion
	cfg6.P2P.IsBootNode = false
	cfg6.P2P.BootNodeHttpServer = bootNodeHttpServer
	cfg6.P2P.BootNodes = bootNodes
	cfg6.P2P.Discovery.MDNSEnabled = true
	bytes, _ := json.Marshal(cfg6)
	return bytes, m.endVersion, err
}

func (m *MigrationV5ToV6) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV5ToV6) EndVersion() int {
	return m.endVersion
}
