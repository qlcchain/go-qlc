// +build  !testnet

package config

type ConfigV10 struct {
	ConfigV8   `mapstructure:",squash"`
	DBOptimize *DBOptimize `json:"dbOptimize"`
}

type DBOptimize struct {
	Enable          bool   `json:"enable"`
	PeriodDay       int    `json:"periodDay"`
	HeightInterval  uint64 `json:"heightInterval"`  //  pov height interval to delete data
	SyncWriteHeight uint64 `json:"syncWriteHeight"` //  min pov height to write trie data when sync
	MaxUsage        int    `json:"maxUsage"`
	FlushInterval   int    `json:"flushInterval"`
}

func DefaultConfigV10(dir string) (*ConfigV10, error) {
	var cfg ConfigV10
	cfg8, _ := DefaultConfigV8(dir)
	cfg.ConfigV8 = *cfg8
	cfg.Version = configVersion
	cfg.DBOptimize = defaultOptimize()
	return &cfg, nil
}

func defaultOptimize() *DBOptimize {
	return &DBOptimize{
		Enable:          true,
		PeriodDay:       7,
		HeightInterval:  1000,
		SyncWriteHeight: 400000,
		MaxUsage:        95,
		FlushInterval:   20,
	}
}
