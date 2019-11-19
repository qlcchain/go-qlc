package pov

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const TestPeerID1 = "peer1"
const TestPeerID2 = "peer2"

type povSyncMockData struct {
	eb     event.EventBus
	ledger ledger.Store
	chain  *povSyncChainReaderMockChain
}

type povSyncChainReaderMockChain struct {
	allBlocks    []*types.PovBlock
	heightBlocks map[uint64]*types.PovBlock
	hashBlocks   map[types.Hash]*types.PovBlock
	hashTDs      map[types.Hash]*types.PovTD
}

func (mc *povSyncChainReaderMockChain) InsertBlock(block *types.PovBlock, td *types.PovTD) {
	if _, ok := mc.hashBlocks[block.GetHash()]; ok {
		return
	}

	mc.hashBlocks[block.GetHash()] = block
	mc.heightBlocks[block.GetHeight()] = block
	mc.allBlocks = append(mc.allBlocks, block)

	mc.hashTDs[block.GetHash()] = td
}
func (mc *povSyncChainReaderMockChain) GenesisBlock() *types.PovBlock {
	return mc.allBlocks[0]
}
func (mc *povSyncChainReaderMockChain) LatestBlock() *types.PovBlock {
	return mc.allBlocks[len(mc.allBlocks)-1]
}
func (mc *povSyncChainReaderMockChain) GetBlockLocator(hash types.Hash) []*types.Hash {
	var hashes []*types.Hash

	gblk := common.GenesisPovBlock()
	bhash := gblk.GetHash()
	hashes = append(hashes, &bhash)
	return hashes
}
func (mc *povSyncChainReaderMockChain) LocateBestBlock(locator []*types.Hash) *types.PovBlock {
	for _, lh := range locator {
		for _, blk := range mc.allBlocks {
			if *lh == blk.GetHash() {
				return blk
			}
		}
	}
	return mc.allBlocks[0]
}
func (mc *povSyncChainReaderMockChain) GetBlockTDByHash(hash types.Hash) *types.PovTD {
	return mc.hashTDs[hash]
}

func setupPovSyncTestCase(t *testing.T) (func(t *testing.T), *povSyncMockData) {
	t.Parallel()

	md := &povSyncMockData{}

	uid := uuid.New().String()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uid)
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	md.ledger = ledger.NewLedger(cm.ConfigFile)

	md.eb = event.GetEventBus(uid)

	md.chain = new(povSyncChainReaderMockChain)
	md.chain.heightBlocks = make(map[uint64]*types.PovBlock)
	md.chain.hashBlocks = make(map[types.Hash]*types.PovBlock)
	md.chain.hashTDs = make(map[types.Hash]*types.PovTD)

	genBlk, genTD := mock.GenerateGenesisPovBlock()
	md.chain.InsertBlock(genBlk, genTD)

	return func(t *testing.T) {
		err := md.ledger.DBStore().Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}

		err = md.eb.Close()
		if err != nil {
			t.Fatal(err)
		}
	}, md
}

func TestPovSync_AddDelPeer1(t *testing.T) {
	teardownTestCase, md := setupPovSyncTestCase(t)
	defer teardownTestCase(t)

	povSync := NewPovSyncer(md.eb, md.ledger, md.chain)
	if povSync == nil {
		t.Fatal("NewPovSyncer is nil")
	}

	povSync.Start()

	peerID1 := TestPeerID1
	peerID2 := TestPeerID2
	povSync.onAddP2PStream(peerID1)
	povSync.onAddP2PStream(peerID2)

	bestPeer := povSync.GetBestPeer("")
	if bestPeer != nil {
		t.Fatalf("bestPeer should be nil")
	}

	genBlk, _ := mock.GenerateGenesisPovBlock()
	blk1, td1 := mock.GeneratePovBlock(genBlk, 0)
	blk2, td2 := mock.GeneratePovBlock(blk1, 0)

	peer1Status := new(protos.PovStatus)
	peer1Status.GenesisHash = genBlk.GetHash()
	peer1Status.CurrentHash = blk1.GetHash()
	peer1Status.CurrentHeight = blk1.GetHeight()
	peer1Status.CurrentTD = td1.Chain.Bytes()
	povSync.onPovStatus(peer1Status, peerID1)

	bestPeer = povSync.GetBestPeer("")
	if bestPeer == nil || bestPeer.peerID != peerID1 {
		t.Fatalf("bestPeer should be %s", peerID1)
	}

	peer2Status := new(protos.PovStatus)
	peer2Status.GenesisHash = genBlk.GetHash()
	peer2Status.CurrentHash = blk2.GetHash()
	peer2Status.CurrentHeight = blk2.GetHeight()
	peer2Status.CurrentTD = td2.Chain.Bytes()
	povSync.onPovStatus(peer2Status, peerID2)

	bestPeer = povSync.GetBestPeer("")
	if bestPeer == nil || bestPeer.peerID != peerID2 {
		t.Fatalf("bestPeer should be %s", peerID2)
	}

	retPeers1 := povSync.GetBestPeers(2)
	if len(retPeers1) != 2 {
		t.Fatalf("retPeers len not 2")
	}

	retPeers2 := povSync.GetRandomPeers(1)
	if len(retPeers2) != 1 {
		t.Fatalf("retPeers len not 1")
	}

	povSync.onDeleteP2PStream(peerID2)

	bestPeer = povSync.GetBestPeer("")
	if bestPeer == nil || bestPeer.peerID != peerID1 {
		t.Fatalf("bestPeer should be %s", peerID1)
	}

	povSync.onDeleteP2PStream(peerID1)

	bestPeer = povSync.GetBestPeer("")
	if bestPeer != nil {
		t.Fatalf("bestPeer should be nil")
	}

	povSync.Stop()
}

func TestPovSync_BulkPullReq1(t *testing.T) {
	teardownTestCase, md := setupPovSyncTestCase(t)
	defer teardownTestCase(t)

	povSync := NewPovSyncer(md.eb, md.ledger, md.chain)
	if povSync == nil {
		t.Fatal("NewPovSyncer is nil")
	}

	povSync.Start()

	peerID1 := TestPeerID1
	povSync.onAddP2PStream(peerID1)

	bestPeer := povSync.GetBestPeer("")
	if bestPeer != nil {
		t.Fatalf("bestPeer should be nil")
	}

	genBlk := md.chain.GenesisBlock()
	genHash := genBlk.GetHash()
	latestBlk := md.chain.LatestBlock()
	latestTD := md.chain.GetBlockTDByHash(latestBlk.GetHash())

	peer1Status := new(protos.PovStatus)
	peer1Status.GenesisHash = genBlk.GetHash()
	peer1Status.CurrentHash = latestBlk.GetHash()
	peer1Status.CurrentHeight = latestBlk.GetHeight()
	peer1Status.CurrentTD = latestTD.Chain.Bytes()
	povSync.onPovStatus(peer1Status, peerID1)

	bestPeer = povSync.GetBestPeer("")
	if bestPeer == nil || bestPeer.peerID != peerID1 {
		t.Fatalf("bestPeer should be %s", peerID1)
	}

	var rsp *protos.PovBulkPullRsp
	_ := md.eb.SubscribeSync(topic.EventSendMsgToSingle, func(msgType p2p.MessageType, msgData interface{}, toPeer string) {
		if msgType == p2p.PovBulkPullRsp {
			rsp = msgData.(*protos.PovBulkPullRsp)
		}
	})

	req1 := new(protos.PovBulkPullReq)
	req1.PullType = protos.PovPullTypeForward
	req1.Reason = protos.PovReasonSync
	req1.Locators = append(req1.Locators, &genHash)
	req1.Count = 1
	povSync.onPovBulkPullReq(req1, bestPeer.peerID)
	time.Sleep(time.Second)

	if rsp == nil {
		t.Fatalf("failed to get Message 1 msg")
	}
	if rsp.Count == 0 || rsp.Blocks[0].GetHash() != genBlk.GetHash() {
		t.Fatalf("failed to get Message 1 Count & Hash")
	}

	blk1, td1 := mock.GeneratePovBlock(genBlk, 0)
	blk1Hash := blk1.GetHash()
	md.chain.InsertBlock(blk1, td1)

	req2 := new(protos.PovBulkPullReq)
	req2.PullType = protos.PovPullTypeForward
	req2.Reason = protos.PovReasonSync
	req2.Locators = append(req2.Locators, &blk1Hash)
	req2.Count = 1
	povSync.onPovBulkPullReq(req2, bestPeer.peerID)
	time.Sleep(time.Second)

	if rsp == nil {
		t.Fatalf("failed to get Message 2 msg")
	}
	if rsp.Count == 0 || rsp.Blocks[0].GetHash() != blk1.GetHash() {
		t.Fatalf("failed to get Message 2 Count & Hash")
	}

	_ = md.eb.Unsubscribe(topic.EventSendMsgToSingle, id)
	povSync.Stop()
}

func TestPovSync_BulkPullRsp1(t *testing.T) {
	teardownTestCase, md := setupPovSyncTestCase(t)
	defer teardownTestCase(t)

	povSync := NewPovSyncer(md.eb, md.ledger, md.chain)
	if povSync == nil {
		t.Fatal("NewPovSyncer is nil")
	}

	povSync.Start()

	peerID1 := TestPeerID1
	povSync.onAddP2PStream(peerID1)

	blk1, td1 := mock.GeneratePovBlock(nil, 0)

	blk2, td2 := mock.GeneratePovBlock(blk1, 0)

	blk3, td3 := mock.GeneratePovBlock(blk2, 0)

	blk4, td4 := mock.GeneratePovBlock(blk3, 0)

	genBlk := md.chain.GenesisBlock()

	peer1Status := new(protos.PovStatus)
	peer1Status.GenesisHash = genBlk.GetHash()
	peer1Status.CurrentHash = blk4.GetHash()
	peer1Status.CurrentHeight = blk4.GetHeight()
	peer1Status.CurrentTD = td4.Chain.Bytes()
	povSync.onPovStatus(peer1Status, peerID1)

	var req *protos.PovBulkPullReq
	_ := md.eb.SubscribeSync(topic.EventSendMsgToSingle, func(msgType p2p.MessageType, msgData interface{}, toPeer string) {
		if msgType == p2p.PovBulkPullReq {
			req = msgData.(*protos.PovBulkPullReq)
		}
	})

	povSync.onPeriodicSyncTimer()

	if req == nil {
		t.Fatalf("failed to get PovBulkPullReq 1 msg")
	}
	if req.PullType != protos.PovPullTypeForward {
		t.Fatalf("failed to get PovBulkPullReq 1 PullType")
	}
	if req.Reason != protos.PovReasonSync {
		t.Fatalf("failed to get PovBulkPullReq 1 Reason")
	}
	if len(req.Locators) == 0 {
		t.Fatalf("failed to get PovBulkPullReq 1 Locators")
	}

	var syncBlocks []*types.PovBlock
	_ := md.eb.SubscribeSync(topic.EventPovRecvBlock, func(block *types.PovBlock, from types.PovBlockFrom, peer string) {
		syncBlocks = append(syncBlocks, block)
	})

	rsp := new(protos.PovBulkPullRsp)
	rsp.Reason = protos.PovReasonSync
	rsp.Blocks = append(rsp.Blocks, blk1, blk2, blk3, blk4)
	rsp.Count = uint32(len(rsp.Blocks))
	povSync.onPovBulkPullRsp(rsp, peerID1)

	time.Sleep(time.Second)
	povSync.onCheckChainTimer()

	md.chain.InsertBlock(blk1, td1)
	md.chain.InsertBlock(blk2, td2)
	md.chain.InsertBlock(blk3, td3)
	md.chain.InsertBlock(blk4, td4)

	time.Sleep(time.Second)
	povSync.onCheckChainTimer()

	_ = md.eb.Unsubscribe(topic.EventSendMsgToSingle, id)
	_ = md.eb.Unsubscribe(topic.EventPovRecvBlock, id2)
	povSync.Stop()
}

func TestPovSync_SimpleTest1(t *testing.T) {
	teardownTestCase, md := setupPovSyncTestCase(t)
	defer teardownTestCase(t)

	povSync := NewPovSyncer(md.eb, md.ledger, md.chain)
	if povSync == nil {
		t.Fatal("NewPovSyncer is nil")
	}

	povSync.Start()

	info := povSync.GetDebugInfo()
	if info == nil || len(info) == 0 {
		t.Fatal("debug info not exist")
	}

	povSync.Stop()
}
