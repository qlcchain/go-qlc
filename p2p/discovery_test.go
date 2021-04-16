package p2p

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"
)

func TestMDNS(t *testing.T) {

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19523"
	cfg1.P2P.Discovery.MDNSEnabled = true
	cfg1.P2P.Discovery.MDNSInterval = 1
	cfg1.P2P.IsBootNode = true
	cfg1.P2P.BootNodes = []string{"127.0.0.1:19533/dis1"}
	http.HandleFunc("/dis1/bootNode", func(w http.ResponseWriter, r *http.Request) {
		bootNode := cfg1.P2P.Listen + "/p2p/" + cfg1.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe("127.0.0.1:19533", nil); err != nil {
			t.Error(err)
			return
		}
	}()

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19524"
	cfg2.P2P.Discovery.MDNSEnabled = true
	cfg1.P2P.Discovery.MDNSInterval = 0
	cfg2.P2P.IsBootNode = true
	cfg2.P2P.BootNodes = []string{"127.0.0.1:19534/dis2"}
	http.HandleFunc("/dis2/bootNode", func(w http.ResponseWriter, r *http.Request) {
		bootNode := cfg2.P2P.Listen + "/p2p/" + cfg2.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe("127.0.0.1:19534", nil); err != nil {
			t.Error(err)
			return
		}
	}()

	//start node1
	node1, err := NewQlcService(dir1)
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//start node2
	node2, err := NewQlcService(dir2)
	err = node2.Start()
	if err != nil {
		t.Fatal(err)
	}

	//remove test file
	defer func() {
		err := node1.msgService.ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = node2.msgService.ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = node1.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = node2.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(filepath.Join(config.QlcTestDataDir(), "p2p"))
		if err != nil {
			t.Fatal(err)
		}
	}()

	//test local discovery
	ticker1 := time.NewTicker(30 * time.Second)
	ticker2 := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker1.C:
			//st.Fatal("local discovery error")
			return
		case <-ticker2.C:
			s := node1.node.streamManager.FindByPeerID(node2.node.cfg.P2P.ID.PeerID)
			if s != nil {
				return
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func TestNodeDiscovery(t *testing.T) {
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "discovery", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19736"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.P2P.IsBootNode = true
	cfg.P2P.BootNodes = []string{"127.0.0.1:19636/discovery"}
	http.HandleFunc("/discovery/bootNode", func(w http.ResponseWriter, r *http.Request) {
		bootNode := cfg.P2P.Listen + "/p2p/" + cfg.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe("127.0.0.1:19636", nil); err != nil {
			t.Error(err)
			return
		}
	}()

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "discovery", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19737"
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.BootNodes = []string{"http://127.0.0.1:19636/discovery/bootNode"}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "discovery", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19738"
	cfg2.P2P.Discovery.MDNSEnabled = false
	cfg2.P2P.BootNodes = []string{"http://127.0.0.1:19636/discovery/bootNode"}
	cfg2.P2P.Discovery.DiscoveryInterval = 1

	//start bootNode
	node, err := NewQlcService(dir)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	//start node1
	node1, err := NewQlcService(dir1)
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//start node2
	node2, err := NewQlcService(dir2)
	err = node2.Start()
	if err != nil {
		t.Fatal(err)
	}

	//remove test file
	defer func() {
		err := node.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = node1.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = node2.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = node.msgService.ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = node1.msgService.ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = node2.msgService.ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(filepath.Join(config.QlcTestDataDir(), "discovery"))
		if err != nil {
			t.Fatal(err)
		}
	}()

	//test remote peer discovery
	ticker1 := time.NewTicker(60 * time.Second)
	ticker2 := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker1.C:
			t.Fatal("find node error")
			return
		case <-ticker2.C:
			s := node1.node.streamManager.FindByPeerID(node2.node.cfg.P2P.ID.PeerID)
			if s != nil {
				node1.node.HandlePeerFound(node1.node.host.Peerstore().PeerInfo(node2.node.ID))
				return
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}
