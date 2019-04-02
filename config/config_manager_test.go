package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common/util"

	peer "github.com/libp2p/go-libp2p-peer"
)

var cfgFile = filepath.Join(QlcTestDataDir(), "config")

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

	if cfg.PerformanceEnabled {
		t.Fatal("Performance test config error")
	}
	pri, err := cfg.DecodePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	ID, err := peer.IDFromPublicKey(pri.GetPublic())
	if err != nil {
		t.Fatal(err)
	}
	if ID.Pretty() != cfg.P2P.ID.PeerID {
		t.Fatal("peer id error")
	}
	bytes, err := json.Marshal(cfg)
	t.Log(string(bytes))
}

func TestConfigManager_parseVersion(t *testing.T) {
	manager := NewCfgManager(cfgFile)
	cfg, err := DefaultConfigV1(manager.cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	bytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if version, err := manager.parseVersion(bytes); err != nil {
		t.Fatal(err)
	} else {
		t.Log(version)
	}
}

func TestDefaultConfigV2(t *testing.T) {
	manager := NewCfgManager(cfgFile)
	cfg, err := DefaultConfigV2(manager.cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(cfg))
	bytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if version, err := manager.parseVersion(bytes); err != nil {
		t.Fatal(err)
	} else {
		if version != 2 {
			t.Fatal("invalid version", version)
		}
	}
}

func TestMigrationV1ToV2_Migration(t *testing.T) {
	manager := NewCfgManager(cfgFile)
	cfg, err := DefaultConfigV1(manager.cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.AutoGenerateReceive = true

	bytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	m := NewMigrationV1ToV2()
	data, v2, err := m.Migration(bytes, 1)

	if err != nil {
		t.Fatal(err)
	}

	if v2 != 2 {
		t.Fatal("invalid version")
	}

	var cfg2 ConfigV2
	err = json.Unmarshal(data, &cfg2)
	if err != nil {
		t.Fatal(err)
	}

	if !cfg2.AutoGenerateReceive {
		t.Fatal("migration failed.")
	}
}

func TestDefaultConfig(t *testing.T) {
	manager := NewCfgManager(cfgFile)
	cfg, err := DefaultConfig(manager.cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(cfg))
}

func TestCfgManager_Load(t *testing.T) {
	manager := NewCfgManager(cfgFile)
	cfg1, err := DefaultConfigV1(manager.cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	err = manager.save(cfg1)
	if err != nil {
		t.Fatal(err)
	}
	cfg2, err := manager.Load(NewMigrationV1ToV2())
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.P2P.Discovery.MDNSEnabled {
		t.Fatal("migration p2p error")
	}

	if cfg2.RPC.PublicModules == nil {
		t.Fatal("migration rpc error")
	}
}
