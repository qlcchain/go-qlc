package config

type ConfigV11 struct {
	ConfigV10 `mapstructure:",squash"`
}

func DefaultConfigV11(dir string) (*ConfigV11, error) {
	var cfg ConfigV11
	cfg10, _ := DefaultConfigV10(dir)
	cfg.ConfigV10 = *cfg10
	cfg.RPC.PublicModules = append(cfg.RPC.PublicModules, "qgasswap")
	return &cfg, nil
}
