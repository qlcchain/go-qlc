package p2p

import (
	"bytes"
	"github.com/qlcchain/go-qlc/chain/context"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

func Test_MessageService_Stop(t *testing.T) {
	//node config
	dir1 := filepath.Join(config.QlcTestDataDir(), config.QlcConfigFile)
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

	_, ok := node.dispatcher.subscribersMap.Load(MessageType(PublishReq))
	if !ok {
		t.Fatal("subscription PublishReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(ConfirmReq))
	if !ok {
		t.Fatal("subscription ConfirmReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(ConfirmAck))
	if !ok {
		t.Fatal("subscription ConfirmAck messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(FrontierRequest))
	if !ok {
		t.Fatal("subscription FrontierRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(FrontierRsp))
	if !ok {
		t.Fatal("subscription FrontierRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(BulkPullRequest))
	if !ok {
		t.Fatal("subscription BulkPullRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(BulkPullRsp))
	if !ok {
		t.Fatal("subscription BulkPullRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(BulkPushBlock))
	if !ok {
		t.Fatal("subscription BulkPushBlock messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(MessageResponse))
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
	data13, err := marshalMessage(BulkPullRsp, blks)
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
}

func Test_SendMessage(t *testing.T) {
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	b := "/ip4/127.0.0.1/tcp/19740/ipfs/" + cfg.P2P.ID.PeerID
	cfg.P2P.BootNodes = []string{}
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19740"
	cfg.P2P.Discovery.MDNSEnabled = false

	//start bootNode
	node, err := NewQlcService(dir)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19741"
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//start1 node
	node1, err := NewQlcService(dir1)
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19742"
	cfg2.P2P.BootNodes = []string{b}
	cfg2.P2P.Discovery.MDNSEnabled = false
	cfg2.P2P.Discovery.DiscoveryInterval = 1

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
	node2.msgService.Stop()
	blk := mock.StateBlockWithoutWork()
	//test send message to peers
	peerID := cfg2.P2P.ID.PeerID
	err = node1.SendMessageToPeer(PublishReq, blk, peerID)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	if len(node2.msgService.publishMessageCh) != 1 {
		return
	}
	msg := <-node2.msgService.publishMessageCh
	if msg.MessageType() != MessageType(PublishReq) {
		t.Fatal("receive message type error")
	}
	if msg.MessageFrom() != node1.node.ID.Pretty() {
		t.Fatal("message from error")
	}
	s, err := protos.PublishBlockFromProto(msg.Data())
	if err != nil {
		t.Fatal(err)
	}
	if blk.GetHash().String() != s.Blk.GetHash().String() {
		t.Fatal("receive data error")
	}

	//test send message to peers
	node1.SendMessageToPeers(PublishReq, blk, peerID)
	time.Sleep(500 * time.Millisecond)
	if len(node2.msgService.publishMessageCh) != 0 {
		t.Fatal("Send Message To Peers error")
	}

	//test broadcast message
	node1.Broadcast(PublishReq, blk)
	time.Sleep(500 * time.Millisecond)
	if len(node2.msgService.publishMessageCh) != 1 {
		return
	}
	msg = <-node2.msgService.publishMessageCh
	if msg.MessageType() != MessageType(PublishReq) {
		t.Fatal("receive message type error")
	}
	if msg.MessageFrom() != node1.node.ID.Pretty() {
		t.Fatal("message from error")
	}
	s, err = protos.PublishBlockFromProto(msg.Data())
	if err != nil {
		t.Fatal(err)
	}
	if blk.GetHash().String() != s.Blk.GetHash().String() {
		t.Fatal("receive broadcast data error")
	}

}

func Test_MessageCache(t *testing.T) {
	//bootNode config
	dir := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc := context.NewChainContext(dir)
	cfg, _ := cc.Config()
	cfg.P2P.Listen = "/ip4/127.0.0.1/tcp/19743"
	cfg.P2P.BootNodes = []string{}
	b := "/ip4/0.0.0.0/tcp/19743/ipfs/" + cfg.P2P.ID.PeerID

	//start bootNode
	node, err := NewQlcService(dir)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc1 := context.NewChainContext(dir1)
	cfg1, _ := cc1.Config()
	cfg1.P2P.Listen = "/ip4/127.0.0.1/tcp/19744"
	cfg1.P2P.BootNodes = []string{b}
	cfg1.P2P.Discovery.MDNSEnabled = false
	cfg1.P2P.Discovery.DiscoveryInterval = 1

	//start1 node
	node1, err := NewQlcService(dir1)
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc2 := context.NewChainContext(dir2)
	cfg2, _ := cc2.Config()
	cfg2.P2P.Listen = "/ip4/127.0.0.1/tcp/19745"
	cfg2.P2P.BootNodes = []string{b}
	cfg2.P2P.Discovery.MDNSEnabled = false
	cfg2.P2P.Discovery.DiscoveryInterval = 1

	//start node2
	node2, err := NewQlcService(dir2)
	err = node2.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node3 config
	dir3 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String(), config.QlcConfigFile)
	cc3 := context.NewChainContext(dir3)
	cfg3, _ := cc3.Config()
	cfg3.P2P.Listen = "/ip4/127.0.0.1/tcp/19746"
	cfg3.P2P.BootNodes = []string{b}
	cfg3.P2P.Discovery.MDNSEnabled = false
	cfg3.P2P.Discovery.DiscoveryInterval = 1

	//start node2
	node3, err := NewQlcService(dir3)
	err = node3.Start()
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
		err = node3.msgService.ledger.Close()
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
		err = node3.Stop()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(config.QlcTestDataDir())
		if err != nil {
			t.Fatal(err)
		}
	}()

	ticker1 := time.NewTicker(60 * time.Second)
	var counts int
	for {

		select {
		case <-ticker1.C:
			t.Fatal("connect peer timeout")
			return
		default:
			time.Sleep(1 * time.Millisecond)
		}
		counts = node1.node.streamManager.PeerCounts()
		if counts < 2 {
			continue
		}
		break
	}
	node2.msgService.Stop()
	node3.msgService.Stop()
	blk := mock.StateBlockWithoutWork()
	//test send message to peers
	time.Sleep(100 * time.Millisecond)
	node1.Broadcast(PublishReq, blk)
	time.Sleep(500 * time.Millisecond)
	if len(node2.msgService.publishMessageCh) != 1 {
		return
	}
	msg := <-node2.msgService.publishMessageCh

	//test message cache
	if node1.msgService.cache.Len(false) != 1 {
		t.Fatal("message cache error")
	}

	if !node1.msgService.cache.Has(msg.Hash()) {
		t.Fatal("message cache key error")
	}
	v, err := node1.msgService.cache.Get(msg.Hash())
	if err != nil {
		t.Fatal(err)
	}
	c := v.([]*cacheValue)

	if len(c) != 2 {
		t.Fatal("message cache value lens error:", len(c))
	}

	if c[0].peerID != node2.node.ID.Pretty() && c[1].peerID != node2.node.ID.Pretty() {
		t.Fatal("message cache peer ID error")
	}
	if c[0].resendTimes != 0 || c[1].resendTimes != 0 {
		t.Fatal("message cache resendTimes error")
	}
	//time.Sleep(10 * time.Second)
	//node1.msgService.checkMessageCache()
	//if c[0].resendTimes != 1 || c[1].resendTimes != 1 {
	//	t.Fatal("message cache resendTimes error")
	//}
	//for i := 0; i < 20; i++ {
	//	node1.msgService.checkMessageCache()
	//}
	//if node1.msgService.cache.Has(msg.Hash()) {
	//	t.Fatal("resendTimes error")
	//}
}
