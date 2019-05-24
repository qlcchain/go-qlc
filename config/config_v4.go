package config

type ConfigV4 struct {
	ConfigV3 `mapstructure:",squash"`
	PoV      *PoVConfig `json:"pov"`
}

type PoVConfig struct {
	MinerEnabled  bool   `json:"minerEnabled"`
	BlockInterval int    `json:"blockInterval"`
	BlockSize     int    `json:"blockSize"`
	TargetCycle   int    `json:"targetCycle"`
	ForkHeight    int    `json:"forkHeight"`
	Coinbase      string `json:"coinbase"`
}

func DefaultConfigV4(dir string) (*ConfigV4, error) {
	var cfg ConfigV4
	cfg3, _ := DefaultConfigV3(dir)
	cfg.ConfigV3 = *cfg3
	cfg.Version = 4

	cfg.RPC.PublicModules = append(cfg.RPC.PublicModules, "pov", "miner")

	cfg.PoV = defaultPoV()

	return &cfg, nil
}

func defaultPoV() *PoVConfig {
	return &PoVConfig{
		MinerEnabled:  false,
		BlockInterval: 30,
		BlockSize:     4 * 1024 * 1024,
		TargetCycle:   20,
		ForkHeight:    3,
		Coinbase:      "",
	}
}
