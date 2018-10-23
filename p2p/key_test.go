package p2p

import (
	"testing"

	"github.com/qlcchain/go-qlc/config"
)

func TestLoadNetworkKeyFromFileOrCreateNew(t *testing.T) {
	var cfg = config.DefaultlConfig
	LoadNetworkKeyFromFileOrCreateNew(cfg.PrivateKeyPath)
}
