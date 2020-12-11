// +build  testnet

package config

type ConfigV10 struct {
	ConfigV9   `mapstructure:",squash"`
	DBOptimize *DBOptimize `json:"dbOptimize"`
}

type DBOptimize struct {
	Enable          bool   `json:"enable"`
	PeriodDay       int    `json:"periodDay"`
	HeightInterval  uint64 `json:"heightInterval"`  //  pov height interval to delete data
	SyncWriteHeight uint64 `json:"syncWriteHeight"` //  min pov height to write trie data when sync
	MaxUsage        int    `json:"maxUsage"`
}

func DefaultConfigV10(dir string) (*ConfigV10, error) {
	var cfg ConfigV10
	cfg9, _ := DefaultConfigV9(dir)
	cfg.ConfigV9 = *cfg9
	cfg.Version = configVersion
	cfg.DBOptimize = defaultOptimize()
	return &cfg, nil
}

func defaultOptimize() *DBOptimize {
	return &DBOptimize{
		Enable:          true,
		PeriodDay:       7,
		HeightInterval:  1000,
		SyncWriteHeight: 0,
		MaxUsage:        90,
	}
}
