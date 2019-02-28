package p2p

import (
	"context"
	"testing"

	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
)

func TestBootstrap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h1 := bhost.New(swarmt.GenSwarm(t, ctx))
	h2 := bhost.New(swarmt.GenSwarm(t, ctx))

	pinfos := make([]pstore.PeerInfo, 1)
	pinfos[0].ID = h2.ID()
	pinfos[0].Addrs = h2.Addrs()

	err := bootstrapConnect(ctx, h1, pinfos)

	if err != nil {
		t.Fatal(err)
	}

}

func TestConvertPeers(t *testing.T) {
	BootNodes := []string{"/ip4/47.90.89.43/tcp/29735/ipfs/QmVSWnHEdCD2AciuCECdxspvH3Ej7VSewY1vtiEMAYqoYN",
		"/ip4/127.0.0.1/tcp/29735/ipfs/QmVSWnHEdCD2AciuCECdxspvH3Ej7VSewY1vtiEMAYqoYN",
		"/ip4/0.0.0.0/tcp/29735/ipfs/QmVSWnHEdCD2AciuCECdxspvH3Ej7VSewY1vtiEMAYqoYN"}
	pinfos, err := convertPeers(BootNodes)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(pinfos))

}
