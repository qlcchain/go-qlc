package p2p

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"
)

//func TestMDNS(t *testing.T) {
//	//node1 config
//	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
//	cfgFile1, _ := config.DefaultConfig(dir1)
//	cfgFile1.P2P.Listen = "/ip4/0.0.0.0/tcp/19523"
//	cfgFile1.P2P.Discovery.MDNSEnabled = true
//	cfgFile1.P2P.BootNodes = []string{}
//
//	//node2 config
//	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
//	cfgFile2, _ := config.DefaultConfig(dir2)
//	cfgFile2.P2P.Listen = "/ip4/0.0.0.0/tcp/19524"
//	cfgFile2.P2P.Discovery.MDNSEnabled = true
//	cfgFile2.P2P.BootNodes = []string{}
//
//	eventBus := event.New()
//	//start node1
//	node1, err := NewQlcService(cfgFile1, eventBus)
//	err = node1.Start()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	//start node2
//	node2, err := NewQlcService(cfgFile2, eventBus)
//	err = node2.Start()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	//remove test file
//	defer func() {
//		err := node1.msgService.ledger.Close()
//		if err != nil {
//			t.Fatal(err)
//		}
//		err = node2.msgService.ledger.Close()
//		if err != nil {
//			t.Fatal(err)
//		}
//		err = node1.Stop()
//		if err != nil {
//			t.Fatal(err)
//		}
//		err = node2.Stop()
//		if err != nil {
//			t.Fatal(err)
//		}
//		err = os.RemoveAll(config.QlcTestDataDir())
//		if err != nil {
//			t.Fatal(err)
//		}
//	}()
//
//	//test local discovery
//	ticker1 := time.NewTicker(60 * time.Second)
//	ticker2 := time.NewTicker(1 * time.Second)
//	for {
//		select {
//		case <-ticker1.C:
//			t.Fatal("local discovery error")
//			return
//		case <-ticker2.C:
//			s := node1.node.streamManager.FindByPeerID(node2.node.cfg.P2P.ID.PeerID)
//			if s != nil {
//				return
//			}
//		default:
//			time.Sleep(5 * time.Millisecond)
//		}
//	}
//}

func TestNodeDiscovery(t *testing.T) {
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile, _ := config.DefaultConfig(dir)
	cfgFile.P2P.Listen = "/ip4/0.0.0.0/tcp/19736"
	cfgFile.P2P.Discovery.MDNSEnabled = false
	cfgFile.P2P.BootNodes = []string{}
	b := "/ip4/0.0.0.0/tcp/19736/ipfs/" + cfgFile.P2P.ID.PeerID

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/0.0.0.0/tcp/19737"
	cfgFile1.P2P.Discovery.MDNSEnabled = false
	cfgFile1.P2P.BootNodes = []string{b}
	cfgFile1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile2, _ := config.DefaultConfig(dir2)
	cfgFile2.P2P.Listen = "/ip4/0.0.0.0/tcp/19738"
	cfgFile2.P2P.Discovery.MDNSEnabled = false
	cfgFile2.P2P.BootNodes = []string{b}
	cfgFile2.P2P.Discovery.DiscoveryInterval = 1

	//start bootNode
	node, err := NewQlcService(cfgFile)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	//start node1
	node1, err := NewQlcService(cfgFile1)
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//start node2
	node2, err := NewQlcService(cfgFile2)
	err = node2.Start()
	if err != nil {
		t.Fatal(err)
	}

	//remove test file
	defer func() {
		err := node.msgService.ledger.Close()
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
		err = node.Stop()
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
		err = os.RemoveAll(config.QlcTestDataDir())
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
				return
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}
