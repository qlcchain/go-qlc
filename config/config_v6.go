package config

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/types"
)

type ConfigV6 struct {
	ConfigV5 `mapstructure:",squash"`
	Genesis  *Genesis `json:"genesis"`
	Privacy  *Privacy `json:"privacy"`
}

type Genesis struct {
	GenesisBlocks []*GenesisInfo `json:"genesisBlocks"`
}

type GenesisInfo struct {
	ChainToken bool             `json:"chainToken"`
	GasToken   bool             `json:"gasToken"`
	Mintage    types.StateBlock `mapstructure:",squash" json:"mintage"`
	Genesis    types.StateBlock `mapstructure:",squash" json:"genesis"`
}

type Privacy struct {
	Enable  bool   `json:"enable"`
	PtmNode string `json:"ptmNode"`
}

func DefaultConfigV6(dir string) (*ConfigV6, error) {
	var cfg ConfigV6
	cfg5, _ := DefaultConfigV5(dir)
	cfg.ConfigV5 = *cfg5
	cfg.Version = configVersion
	cfg.Genesis = defaultGenesis()
	cfg.Privacy = defaultPrivacy()
	cfg.RPC.PublicModules = []string{"ledger", "account", "net", "util", "mintage", "contract", "pledge",
		"rewards", "pov", "miner", "config", "debug", "destroy", "metrics", "rep", "chain", "dpki", "settlement", "privacy"}
	return &cfg, nil
}

func defaultGenesis() *Genesis {
	var mintageBlock, genesisBlock, gasMintageBlock, gasGensisBlock types.StateBlock
	_ = json.Unmarshal([]byte(jsonGenesis), &genesisBlock)
	_ = json.Unmarshal([]byte(jsonMintage), &mintageBlock)
	_ = json.Unmarshal([]byte(jsonGenesisQGAS), &gasGensisBlock)
	_ = json.Unmarshal([]byte(jsonMintageQGAS), &gasMintageBlock)
	genesisInfo := &GenesisInfo{
		ChainToken: true,
		GasToken:   false,
		Mintage:    mintageBlock,
		Genesis:    genesisBlock,
	}
	gasInfo := &GenesisInfo{
		ChainToken: false,
		GasToken:   true,
		Mintage:    gasMintageBlock,
		Genesis:    gasGensisBlock,
	}

	return &Genesis{
		GenesisBlocks: []*GenesisInfo{genesisInfo, gasInfo},
	}
}

func defaultPrivacy() *Privacy {
	return &Privacy{}
}
