package config

import (
	"github.com/json-iterator/go"
	"github.com/libp2p/go-libp2p-peer"
	"os"
	"testing"
)

var cfgFile = DefaultConfigFile()

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")

	return func(t *testing.T) {
		t.Log("teardown test case")
		err := os.RemoveAll(cfgFile)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestConfigManager_Load(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	manager := NewCfgManager(cfgFile)
	cfg, err := manager.Load()
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
	bytes, err := jsoniter.Marshal(cfg)
	t.Log(string(bytes))
}
