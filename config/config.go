package config

import (
	"path"

	"github.com/qlcchain/go-qlc/common/util"

	"github.com/qlcchain/go-qlc/common"
)

type P2P struct {
	Peers []string `json:"Peers"`
	Port  uint     `json:"Port"`
}
type RPC struct {
	Enable bool   `json:"Enable"`
	Host   string `json:"Host"`
	Port   uint   `json:"Port"`
}
type Node struct {
	Version        uint     `json:"Version"`
	Network        string   `json:"Network"`
	PrivateKeyPath string   `json:"PrivateKeyPath"`
	MdnsEnabled    bool     `json:"MdnsEnabled"`
	MdnsInterval   int      `json:"MdnsInterval"`
	BootNodes      []string `json:"BootNode"`
	Listen         string   `json:"Listen"`
}

type Config struct {
	*RPC  `json:"RPC"`
	*P2P  `json:"P2P"`
	*Node `json:"Node"`
}

var (
	qlccfg *ConfigManager

	DefaultlConfig = &Config{
		P2P: &P2P{
			Peers: []string{"47.90.89.43", "47.91.166.18"},
			Port:  29734,
		},
		RPC: &RPC{
			Enable: false,
			Host:   "127.0.0.1",
			Port:   29735,
		},
		Node: &Node{
			Version:        1,
			Network:        "testnet",
			PrivateKeyPath: "",
			MdnsEnabled:    true,
			MdnsInterval:   30,
			BootNodes:      []string{"/ip4/127.0.0.1/tcp/29734/ipfs/QmSv2rmNMNvyN548Bx75ii4B6hFYTsbYL3ANJw9dZmry11"},
			Listen:         "/ip4/0.0.0.0/tcp/29735",
		},
	}
)

func privateKeyPath() (string, error) {
	return path.Join(util.QlcDir(), "network.key"), nil
}

var log = common.NewLogger("config")

func init() {
	cfg := DefaultlConfig
	privateKeyPath, err := privateKeyPath()
	if err != nil {
		log.Errorf("privateKeyPath error: %s", err)
	}
	cfg.PrivateKeyPath = privateKeyPath
	if qlccfg, err = NewCfgManager(); err != nil {
		log.Errorf("config manager error: %s", err)
	}

	if err = qlccfg.Write(&cfg); err != nil {
		log.Errorf("config load error: %s", err)
	}

	if err = qlccfg.Read(&cfg); err != nil {
		log.Errorf("config load error: %s", err)
	}
}
