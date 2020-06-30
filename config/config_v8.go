package config

type ConfigV8 struct {
	ConfigV7 `mapstructure:",squash"`
}

func DefaultConfigV8(dir string) (*ConfigV8, error) {
	var cfg ConfigV8
	cfg7, _ := DefaultConfigV7(dir)
	cfg.ConfigV7 = *cfg7
	cfg.RPC.PublicModules = defaultModules()
	cfg.RPC.GRPCConfig = defaultGRPCConfig()
	return &cfg, nil
}
