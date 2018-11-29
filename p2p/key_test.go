package p2p

import (
	"testing"

	"github.com/libp2p/go-libp2p-peer"
	"github.com/qlcchain/go-qlc/config"
)

func TestKey(t *testing.T) {
	var cfg = config.DefaultlConfig
	PrivKey, err := LoadNetworkKeyFromFileOrCreateNew(cfg.PrivateKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := peer.IDFromPublicKey(PrivKey.GetPublic())
	t.Log(pid)
	pri1, err := generateEd25519Key()
	if err != nil {
		t.Fatal(err)
	}
	pid1, err := peer.IDFromPublicKey(pri1.GetPublic())
	t.Log(pid1)
	str, err := marshalNetworkKey(pri1)
	if err != nil {
		t.Fatal(err)
	}
	pri2, err := unmarshalNetworkKey(str)
	pid2, err := peer.IDFromPublicKey(pri2.GetPublic())
	t.Log(pid2)
	if err != nil {
		t.Fatal(err)
	}
	if pid1 != pid2 {
		t.Fatal("error")
	}
}
