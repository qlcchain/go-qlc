package p2p

import (
	"context"

	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/qlcchain/go-qlc/config"

	"testing"
)

func TestQlcNode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h1 := bhost.New(swarmt.GenSwarm(t, ctx))
	cfg := config.DefaultlConfig
	cfg.BootNodes[0] = h1.Addrs()[0].String() + "/" + "ipfs/" + h1.ID().Pretty()
	node, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = node.startHost()
	if err != nil {
		t.Fatal(err)
	}
}
