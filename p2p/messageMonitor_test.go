package p2p

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/chain/context"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

func Test_MessageService_Stop(t *testing.T) {
	//node config
	dir1 := filepath.Join(config.QlcTestDataDir(), uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg, _ := cc1.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19739"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.P2P.BootNodes = []string{}

	//start node
	node, err := NewQlcService(dir1)
	if node == nil {
		t.Fatal(err)
	}
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	_, ok := node.dispatcher.subscribersMap.Load(PublishReq)
	if !ok {
		t.Fatal("subscription PublishReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(ConfirmReq)
	if !ok {
		t.Fatal("subscription ConfirmReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(ConfirmAck)
	if !ok {
		t.Fatal("subscription ConfirmAck messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(FrontierRequest)
	if !ok {
		t.Fatal("subscription FrontierRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(FrontierRsp)
	if !ok {
		t.Fatal("subscription FrontierRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(BulkPullRequest)
	if !ok {
		t.Fatal("subscription BulkPullRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(BulkPullRsp)
	if !ok {
		t.Fatal("subscription BulkPullRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(BulkPushBlock)
	if !ok {
		t.Fatal("subscription BulkPushBlock messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageResponse)
	if !ok {
		t.Fatal("subscription MessageResponse messageType error")
	}

	node.msgService.Stop()
	_, ok = node.dispatcher.subscribersMap.Load(PublishReq)
	if ok {
		t.Fatal("subscription PublishReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(ConfirmReq)
	if ok {
		t.Fatal("subscription ConfirmReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(ConfirmAck)
	if ok {
		t.Fatal("subscription ConfirmAck messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(FrontierRequest)
	if ok {
		t.Fatal("subscription FrontierRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(FrontierRsp)
	if ok {
		t.Fatal("subscription FrontierRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(BulkPullRequest)
	if ok {
		t.Fatal("subscription BulkPullRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(BulkPullRsp)
	if ok {
		t.Fatal("subscription BulkPullRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(BulkPushBlock)
	if ok {
		t.Fatal("subscription BulkPushBlock messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageResponse)
	if ok {
		t.Fatal("subscription MessageResponse messageType error")
	}

	//remove test file
	defer func() {
		err := node.msgService.ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = node.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir1)
		if err != nil {
			t.Fatal(err)
		}
	}()
}

func Test_MarshalMessage(t *testing.T) {
	blk := mock.StateBlockWithoutWork()
	data1, err := marshalMessage(PublishReq, blk)
	if err != nil {
		t.Fatal("Marshal PublishReq err1")
	}
	pushBlock := protos.PublishBlock{
		Blk: blk,
	}
	data2, err := protos.PublishBlockToProto(&pushBlock)
	if err != nil {
		t.Fatal("Marshal PublishReq err2")
	}
	if bytes.Compare(data1, data2) != 0 {
		t.Fatal("Marshal PublishReq err3")
	}
	ConfirmReqBlocks := make([]*types.StateBlock, 0)
	ConfirmReqBlocks = append(ConfirmReqBlocks, blk)
	data3, err := marshalMessage(ConfirmReq, ConfirmReqBlocks)
	if err != nil {
		t.Fatal("Marshal ConfirmReq err1")
	}
	packet := &protos.ConfirmReqBlock{
		Blk: ConfirmReqBlocks,
	}
	data4, err := protos.ConfirmReqBlockToProto(packet)
	if err != nil {
		t.Fatal("Marshal ConfirmReq err2")
	}
	if bytes.Compare(data3, data4) != 0 {
		t.Fatal("Marshal ConfirmReq err3")
	}
	var va protos.ConfirmAckBlock
	a := mock.Account()
	va.Sequence = 0
	va.Hash = append(va.Hash, blk.GetHash())
	va.Account = a.Address()
	va.Signature = a.Sign(blk.GetHash())
	data5, err := marshalMessage(ConfirmAck, &va)
	if err != nil {
		t.Fatal("Marshal ConfirmAck err1")
	}
	data6, err := protos.ConfirmAckBlockToProto(&va)
	if err != nil {
		t.Fatal("Marshal ConfirmAck err2")
	}
	if bytes.Compare(data5, data6) != 0 {
		t.Fatal("Marshal ConfirmAck err3")
	}
	address := types.Address{}
	Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	data7, err := marshalMessage(FrontierRequest, Req)
	if err != nil {
		t.Fatal("Marshal FrontierRequest err1")
	}
	data8, err := protos.FrontierReqToProto(Req)
	if err != nil {
		t.Fatal("Marshal FrontierRequest err2")
	}
	if bytes.Compare(data7, data8) != 0 {
		t.Fatal("Marshal FrontierRequest err3")
	}
	zeroFrontier := new(types.Frontier)
	frb := &types.FrontierBlock{
		Fr:        zeroFrontier,
		HeaderBlk: blk,
	}
	frs := &protos.FrontierResponse{}

	frs.Fs = append(frs.Fs, frb)
	data9, err := marshalMessage(FrontierRsp, frs)
	if err != nil {
		t.Fatal("Marshal FrontierRsp err1")
	}
	data10, err := protos.FrontierResponseToProto(frs)
	if err != nil {
		t.Fatal("Marshal FrontierRsp err2")
	}
	if bytes.Compare(data9, data10) != 0 {
		t.Fatal("Marshal FrontierRsp err3")
	}
	b := &protos.BulkPullReqPacket{
		StartHash: types.ZeroHash,
		EndHash:   types.ZeroHash,
	}
	data11, err := marshalMessage(BulkPullRequest, b)
	if err != nil {
		t.Fatal("Marshal BulkPullRequest err1")
	}
	data12, err := protos.BulkPullReqPacketToProto(b)
	if err != nil {
		t.Fatal("Marshal BulkPullRequest err2")
	}
	if bytes.Compare(data11, data12) != 0 {
		t.Fatal("Marshal BulkPullRequest err3")
	}
	blks := make(types.StateBlockList, 0)
	blks = append(blks, blk)
	pullRspMsg := &protos.BulkPullRspPacket{
		Blocks: blks,
	}
	data13, err := marshalMessage(BulkPullRsp, pullRspMsg)
	if err != nil {
		t.Fatal("Marshal BulkPullRsp err1")
	}
	//blks := make(types.StateBlockList, 0)
	//blks = append(blks, blk)
	r := &protos.BulkPullRspPacket{
		Blocks: blks,
	}
	data14, err := protos.BulkPullRspPacketToProto(r)
	if err != nil {
		t.Fatal("Marshal BulkPullRsp err2")
	}
	if bytes.Compare(data13, data14) != 0 {
		t.Fatal("Marshal BulkPullRsp err3")
	}
	data15, err := marshalMessage(BulkPushBlock, blks)
	if err != nil {
		t.Fatal("Marshal BulkPushBlock err1")
	}
	push := &protos.BulkPush{
		Blocks: blks,
	}
	data16, err := protos.BulkPushBlockToProto(push)
	if err != nil {
		t.Fatal("Marshal BulkPushBlock err2")
	}
	if bytes.Compare(data15, data16) != 0 {
		t.Fatal("Marshal BulkPushBlock err3")
	}
	var start1, start2 types.Hash
	err = start1.Of("D2F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47")
	if err != nil {
		t.Fatal("Of StartHash1 error")
	}
	err = start2.Of("12F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47")
	if err != nil {
		t.Fatal("Of StartHash2 error")
	}
	var Locators []*types.Hash
	Locators = append(Locators, &start1)
	Locators = append(Locators, &start2)
	povReq := &protos.PovBulkPullReq{
		StartHash:   start1,
		StartHeight: 1000,
		Count:       2,
		PullType:    protos.PovPullTypeBackward,
		Reason:      protos.PovReasonFetch,
		Locators:    Locators,
	}

	_, err = marshalMessage(PovBulkPullReq, povReq)
	if err != nil {
		t.Fatal("Marshal PovBulkPullReq err")
	}
	blk1, _ := mock.GeneratePovBlock(nil, 0)
	blk2, _ := mock.GeneratePovBlock(nil, 0)
	rsp := &protos.PovBulkPullRsp{
		Count:  2,
		Reason: protos.PovReasonFetch,
		Blocks: types.PovBlocks{blk1, blk2},
	}
	_, err = marshalMessage(PovBulkPullRsp, rsp)
	if err != nil {
		t.Fatal("Marshal PovBulkPullRsp err")
	}
	_, err = marshalMessage(MessageResponse, start1)
	if err != nil {
		t.Fatal("Marshal MessageResponse err")
	}
	_, err = marshalMessage(MessageType(100), start1)
	if err == nil {
		t.Fatal("should return unKnown Message Type")
	}
}

func Test_SendMessage(t *testing.T) {
	removeDir := filepath.Join(config.QlcTestDataDir(), "sendMessage")
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "sendMessage", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	b := "/ip4/127.0.0.1/tcp/19740/ipfs/" + cfg.P2P.ID.PeerID
	cfg.P2P.BootNodes = []string{}
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19740"
	cfg.P2P.Discovery.MDNSEnabled = false
	cfg.LogLevel = "error"

	//start bootNode
	node, err := NewQlcService(dir)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "sendMessage", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.Discovery.DiscoveryInterval = 1
	cfg1.LogLevel = "error"

	//start1 node
	node1, err := NewQlcService(dir1)
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "sendMessage", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19742"
	cfg2.P2P.BootNodes = []string{b}
	cfg2.P2P.Discovery.MDNSEnabled = false
	cfg2.P2P.Discovery.DiscoveryInterval = 1
	cfg2.LogLevel = "error"

	//start node2
	node2, err := NewQlcService(dir2)
	err = node2.Start()
	if err != nil {
		t.Fatal(err)
	}

	//remove test file
	defer func() {
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
			t.Fatal("connect peer timeout")
			return
		default:
			time.Sleep(1 * time.Millisecond)
		}
		count := node1.node.streamManager.PeerCounts()
		if count < 1 {
			continue
		}
		break
	}

	//test send message to peers
	blk := mock.StateBlockWithoutWork()
	//test send message to peers
	peerID := cfg2.P2P.ID.PeerID
	err = node1.SendMessageToPeer(PublishReq, blk, peerID)
	if err != nil {
		t.Fatal(err)
	}
	node1.Broadcast(PublishReq, blk)
	var start1, start2 types.Hash
	err = start1.Of("D2F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47")
	if err != nil {
		t.Fatal("Of StartHash1 error")
	}
	err = start2.Of("12F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47")
	if err != nil {
		t.Fatal("Of StartHash2 error")
	}
	var Locators []*types.Hash
	Locators = append(Locators, &start1)
	Locators = append(Locators, &start2)
	povReq := &protos.PovBulkPullReq{
		StartHash:   start1,
		StartHeight: 1000,
		Count:       2,
		PullType:    protos.PovPullTypeBackward,
		Reason:      protos.PovReasonFetch,
		Locators:    Locators,
	}
	node1.Broadcast(PovBulkPullReq, povReq)
	blk1, _ := mock.GeneratePovBlock(nil, 0)
	blk2, _ := mock.GeneratePovBlock(nil, 0)
	rsp := &protos.PovBulkPullRsp{
		Count:  2,
		Reason: protos.PovReasonFetch,
		Blocks: types.PovBlocks{blk1, blk2},
	}
	node1.Broadcast(PovBulkPullRsp, rsp)
	node1.Broadcast(MessageResponse, start1)
	address := types.Address{}
	Req := protos.NewFrontierReq(address, math.MaxUint32, math.MaxUint32)
	node1.Broadcast(FrontierRequest, Req)
	frontier := &types.Frontier{
		HeaderBlock: start1,
		OpenBlock:   start2,
	}
	frs := &types.FrontierBlock{
		HeaderBlk: blk,
		Fr:        frontier,
	}
	var f []*types.FrontierBlock
	f = append(f, frs)
	fs := &protos.FrontierResponse{
		Fs: f,
	}
	node1.Broadcast(FrontierRsp, fs)
	bp := protos.NewBulkPullReqPacket(start1, start2)
	bp.Hashes = append(bp.Hashes, &start1)
	bp.Hashes = append(bp.Hashes, &start2)
	node1.Broadcast(BulkPullRequest, bp)

	blk = mock.StateBlockWithoutWork()
	br := &protos.BulkPullRspPacket{}
	br.Blocks = append(br.Blocks, blk)
	br.PullType = protos.PullTypeSegment
	node1.Broadcast(BulkPullRsp, br)
	var blks types.StateBlockList
	blks = append(blks, blk)
	node1.Broadcast(BulkPushBlock, blks)
	node1.Broadcast(MessageType(100), "")
	_ = node1.node.SendMessageToPeer(FrontierRsp, fs, node2.node.ID.Pretty())
	_ = node1.node.SendMessageToPeer(MessageResponse, start1, node2.node.ID.Pretty())
	time.Sleep(500 * time.Millisecond)
}
