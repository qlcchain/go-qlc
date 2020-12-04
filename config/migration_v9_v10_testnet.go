// +build testnet

package config

import "encoding/json"

type MigrationV9ToV10 struct {
	startVersion int
	endVersion   int
}

func NewMigrationV9ToV10() *MigrationV9ToV10 {
	return &MigrationV9ToV10{startVersion: 9, endVersion: 10}
}

func (m *MigrationV9ToV10) Migration(data []byte, version int) ([]byte, int, error) {
	var cfg9 ConfigV9
	err := json.Unmarshal(data, &cfg9)
	if err != nil {
		return data, version, err
	}

	cfg10, err := DefaultConfigV10(cfg9.DataDir)
	if err != nil {
		return data, version, err
	}
	cfg10.ConfigV9 = cfg9
	cfg10.Version = configVersion
	cfg10.TrieClean = defaultTrieClean()

	bytes, _ := json.Marshal(cfg10)
	return bytes, m.endVersion, err
}

func (m *MigrationV9ToV10) StartVersion() int {
	return m.startVersion
}

func (m *MigrationV9ToV10) EndVersion() int {
	return m.endVersion
}
