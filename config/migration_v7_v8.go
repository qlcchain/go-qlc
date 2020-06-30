package config

import (
	"encoding/json"
)

type MigrationV7ToV8 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV7ToV8() *MigrationV7ToV8 {
	return &MigrationV7ToV8{startVersion: 7, endVersion: 8}
}

func (m *MigrationV7ToV8) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg7 ConfigV7
	err := json.Unmarshal(data, &cfg7)
	if err != nil {
		return data, version, err
	}

	cfg8, err := DefaultConfigV8(cfg7.DataDir)
	if err != nil {
		return data, version, err
	}
	cfg8.ConfigV7 = cfg7
	cfg8.Version = configVersion
	cfg8.RPC.PublicModules = defaultModules()

	cfg8.RPC.GRPCConfig = defaultGRPCConfig()

	bytes, _ := json.Marshal(cfg8)
	return bytes, m.endVersion, err
}

func (m *MigrationV7ToV8) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV7ToV8) EndVersion() int {
	return m.endVersion
}
