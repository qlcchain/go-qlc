// +build  testnet

package config

type ConfigV9 struct {
	ConfigV8 `mapstructure:",squash"`
}

func DefaultConfigV9(dir string) (*ConfigV9, error) {
	var cfg ConfigV9
	cfg8, _ := DefaultConfigV8(dir)
	cfg.ConfigV8 = *cfg8
	cfg.RPC.PublicModules = append(cfg.RPC.PublicModules, "KYC")
	return &cfg, nil
}
