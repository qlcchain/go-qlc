package p2p

import (
	"context"
	"github.com/google/uuid"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	qcontext "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"
	"os"
	"path/filepath"
	"testing"
)

func TestQlcNode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h1 := bhost.New(swarmt.GenSwarm(t, ctx))
	cfgFile := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	defer func() {
		err := os.RemoveAll(config.QlcTestDataDir())
		if err != nil {
			t.Fatal(err)
		}
	}()
	cfg, err := config.NewCfgManager(cfgFile).Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.P2P.BootNodes[0] = h1.Addrs()[0].String() + "/" + "ipfs/" + h1.ID().Pretty()
	node, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = node.buildHost()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLocalNetAttribute(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc := qcontext.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19998"
	b := "/ip4/127.0.0.1/tcp/19747/ipfs/" + cfg.P2P.ID.PeerID
	cfg.P2P.BootNodes = []string{b}
	node, _ := NewQlcService(dir)
	err := node.Start()
	if err != nil {
		t.Fatal(err)
	}
	if node.node.localNetAttribute != Intranet {
		t.Fatal("net attribute should be Intranet")
	}
}
