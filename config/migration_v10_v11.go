package config

import "encoding/json"

type MigrationV10ToV11 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV10ToV11() *MigrationV10ToV11 {
	return &MigrationV10ToV11{startVersion: 10, endVersion: 11}
}

func (m *MigrationV10ToV11) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg10 ConfigV10
	err := json.Unmarshal(data, &cfg10)
	if err != nil {
		return data, version, err
	}

	cfg11, err := DefaultConfigV11(cfg10.DataDir)
	if err != nil {
		return data, version, err
	}

	cfg11.ConfigV10 = cfg10
	cfg11.Version = configVersion
	cfg11.RPC.PublicModules = append(cfg11.RPC.PublicModules, "qgasswap")

	bytes, _ := json.Marshal(cfg11)
	return bytes, m.endVersion, err
}

func (m *MigrationV10ToV11) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV10ToV11) EndVersion() int {
	return m.endVersion
}
