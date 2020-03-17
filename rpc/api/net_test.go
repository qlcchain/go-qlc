package api

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/mock"

	"github.com/google/uuid"
	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

type peersCount struct {
	Connect int `json:"connect"`
	Online  int `json:"online"`
	All     int `json:"all"`
}

func setupTestCaseNet(t *testing.T) (func(t *testing.T), *ledger.Ledger, *NetApi) {
	t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	_ = cc.Init(nil)
	l := ledger.NewLedger(cm.ConfigFile)
	netApi := NewNetApi(l, cc.EventBus(), cc)
	return func(t *testing.T) {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
		_ = cc.Stop()
	}, l, netApi
}

func TestNetApi_ConnectPeersInfo(t *testing.T) {
	teardownTestCase, _, netApi := setupTestCaseNet(t)
	defer teardownTestCase(t)
	var peersInfo []*types.PeerInfo
	peer1 := &types.PeerInfo{PeerID: "QmYPq8Cqqfyhaj6pKCiCMVX3KFRMZwi4w6fU6wGLU2T9JC"}
	peer2 := &types.PeerInfo{PeerID: "QmfMSZSGBaLobW6WKzqaVhXnbVg8kJEaRbWyEfsxi94dMw"}
	peersInfo = append(peersInfo, peer1, peer2)
	netApi.eb.Publish(topic.EventPeersInfo, &topic.EventP2PConnectPeersMsg{PeersInfo: peersInfo})
	time.Sleep(100 * time.Millisecond)
	var offset *int
	peers, err := netApi.ConnectPeersInfo(-1, offset)
	if err == nil {
		t.Fatal("should return count error")
	}
	peers, err = netApi.ConnectPeersInfo(2, offset)
	if err != nil {
		t.Fatal(err)
	}
	if len(peers) != 2 {
		t.Fatalf("connect peers info error,want 2,have %d", len(peers))
	}
	if peers[0].PeerID != peer1.PeerID || peers[1].PeerID != peer2.PeerID {
		t.Fatal("connect peers info error")
	}
}

func TestNetApi_GetAllPeersInfo(t *testing.T) {
	teardownTestCase, l, netApi := setupTestCaseNet(t)
	defer teardownTestCase(t)
	var peersInfo []*types.PeerInfo
	peer1 := &types.PeerInfo{PeerID: "QmYPq8Cqqfyhaj6pKCiCMVX3KFRMZwi4w6fU6wGLU2T9JC"}
	peer2 := &types.PeerInfo{PeerID: "QmfMSZSGBaLobW6WKzqaVhXnbVg8kJEaRbWyEfsxi94dMw"}
	peer3 := &types.PeerInfo{PeerID: "QmUGgsPH6kaZFHCy392vBDYqZC3HPexewuZrSoEHgNaHYz"}
	peer4 := &types.PeerInfo{PeerID: "QmToDLHJJ8mgSZcWewTyWBiArqmNMfLhGGkJpA5E62M6yW"}
	peer5 := &types.PeerInfo{PeerID: "QmU7NP5C9LnjoR4XeWvS9KdiLMAJG3ocyTb1bgtZzDfNWZ"}
	peersInfo = append(peersInfo, peer1, peer2, peer3, peer4, peer5)
	for _, v := range peersInfo {
		_ = l.AddPeerInfo(v)
	}
	var offset *int
	peers, err := netApi.GetAllPeersInfo(-1, offset)
	if err == nil {
		t.Fatal("should return count error")
	}
	peers, err = netApi.GetAllPeersInfo(5, offset)
	if err != nil {
		t.Fatal(err)
	}
	if len(peers) != 5 {
		t.Fatalf("all peers info error,want 5,have %d", len(peers))
	}
}

func TestNetApi_GetOnlinePeersInfo(t *testing.T) {
	teardownTestCase, _, netApi := setupTestCaseNet(t)
	defer teardownTestCase(t)
	var peersInfo []*types.PeerInfo
	peer1 := &types.PeerInfo{PeerID: "QmYPq8Cqqfyhaj6pKCiCMVX3KFRMZwi4w6fU6wGLU2T9JC"}
	peer2 := &types.PeerInfo{PeerID: "QmfMSZSGBaLobW6WKzqaVhXnbVg8kJEaRbWyEfsxi94dMw"}
	peer3 := &types.PeerInfo{PeerID: "QmUGgsPH6kaZFHCy392vBDYqZC3HPexewuZrSoEHgNaHYz"}
	peersInfo = append(peersInfo, peer1, peer2, peer3)
	netApi.eb.Publish(topic.EventOnlinePeersInfo, &topic.EventP2POnlinePeersMsg{PeersInfo: peersInfo})
	time.Sleep(100 * time.Millisecond)
	var offset *int
	peers, err := netApi.GetOnlinePeersInfo(-1, offset)
	if err == nil {
		t.Fatal("should return count error")
	}
	peers, err = netApi.GetOnlinePeersInfo(3, offset)
	if err != nil {
		t.Fatal(err)
	}
	if len(peers) != 3 {
		t.Fatalf("online peers info error,want 3,have %d", len(peers))
	}
	if peers[0].PeerID != peer1.PeerID || peers[1].PeerID != peer2.PeerID || peers[2].PeerID != peer3.PeerID {
		t.Fatal("connect peers info error")
	}
}

func TestNetApi_PeersCount(t *testing.T) {
	teardownTestCase, l, netApi := setupTestCaseNet(t)
	defer teardownTestCase(t)
	var online, connect, all []*types.PeerInfo
	peer1 := &types.PeerInfo{PeerID: "QmYPq8Cqqfyhaj6pKCiCMVX3KFRMZwi4w6fU6wGLU2T9JC"}
	peer2 := &types.PeerInfo{PeerID: "QmfMSZSGBaLobW6WKzqaVhXnbVg8kJEaRbWyEfsxi94dMw"}
	peer3 := &types.PeerInfo{PeerID: "QmUGgsPH6kaZFHCy392vBDYqZC3HPexewuZrSoEHgNaHYz"}
	peer4 := &types.PeerInfo{PeerID: "QmToDLHJJ8mgSZcWewTyWBiArqmNMfLhGGkJpA5E62M6yW"}
	peer5 := &types.PeerInfo{PeerID: "QmU7NP5C9LnjoR4XeWvS9KdiLMAJG3ocyTb1bgtZzDfNWZ"}
	online = append(online, peer1, peer2, peer3)
	connect = append(connect, peer1, peer2)
	all = append(all, peer1, peer2, peer3, peer4, peer5)
	for _, v := range all {
		_ = l.AddPeerInfo(v)
	}
	netApi.eb.Publish(topic.EventOnlinePeersInfo, &topic.EventP2POnlinePeersMsg{PeersInfo: online})
	netApi.eb.Publish(topic.EventPeersInfo, &topic.EventP2PConnectPeersMsg{PeersInfo: connect})
	time.Sleep(100 * time.Millisecond)
	pc, err := netApi.PeersCount()
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(pc)
	if err != nil {
		t.Fatal(err)
	}
	var p peersCount
	err = json.Unmarshal(b, &p)
	if err != nil {
		t.Fatal(err)
	}
	if p.Connect != 2 || p.Online != 3 || p.All != 5 {
		t.Fatal("peers count error")
	}
}

func TestNetApi_GetBandwidthStats(t *testing.T) {
	teardownTestCase, _, netApi := setupTestCaseNet(t)
	defer teardownTestCase(t)
	bwState := &topic.EventBandwidthStats{
		TotalIn:  100000,
		TotalOut: 200000,
		RateIn:   10000,
		RateOut:  20000,
	}
	netApi.eb.Publish(topic.EventGetBandwidthStats, bwState)
	time.Sleep(100 * time.Millisecond)
	bs := netApi.GetBandwidthStats()
	if bs.RateIn != 10000 || bs.RateOut != 20000 || bs.TotalIn != 100000 || bs.TotalOut != 200000 {
		t.Fatal("bandWith stat error")
	}
}

func TestNetApi_Syncing(t *testing.T) {
	teardownTestCase, _, netApi := setupTestCaseNet(t)
	defer teardownTestCase(t)
	netApi.eb.Publish(topic.EventSyncStateChange, &topic.EventP2PSyncStateMsg{P2pSyncState: topic.Syncing})
	time.Sleep(100 * time.Millisecond)
	b := netApi.Syncing()
	if !b {
		t.Fatal("sync state error1")
	}
	netApi.eb.Publish(topic.EventSyncStateChange, &topic.EventP2PSyncStateMsg{P2pSyncState: topic.SyncDone})
	time.Sleep(100 * time.Millisecond)
	b = netApi.Syncing()
	if !b {
		t.Fatal("sync state error2")
	}
	netApi.eb.Publish(topic.EventSyncStateChange, &topic.EventP2PSyncStateMsg{P2pSyncState: topic.SyncFinish})
	time.Sleep(100 * time.Millisecond)
	b = netApi.Syncing()
	if b {
		t.Fatal("sync state error3")
	}
}

func TestNetApi_OnlineRepresentatives(t *testing.T) {
	teardownTestCase, l, netApi := setupTestCaseNet(t)
	defer teardownTestCase(t)
	var addrs []*types.Address
	addr1 := mock.Address()
	addr2 := mock.Address()
	addr3 := mock.Address()
	addrs = append(addrs, &addr1, &addr2, &addr3)
	err := l.SetOnlineRepresentations(addrs)
	if err != nil {
		t.Fatal(err)
	}
	addresses := netApi.OnlineRepresentatives()
	if len(addresses) != 3 {
		t.Fatalf("online representatives count err,want 3,have %d", len(addresses))
	}
}

func TestNetApi_OnlineRepsInfo(t *testing.T) {
	teardownTestCase, l, netApi := setupTestCaseNet(t)
	defer teardownTestCase(t)
	address := mock.Address()
	ac := mock.AccountMeta(address)
	ac.CoinBalance = types.Balance{Int: big.NewInt(int64(20000000000000000))}
	ac.CoinVote = types.Balance{Int: big.NewInt(int64(10000000000000000))}
	benefit := &types.Benefit{
		Vote:    ac.CoinVote,
		Storage: ac.CoinStorage,
		Network: ac.CoinNetwork,
		Oracle:  ac.CoinOracle,
		Balance: ac.CoinBalance,
		Total:   ac.TotalBalance(),
	}
	var addrs []*types.Address
	addrs = append(addrs, &address)
	err := l.SetOnlineRepresentations(addrs)
	if err != nil {
		t.Fatal(err)
	}
	err = l.AddRepresentation(address, benefit, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}
	or := netApi.OnlineRepsInfo()
	if len(or.Reps) != 1 || or.ValidVotesPercent != "50.00%" || or.ValidVotes.Compare(types.Balance{Int: big.NewInt(int64(30000000000000000))}) != 0 {
		t.Fatal("online info error")
	}
}
