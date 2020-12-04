// +build  testnet

package config

type ConfigV10 struct {
	ConfigV9  `mapstructure:",squash"`
	TrieClean *TrieClean `json:"trieClean"`
}

type TrieClean struct {
	Enable          bool   `json:"enable"`
	PeriodDay       int    `json:"periodDay"`
	HeightInterval  uint64 `json:"heightInterval"`  //  pov height interval to delete data
	SyncWriteHeight uint64 `json:"syncWriteHeight"` //  pov height to write trie data when sync
}

func DefaultConfigV10(dir string) (*ConfigV10, error) {
	var cfg ConfigV10
	cfg9, _ := DefaultConfigV9(dir)
	cfg.ConfigV9 = *cfg9
	cfg.Version = configVersion
	cfg.TrieClean = defaultTrieClean()
	return &cfg, nil
}

func defaultTrieClean() *TrieClean {
	return &TrieClean{
		Enable:          true,
		PeriodDay:       5,
		HeightInterval:  1000,
		SyncWriteHeight: 0,
	}
}
