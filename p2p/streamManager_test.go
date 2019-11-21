package p2p

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/mock"

	"github.com/qlcchain/go-qlc/chain/context"

	"github.com/google/uuid"
	p2pmetrics "github.com/libp2p/go-libp2p-core/metrics"

	"github.com/qlcchain/go-qlc/config"
)

func Test_StreamManager(t *testing.T) {
	//bootNode config
	removeDir := filepath.Join(config.QlcTestDataDir(), "streamManager")
	dir := filepath.Join(config.QlcTestDataDir(), "streamManager", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19747"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.P2P.BootNodes = []string{}
	b := "/ip4/127.0.0.1/tcp/19747/ipfs/" + cfg.P2P.ID.PeerID

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "streamManager", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19748"
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "streamManager", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19749"
	cfg2.P2P.Discovery.MDNSEnabled = false
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
		err = os.RemoveAll(removeDir)
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
	removeDir := filepath.Join(config.QlcTestDataDir(), "peersInfo")
	dir := filepath.Join(config.QlcTestDataDir(), "peersInfo", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19750"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.P2P.BootNodes = []string{}
	b := "/ip4/127.0.0.1/tcp/19750/ipfs/" + cfg.P2P.ID.PeerID

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "peersInfo", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19751"
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "peersInfo", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19752"
	cfg2.P2P.Discovery.MDNSEnabled = false
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
		err = os.RemoveAll(removeDir)
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
	removeDir := filepath.Join(config.QlcTestDataDir(), "isConnect")
	dir := filepath.Join(config.QlcTestDataDir(), "isConnect", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19753"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.P2P.BootNodes = []string{}
	b := "/ip4/127.0.0.1/tcp/19753/ipfs/" + cfg.P2P.ID.PeerID

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "isConnect", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19754"
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "isConnect", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19755"
	cfg2.P2P.Discovery.MDNSEnabled = false
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
		err = os.RemoveAll(removeDir)
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

func TestStreamManager_lowestLatencyPeer(t *testing.T) {
	//bootNode config
	removeDir := filepath.Join(config.QlcTestDataDir(), "latency")
	dir := filepath.Join(config.QlcTestDataDir(), "latency", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19756"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.P2P.BootNodes = []string{}
	b := "/ip4/127.0.0.1/tcp/19756/ipfs/" + cfg.P2P.ID.PeerID

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "latency", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19757"
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "latency", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19758"
	cfg2.P2P.Discovery.MDNSEnabled = false
	cfg2.P2P.BootNodes = []string{b}
	cfg2.P2P.Discovery.DiscoveryInterval = 1

	//start bootNode
	node, err := NewQlcService(dir)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}
	_, err = node.node.streamManager.lowestLatencyPeer()
	if err == nil {
		t.Fatal("should be err")
	} else {
		if err != ErrNoStream {
			t.Fatal("err should be no stream")
		}
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
		err = os.RemoveAll(removeDir)
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
	peerId, err := node1.node.streamManager.lowestLatencyPeer()
	if err != nil {
		t.Fatal("err should not exist")
	}
	if peerId != node2.node.cfg.P2P.ID.PeerID {
		t.Fatal("peerId error")
	}
}

func TestGetBandwidthStats(t *testing.T) {
	//bootNode config
	removeDir := filepath.Join(config.QlcTestDataDir(), "bandwidth", uuid.New().String(), config.QlcConfigFile)
	dir := filepath.Join(config.QlcTestDataDir(), "bandwidth", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19762"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.P2P.BootNodes = []string{}
	b := "/ip4/127.0.0.1/tcp/19762/ipfs/" + cfg.P2P.ID.PeerID

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "bandwidth", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19760"
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "bandwidth", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19761"
	cfg2.P2P.Discovery.MDNSEnabled = false
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
		err = os.RemoveAll(removeDir)
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
	stats := new(p2pmetrics.Stats)
	blk := mock.StateBlockWithoutWork()
	node1.Broadcast(PublishReq, blk)
	time.Sleep(2 * time.Second)
	node1.node.GetBandwidthStats(stats)
	if stats.RateIn == 0 || stats.RateOut == 0 || stats.TotalIn == 0 || stats.TotalOut == 0 {
		t.Fatal("stats of bandWith error 2")
	}
	if (int64(stats.RateIn) > stats.TotalIn) || (int64(stats.RateOut) > stats.TotalOut) {
		t.Fatal("stats of bandWith error 3")
	}
}
