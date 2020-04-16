/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
)

func TestNewBootNodeHttpService(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	params := []string{"p2p.bootNodeHttpServer=127.0.0.1:19363"}
	_, _ = cm.UpdateParams(params)
	_ = cm.CommitAndSave()
	hs := NewHttpService(cm.ConfigFile)
	err := hs.Init()
	if err != nil {
		t.Fatal(err)
	}
	if hs.State() != 2 {
		t.Fatal("init failed")
	}
	_ = hs.Start()
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	url := "http://" + cfg.P2P.BootNodeHttpServer + "/bootNode"
	rsp, err := http.Get(url)
	defer func() {
		_ = rsp.Body.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != (strings.ReplaceAll(hs.cfg.P2P.Listen, "0.0.0.0", hs.cfg.P2P.ListeningIp) + "/p2p/" + cfg.P2P.ID.PeerID) {
		t.Fatal("bootNode mismatch")
	}
	err = hs.Stop()
	if err != nil {
		t.Fatal(err)
	}
	if hs.Status() != 6 {
		t.Fatal("stop failed.")
	}
}
