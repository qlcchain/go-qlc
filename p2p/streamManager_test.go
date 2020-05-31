package p2p

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	p2pmetrics "github.com/libp2p/go-libp2p-core/metrics"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func Test_StreamManager(t *testing.T) {
	//bootNode config
	removeDir := filepath.Join(config.QlcTestDataDir(), "streamManager")
	dir := filepath.Join(config.QlcTestDataDir(), "streamManager", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19747"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.P2P.IsBootNode = true
	cfg.P2P.BootNodes = []string{"http://127.0.0.1:19647/stream/bootNode"}
	http.HandleFunc("/stream/bootNode", func(w http.ResponseWriter, r *http.Request) {
		bootNode := cfg.P2P.Listen + "/p2p/" + cfg.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe(":19647", nil); err != nil {
			t.Fatal(err)
		}
	}()

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "streamManager", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19748"
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.BootNodes = []string{"http://127.0.0.1:19647/stream/bootNode"}
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "streamManager", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19749"
	cfg2.P2P.Discovery.MDNSEnabled = false
	cfg2.P2P.BootNodes = []string{"http://127.0.0.1:19647/stream/bootNode"}
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
	var p []*types.PeerInfo
	node1.node.streamManager.GetAllConnectPeersInfo(&p)
	if len(p) != 1 {
		t.Fatal("get info error")
	}
	node1.node.streamManager.allStreams.Range(func(key, value interface{}) bool {
		stream := value.(*Stream)
		t.Log(stream.String())
		return true
	})
	var p1 []*types.PeerInfo
	node1.node.streamManager.GetOnlinePeersInfo(&p1)
	if len(p1) != 2 {
		return
	}
	c := node1.node.streamManager.IsConnectWithPeerId(node2.node.cfg.P2P.ID.PeerID)
	if !c {
		t.Fatal("should be connect")
	}
	peerId, err := node1.node.streamManager.lowestLatencyPeer()
	if err != nil {
		t.Fatal("err should not exist")
	}
	if peerId != node2.node.cfg.P2P.ID.PeerID {
		t.Fatal("peerId error")
	}
	s1, err := node1.node.streamManager.randomLowerLatencyPeer()
	if err != nil {
		t.Fatal("err should not exist")
	}
	if s1 != node2.node.cfg.P2P.ID.PeerID {
		t.Fatal("randomLowerLatencyPeer error")
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
	if node1.node.streamManager.PeerCounts() != 1 {
		t.Fatal("peer1 count error")
	}

	s := node1.node.streamManager.FindByPeerID(node2.node.ID.Pretty())
	if s == nil {
		t.Fatal("find peer2 error")
	}
	p2, err := node1.node.streamManager.RandomPeer()
	if p2 != cfg2.P2P.ID.PeerID || err != nil {
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
