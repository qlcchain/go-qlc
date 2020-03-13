package p2p

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
)

func TestQlcNode(t *testing.T) {
	cfgFile := filepath.Join(config.QlcTestDataDir(), "node", uuid.New().String())
	defer func() {
		err := os.RemoveAll(filepath.Join(config.QlcTestDataDir(), "node"))
		if err != nil {
			t.Fatal(err)
		}
	}()
	cfg, err := config.NewCfgManager(cfgFile).Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/18888"
	cfg.P2P.IsBootNode = true
	cfg.P2P.BootNodes = []string{"127.0.0.1:18889/msg2"}
	http.HandleFunc("/msg2/bootNode", func(w http.ResponseWriter, r *http.Request) {
		bootNode := cfg.P2P.Listen + "/p2p/" + cfg.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe("127.0.0.1:18889", nil); err != nil {
			t.Fatal(err)
		}
	}()
	node, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = node.buildHost()
	if err != nil {
		t.Fatal(err)
	}
}
