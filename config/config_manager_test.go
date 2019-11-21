package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/qlcchain/go-qlc/common/util"
)

var configDir = filepath.Join(QlcTestDataDir(), "config")

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")

	return func(t *testing.T) {
		t.Log("teardown test case")
		err := os.RemoveAll(configDir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestConfigManager_Load(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	manager := NewCfgManager(configDir)
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
	t.Log(util.ToIndentString(cfg))
}

func TestConfigManager_parseVersion(t *testing.T) {
	manager := NewCfgManager(configDir)
	cfg, err := DefaultConfigV1(manager.ConfigDir())
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

func TestNewCfgManagerWithFile(t *testing.T) {
	cfgFile := filepath.Join(configDir, "test.json")
	defer func() {
		_ = os.Remove(cfgFile)
	}()
	cm := NewCfgManagerWithFile(cfgFile)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	} else {
		_ = cm.Save()
	}
}

func TestDefaultConfigV2(t *testing.T) {
	cm := NewCfgManager(configDir)
	_ = cm.createAndSave()
	cfg, err := DefaultConfigV2(cm.ConfigDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(cfg))
	bytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if version, err := cm.parseVersion(bytes); err != nil {
		t.Fatal(err)
	} else {
		if version != 2 {
			t.Fatal("invalid version", version)
		}
	}

	if dir, err := cm.ParseDataDir(); err != nil {
		t.Fatal(err)
	} else {
		if len(dir) == 0 {
			t.Fatal("invalid data dir")
		}
	}
}

func TestMigrationV1ToV2_Migration(t *testing.T) {
	manager := NewCfgManager(configDir)
	cfg, err := DefaultConfigV1(manager.ConfigDir())
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
	manager := NewCfgManager(configDir)
	cfg, err := DefaultConfig(manager.ConfigDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.ToIndentString(cfg))
}

func TestCfgManager_Load(t *testing.T) {
	manager := NewCfgManager(configDir)
	cfg1, err := DefaultConfigV1(manager.ConfigDir())
	if err != nil {
		t.Fatal(err)
	}

	err = manager.Save(cfg1)
	if err != nil {
		t.Fatal(err)
	}
	cfg3, err := manager.Load(NewMigrationV1ToV2(), NewMigrationV2ToV3(), NewMigrationV3ToV4())
	if err != nil {
		t.Fatal(err)
	}
	if !cfg3.P2P.Discovery.MDNSEnabled {
		t.Fatal("migration p2p error")
	}

	if cfg3.RPC.PublicModules == nil {
		t.Fatal("migration rpc error")
	}

	if len(cfg3.RPC.HttpVirtualHosts) == 0 {
		t.Fatal("invalid HttpVirtualHosts")
	}
}

func TestCfgManager_UpdateParams(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	params := []string{"rpc.rpcEnabled=true", "rpc.httpCors=localhost,localhost2", "p2p.syncInterval=200", "rpc.rpcEnabled="}
	cm := NewCfgManager(configDir)
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err = cm.UpdateParams(params)
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	if len(cfg.RPC.HTTPCors) != 2 || cfg.RPC.HTTPCors[0] != "localhost" || cfg.RPC.HTTPCors[1] != "localhost2" {
		t.Fatal("invalid rpc.httpCors", cfg.RPC.HTTPCors)
	}

	if cfg.P2P.SyncInterval != 200 {
		t.Fatal("invalid p2p.syncInterval", cfg.P2P.SyncInterval)
	}

	params = []string{"rpc.rpcEnabled=false", "rpc.httpCors=*"}
	cfg, err = cm.UpdateParams(params)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	if len(cfg.RPC.HTTPCors) == 0 || cfg.RPC.HTTPCors[0] != "*" {
		t.Fatal("invalid rpc.httpCors", cfg.RPC.HTTPCors)
	}
}

func Test_updateConfig(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	params := []string{"rpc.rpcEnabled=true", "rpc.httpCors=localhost,localhost2", "p2p.syncInterval=200", "rpc.rpcEnabled="}
	cm := NewCfgManager(configDir)
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err = cm.UpdateParams(params)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	if len(cfg.RPC.HTTPCors) != 2 || cfg.RPC.HTTPCors[0] != "localhost" || cfg.RPC.HTTPCors[1] != "localhost2" {
		t.Fatal("invalid rpc.httpCors", cfg.RPC.HTTPCors)
	}

	if cfg.P2P.SyncInterval != 200 {
		t.Fatal("invalid p2p.syncInterval", cfg.P2P.SyncInterval)
	}

	used, err := cm.Config()
	if err != nil {
		t.Fatal(err)
	}

	if used.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	if len(used.RPC.HTTPCors) != 1 || used.RPC.HTTPCors[0] != "*" {
		t.Fatal("invalid rpc.httpCors", used.RPC.HTTPCors)
	}

	if used.P2P.SyncInterval != 120 {
		t.Fatal("invalid p2p.syncInterval", used.P2P.SyncInterval)
	}

	err = cm.Commit()
	if err != nil {
		t.Fatal(err)
	}

	used2, err := cm.Config()
	if err != nil {
		t.Fatal(err)
	}

	if !used2.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	if len(used2.RPC.HTTPCors) != 2 || used2.RPC.HTTPCors[0] != "localhost" || used2.RPC.HTTPCors[1] != "localhost2" {
		t.Fatal("invalid rpc.httpCors", cfg.RPC.HTTPCors)
	}

	if used2.P2P.SyncInterval != 200 {
		t.Fatal("invalid p2p.syncInterval", used2.P2P.SyncInterval)
	}
}

func TestCfgManager_Discard(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	params := []string{"rpc.rpcEnabled=true", "rpc.httpCors=localhost,localhost2", "p2p.syncInterval=200", "rpc.rpcEnabled="}
	cm := NewCfgManager(configDir)
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err = cm.UpdateParams(params)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	if len(cfg.RPC.HTTPCors) != 2 || cfg.RPC.HTTPCors[0] != "localhost" || cfg.RPC.HTTPCors[1] != "localhost2" {
		t.Fatal("invalid rpc.httpCors", cfg.RPC.HTTPCors)
	}

	if cfg.P2P.SyncInterval != 200 {
		t.Fatal("invalid p2p.syncInterval", cfg.P2P.SyncInterval)
	}

	used, err := cm.Config()
	if err != nil {
		t.Fatal(err)
	}

	if used.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	if len(used.RPC.HTTPCors) != 1 || used.RPC.HTTPCors[0] != "*" {
		t.Fatal("invalid rpc.httpCors", used.RPC.HTTPCors)
	}

	if used.P2P.SyncInterval != 120 {
		t.Fatal("invalid p2p.syncInterval", used.P2P.SyncInterval)
	}

	cm.Discard()
	if cm.isDirty.Load() {
		t.Fatal("invalid is dirty")
	}
}

func TestCfgManager_CommitAndSave(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	params := []string{"rpc.rpcEnabled=true", "rpc.httpCors=localhost,localhost2", "p2p.syncInterval=200", "rpc.rpcEnabled="}
	cm := NewCfgManager(configDir)
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err = cm.UpdateParams(params)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	if len(cfg.RPC.HTTPCors) != 2 || cfg.RPC.HTTPCors[0] != "localhost" || cfg.RPC.HTTPCors[1] != "localhost2" {
		t.Fatal("invalid rpc.httpCors", cfg.RPC.HTTPCors)
	}

	if cfg.P2P.SyncInterval != 200 {
		t.Fatal("invalid p2p.syncInterval", cfg.P2P.SyncInterval)
	}
	err = cm.CommitAndSave()
	if err != nil {
		t.Fatal(err)
	}

	cm2 := NewCfgManager(configDir)
	cfg2, err := cm2.Load()
	if err != nil {
		t.Fatal(err)
	}

	if !cfg2.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	if len(cfg2.RPC.HTTPCors) != 2 || cfg2.RPC.HTTPCors[0] != "localhost" || cfg2.RPC.HTTPCors[1] != "localhost2" {
		t.Fatal("invalid rpc.httpCors", cfg2.RPC.HTTPCors)
	}

	if cfg2.P2P.SyncInterval != 200 {
		t.Fatal("invalid p2p.syncInterval", cfg2.P2P.SyncInterval)
	}
}

func TestCfgManager_DiffOther(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	cm := NewCfgManager(configDir)
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg1, _ := cfg.Clone()
	params := []string{"rpc.rpcEnabled=true", "rpc.httpCors=localhost,localhost2", "p2p.syncInterval=200", "rpc.rpcEnabled="}
	cfg, err = cm.UpdateParams(params)
	if err != nil {
		t.Fatal(err)
	}

	diff, err := cm.DiffOther(cfg1)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(diff)
	}
}

func TestCfgManager_Diff(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	cm := NewCfgManager(configDir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	params := []string{"rpc.rpcEnabled=true", "rpc.httpCors=localhost,localhost2", "p2p.syncInterval=200", "rpc.rpcEnabled="}
	_, err = cm.UpdateParams(params)
	if err != nil {
		t.Fatal(err)
	}

	diff, err := cm.Diff()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(diff)
	}
}

func TestCfgManager_PatchParams(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	params := []string{"rpc.rpcEnabled=true", "rpc.httpCors=localhost,localhost2", "p2p.syncInterval=200", "rpc.rpcEnabled="}
	cm := NewCfgManager(configDir)
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}

	cfg2, err := cm.PatchParams(params, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg2.RPC.Enable {
		t.Fatal("invalid rpc.rpcEnabled")
	}

	p1 := fmt.Sprintf("%p", cfg)
	p2 := fmt.Sprintf("%p", cfg2)
	if p2 != p1 {
		t.Fatal("invalid cfg")
	}
}
