package config

import (
	"testing"

	"github.com/libp2p/go-libp2p-peer"
)

func TestInitConfig(t *testing.T) {
	cfg, err := InitConfig()
	if err != nil {
		t.Fatal(err)
	}
	pri, err := cfg.DecodePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	ID, err := peer.IDFromPublicKey(pri.GetPublic())
	if err != nil {
		t.Fatal(err)
	}
	if ID.Pretty() != cfg.ID.PeerID {
		t.Fatal("peer id error")
	}
}
