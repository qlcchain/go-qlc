package p2p

import (
	"github.com/qlcchain/go-qlc/chain/context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/config"
)

func Test_StreamManager(t *testing.T) {
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19747"
	cfg.P2P.BootNodes = []string{}
	b := "/ip4/127.0.0.1/tcp/19747/ipfs/" + cfg.P2P.ID.PeerID

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19748"
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19749"
	cfg2.P2P.BootNodes = []string{b}
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

	ticker1 := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker1.C:
			t.Fatal("find node error")
			return
		default:
		}

		s := node1.node.streamManager.FindByPeerID(node2.node.cfg.P2P.ID.PeerID)
		if s != nil {
			if s.IsConnected() {
				break
			}
		}
	}

	if node1.node.streamManager.PeerCounts() != 1 {
		t.Fatal("peer1 count error")
	}

	s := node1.node.streamManager.FindByPeerID(node2.node.ID.Pretty())
	if s == nil {
		t.Fatal("find peer2 error")
	}
	p1, err := node1.node.streamManager.RandomPeer()
	if p1 != cfg2.P2P.ID.PeerID || err != nil {
		t.Fatal("node1 random peer error")
	}
	node1.node.streamManager.RemoveStream(s)
	if node1.node.streamManager.FindByPeerID(node2.node.ID.Pretty()) != nil {
		t.Fatal("node1 RemoveStream error")
	}
	node1.node.streamManager.createStreamWithPeer(node2.node.ID)
	if node1.node.streamManager.FindByPeerID(node2.node.ID.Pretty()) == nil {
		t.Fatal("node1 create Stream With node2 error")
	}
}

func TestStreamManager_GetAllConnectPeersInfo(t *testing.T) {
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19747"
	cfg.P2P.BootNodes = []string{}
	b := "/ip4/127.0.0.1/tcp/19747/ipfs/" + cfg.P2P.ID.PeerID

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19748"
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19749"
	cfg2.P2P.BootNodes = []string{b}
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

	ticker1 := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker1.C:
			t.Fatal("find node error")
			return
		default:
		}

		s := node1.node.streamManager.FindByPeerID(node2.node.cfg.P2P.ID.PeerID)
		if s != nil {
			if s.IsConnected() {
				break
			}
		}
	}
	p := make(map[string]string)
	node1.node.streamManager.GetAllConnectPeersInfo(p)
	if len(p) != 1 {
		t.Fatal("get info error")
	}
}

func TestStreamManager_IsConnectWithPeerId(t *testing.T) {
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19747"
	cfg.P2P.BootNodes = []string{}
	b := "/ip4/127.0.0.1/tcp/19747/ipfs/" + cfg.P2P.ID.PeerID

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19748"
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19749"
	cfg2.P2P.BootNodes = []string{b}
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

	ticker1 := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker1.C:
			t.Fatal("find node error")
			return
		default:
		}

		s := node1.node.streamManager.FindByPeerID(node2.node.cfg.P2P.ID.PeerID)
		if s != nil {
			if s.IsConnected() {
				break
			}
		}
	}
	c := node1.node.streamManager.IsConnectWithPeerId(node2.node.cfg.P2P.ID.PeerID)
	if !c {
		t.Fatal("should be connect")
	}
}
