package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
)

func TestConfig(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	cm.Load()
	cfg, err := cm.Config()
	if err != nil {
		t.Fatal(err)
	}
	token := cfg.Manager.AdminToken
	c := NewConfigApi(cm.ConfigFile)
	mark := "abc"
	if _, err := c.Update([]string{"pov.povEnabled==false"}, token, mark); err != nil {
		t.Fatal(err)
	}
	diff, err := c.Difference(token, mark)
	if err == nil {
		t.Fatal(err)
	}
	t.Log(diff)
	if b, err := c.Commit(token, mark); !b || err != nil {
		t.Fatal(err)
	}
	if b, err := c.Commit(token, mark); !b || err != nil {
		t.Fatal(err)
	}
	mark2 := "abcd"
	if _, err := c.Commit(token, mark2); err == nil {
		t.Fatal(err)
	}
	if _, err := c.Save(token, mark2); err == nil {
		t.Fatal(err)
	}
}
