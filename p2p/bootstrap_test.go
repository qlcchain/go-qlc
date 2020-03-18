package p2p

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
)

func TestBootstrap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h1 := bhost.New(swarmt.GenSwarm(t, ctx))
	h2 := bhost.New(swarmt.GenSwarm(t, ctx))

	pInfoS := make([]peer.AddrInfo, 1)
	pInfoS[0].ID = h2.ID()
	pInfoS[0].Addrs = h2.Addrs()

	err := bootstrapConnect(ctx, h1, pInfoS)

	if err != nil {
		t.Fatal(err)
	}
}

func TestConvertPeers(t *testing.T) {
	BootNodes := []string{"/ip4/47.90.89.43/tcp/29735/p2p/QmVSWnHEdCD2AciuCECdxspvH3Ej7VSewY1vtiEMAYqoYN",
		"/ip4/127.0.0.1/tcp/29735/p2p/QmVSWnHEdCD2AciuCECdxspvH3Ej7VSewY1vtiEMAYqoYN",
		"/ip4/0.0.0.0/tcp/29735/p2p/QmVSWnHEdCD2AciuCECdxspvH3Ej7VSewY1vtiEMAYqoYN"}
	pInfoS, err := convertPeers(BootNodes)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(pInfoS))
}
