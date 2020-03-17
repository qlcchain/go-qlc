package config

type ConfigV4 struct {
	ConfigV3 `mapstructure:",squash"`
	PoV      *PoVConfig `json:"pov"`
}

type PoVConfig struct {
	PovEnabled   bool   `json:"povEnabled"`
	MinerEnabled bool   `json:"minerEnabled"`
	Coinbase     string `json:"coinbase" validate:"address"`
	AlgoName     string `json:"algoName"`
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
		PovEnabled:   false,
		MinerEnabled: false,
		Coinbase:     "",
		AlgoName:     "",
	}
}
