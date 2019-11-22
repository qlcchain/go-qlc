package config

type ConfigV6 struct {
	ConfigV5 `mapstructure:",squash"`
	Genesis  *Genesis `json:"genesis"`
}

type Genesis struct {
	GenesisBlocks []*GenesisInfo `json:"genesisBlocks"`
}

type GenesisInfo struct {
	ChainToken bool   `json:"chainToken"`
	GasToken   bool   `json:"gasToken"`
	Mintage    string `json:"mintage"`
	Genesis    string `json:"genesis"`
}

func DefaultConfigV6(dir string) (*ConfigV6, error) {
	var cfg ConfigV6
	cfg5, _ := DefaultConfigV5(dir)
	cfg.ConfigV5 = *cfg5
	cfg.Version = configVersion
	cfg.Genesis = defaultGenesis()
	return &cfg, nil
}

func defaultGenesis() *Genesis {
	genesisInfo := &GenesisInfo{
		ChainToken: true,
		GasToken:   false,
		Mintage:    jsonMintage,
		Genesis:    jsonGenesis,
	}
	gasInfo := &GenesisInfo{
		ChainToken: false,
		GasToken:   true,
		Mintage:    jsonMintageQGAS,
		Genesis:    jsonGenesisQGAS,
	}

	return &Genesis{
		GenesisBlocks: []*GenesisInfo{genesisInfo, gasInfo},
	}
}
