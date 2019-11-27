/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
)

func TestNewP2PService(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	cfg, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	http.HandleFunc("/ps/bootNode", func(w http.ResponseWriter, r *http.Request) {
		bootNode := cfg.P2P.Listen + "/ipfs/" + cfg.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe("127.0.0.1:19362", nil); err != nil {
			t.Fatal(err)
		}
	}()
	params := []string{"p2p.bootNode=127.0.0.1:19362/ps"}
	_, _ = cm.UpdateParams(params)
	_ = cm.CommitAndSave()
	p, err := NewP2PService(cm.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Init()
	if err != nil {
		t.Fatal(err)
	}
	if p.State() != 2 {
		t.Fatal("p2p init failed")
	}
	err = p.Start()
	if err != nil {
		t.Fatal(err)
	}
	if p.State() != 4 {
		t.Fatal("p2p start failed")
	}
	err = p.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if p.Status() != 6 {
		t.Fatal("stop failed.")
	}
}
