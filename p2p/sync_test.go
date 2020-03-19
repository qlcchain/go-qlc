package p2p

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"

	"github.com/qlcchain/go-qlc/mock"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"
)

func TestServiceSync(t *testing.T) {
	//bootNode config
	removeDir := filepath.Join(config.QlcTestDataDir(), "sync")
	dir := filepath.Join(config.QlcTestDataDir(), "sync", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/29747"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.P2P.IsBootNode = true
	cfg.P2P.BootNodes = []string{"127.0.0.1:29647/sync"}
	http.HandleFunc("/sync/bootNode", func(w http.ResponseWriter, r *http.Request) {
		bootNode := cfg.P2P.Listen + "/p2p/" + cfg.P2P.ID.PeerID
		_, _ = fmt.Fprintf(w, bootNode)
	})
	go func() {
		if err := http.ListenAndServe(":29647", nil); err != nil {
			t.Fatal(err)
		}
	}()

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "sync", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/29748"
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.BootNodes = []string{"127.0.0.1:29647/sync"}
	cfg1.P2P.Discovery.DiscoveryInterval = 1
	cfg1.P2P.SyncInterval = 5

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "sync", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/29749"
	cfg2.P2P.Discovery.MDNSEnabled = false
	cfg2.P2P.BootNodes = []string{"127.0.0.1:29647/sync"}
	cfg2.P2P.Discovery.DiscoveryInterval = 1
	cfg1.P2P.SyncInterval = 4

	//start bootNode
	node, err := NewQlcService(dir)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	//start node1
	node1, err := NewQlcService(dir1)
	err = node1.cc.Register(context.ConsensusService, nil)
	if err != nil {
		t.Fatal("register consensus service error")
	}
	l1 := node1.msgService.syncService.qlcLedger
	lv1 := process.NewLedgerVerifier(l1)
	bs1, err := mock.BlockChain(false)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(config.QlcTestDataDir(), "blocks.json")
	_, err = os.Stat(path)
	if err == nil {
		_ = os.Remove(path)
	}
	if err := lv1.BlockProcess(bs1[0]); err != nil {
		t.Fatal(err)
	}
	t.Log("bs1 hash", bs1[0].GetHash())
	for _, b := range bs1[1:] {
		if p, err := lv1.Process(b); err != nil || p != process.Progress {
			t.Fatal(p, err)
		}
	}
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//start node2
	node2, err := NewQlcService(dir2)
	err = node2.cc.Register(context.ConsensusService, nil)
	if err != nil {
		t.Fatal("register consensus service error")
	}
	l2 := node2.msgService.syncService.qlcLedger
	lv2 := process.NewLedgerVerifier(l2)
	bs2, err := mock.BlockChain(false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(path)
	if err == nil {
		_ = os.Remove(path)
	}
	if err := lv2.BlockProcess(bs2[0]); err != nil {
		t.Fatal(err)
	}
	t.Log("bs2 hash", bs2[0].GetHash())
	for _, b := range bs2[1:] {
		if p, err := lv2.Process(b); err != nil || p != process.Progress {
			t.Fatal(p, err)
		}
	}
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
	node1.msgService.syncService.requestFrontiersFromPov(node2.node.cfg.P2P.ID.PeerID)
	var hashs []*types.Hash
	hash1 := bs2[0].GetHash()
	hash2 := bs2[1].GetHash()
	hash3 := bs2[2].GetHash()
	hash4 := bs2[3].GetHash()
	hash5 := bs2[4].GetHash()
	hash6 := bs2[5].GetHash()
	hashs = append(hashs, &hash1, &hash2, &hash3, &hash4, &hash5, &hash6)
	node1.msgService.syncService.requestTxsByHashes(hashs, node2.node.cfg.P2P.ID.PeerID)
	node1.msgService.syncService.onConsensusSyncFinished()
	time.Sleep(10 * time.Second)
}
