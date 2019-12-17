package p2p

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"

	"github.com/qlcchain/go-qlc/config"
)

func TestQlcNode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h1 := bhost.New(swarmt.GenSwarm(t, ctx))
	cfgFile := filepath.Join(config.QlcTestDataDir(), "node", uuid.New().String())
	defer func() {
		err := os.RemoveAll(filepath.Join(config.QlcTestDataDir(), "node"))
		if err != nil {
			t.Fatal(err)
		}
	}()
	cfg, err := config.NewCfgManager(cfgFile).Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.P2P.BootNodes[0] = h1.Addrs()[0].String() + "/" + "ipfs/" + h1.ID().Pretty()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/18888"
	node, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = node.buildHost()
	if err != nil {
		t.Fatal(err)
	}
}
