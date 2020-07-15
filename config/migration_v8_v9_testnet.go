// +build  testnet

package config

import "encoding/json"

type MigrationV8ToV9 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV8ToV9() *MigrationV8ToV9 {
	return &MigrationV8ToV9{startVersion: 8, endVersion: 9}
}

func (m *MigrationV8ToV9) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg8 ConfigV8
	err := json.Unmarshal(data, &cfg8)
	if err != nil {
		return data, version, err
	}

	cfg9, err := DefaultConfigV9(cfg8.DataDir)
	if err != nil {
		return data, version, err
	}

	cfg9.ConfigV8 = cfg8
	cfg9.Version = configVersion
	cfg9.RPC.PublicModules = append(cfg9.RPC.PublicModules, "KYC")

	bytes, _ := json.Marshal(cfg9)
	return bytes, m.endVersion, err
}

func (m *MigrationV8ToV9) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV8ToV9) EndVersion() int {
	return m.endVersion
}
