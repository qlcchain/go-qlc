// +build !testnet

package config

import "encoding/json"

type MigrationV8ToV10 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV8ToV10() *MigrationV8ToV10 {
	return &MigrationV8ToV10{startVersion: 8, endVersion: 10}
}

func (m *MigrationV8ToV10) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg8 ConfigV8
	err := json.Unmarshal(data, &cfg8)
	if err != nil {
		return data, version, err
	}

	cfg10, err := DefaultConfigV10(cfg8.DataDir)
	if err != nil {
		return data, version, err
	}
	cfg10.ConfigV8 = cfg8
	cfg10.Version = configVersion
	cfg10.TrieClean = defaultTrieClean()

	bytes, _ := json.Marshal(cfg10)
	return bytes, m.endVersion, err
}

func (m *MigrationV8ToV10) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV8ToV10) EndVersion() int {
	return m.endVersion
}
