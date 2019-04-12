package p2p

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/p2p/protos"
	"github.com/qlcchain/go-qlc/test/mock"
)

func Test_MessageService_Stop(t *testing.T) {
	//node config
	dir1 := filepath.Join(config.QlcTestDataDir())
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/0.0.0.0/tcp/19739"
	cfgFile1.P2P.Discovery.MDNSEnabled = false
	cfgFile1.P2P.BootNodes = []string{}

	eventBus := event.New()
	//start node
	node, err := NewQlcService(cfgFile1, eventBus)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	_, ok := node.dispatcher.subscribersMap.Load(MessageType(common.PublishReq))
	if !ok {
		t.Fatal("subscription PublishReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(common.ConfirmReq))
	if !ok {
		t.Fatal("subscription ConfirmReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(common.ConfirmAck))
	if !ok {
		t.Fatal("subscription ConfirmAck messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(common.FrontierRequest))
	if !ok {
		t.Fatal("subscription FrontierRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(common.FrontierRsp))
	if !ok {
		t.Fatal("subscription FrontierRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(common.BulkPullRequest))
	if !ok {
		t.Fatal("subscription BulkPullRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(common.BulkPullRsp))
	if !ok {
		t.Fatal("subscription BulkPullRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(common.BulkPushBlock))
	if !ok {
		t.Fatal("subscription BulkPushBlock messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(MessageType(common.MessageResponse))
	if !ok {
		t.Fatal("subscription MessageResponse messageType error")
	}

	node.msgService.Stop()
	_, ok = node.dispatcher.subscribersMap.Load(common.PublishReq)
	if ok {
		t.Fatal("subscription PublishReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(common.ConfirmReq)
	if ok {
		t.Fatal("subscription ConfirmReq messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(common.ConfirmAck)
	if ok {
		t.Fatal("subscription ConfirmAck messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(common.FrontierRequest)
	if ok {
		t.Fatal("subscription FrontierRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(common.FrontierRsp)
	if ok {
		t.Fatal("subscription FrontierRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(common.BulkPullRequest)
	if ok {
		t.Fatal("subscription BulkPullRequest messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(common.BulkPullRsp)
	if ok {
		t.Fatal("subscription BulkPullRsp messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(common.BulkPushBlock)
	if ok {
		t.Fatal("subscription BulkPushBlock messageType error")
	}
	_, ok = node.dispatcher.subscribersMap.Load(common.MessageResponse)
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
	blk := mock.StateBlock()
	data1, err := marshalMessage(common.PublishReq, blk)
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
	data3, err := marshalMessage(common.ConfirmReq, blk)
	if err != nil {
		t.Fatal("Marshal ConfirmReq err1")
	}
	packet := &protos.ConfirmReqBlock{
		Blk: blk,
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
	va.Blk = blk
	va.Account = a.Address()
	va.Signature = a.Sign(blk.GetHash())
	data5, err := marshalMessage(common.ConfirmAck, &va)
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
	data7, err := marshalMessage(common.FrontierRequest, Req)
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
	frontierRspTest := protos.NewFrontierRsp(zeroFrontier, 0)
	data9, err := marshalMessage(common.FrontierRsp, frontierRspTest)
	if err != nil {
		t.Fatal("Marshal FrontierRsp err1")
	}
	f := protos.NewFrontierRsp(zeroFrontier, 0)
	data10, err := protos.FrontierResponseToProto(f)
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
	data11, err := marshalMessage(common.BulkPullRequest, b)
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
	data13, err := marshalMessage(common.BulkPullRsp, blk)
	if err != nil {
		t.Fatal("Marshal BulkPullRsp err1")
	}
	r := &protos.BulkPullRspPacket{
		Blk: blk,
	}
	data14, err := protos.BulkPullRspPacketToProto(r)
	if err != nil {
		t.Fatal("Marshal BulkPullRsp err2")
	}
	if bytes.Compare(data13, data14) != 0 {
		t.Fatal("Marshal BulkPullRsp err3")
	}
	data15, err := marshalMessage(common.BulkPushBlock, blk)
	if err != nil {
		t.Fatal("Marshal BulkPushBlock err1")
	}
	push := &protos.BulkPush{
		Blk: blk,
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
	dir := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile, _ := config.DefaultConfig(dir)
	cfgFile.P2P.Listen = "/ip4/0.0.0.0/tcp/19740"
	cfgFile.P2P.Discovery.MDNSEnabled = false
	cfgFile.P2P.BootNodes = []string{}
	b := "/ip4/0.0.0.0/tcp/19740/ipfs/" + cfgFile.P2P.ID.PeerID

	eventBus := event.New()
	//start bootNode
	node, err := NewQlcService(cfgFile, eventBus)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/0.0.0.0/tcp/19741"
	cfgFile1.P2P.BootNodes = []string{b}
	cfgFile1.P2P.Discovery.MDNSEnabled = false
	cfgFile1.P2P.Discovery.DiscoveryInterval = 1

	//start1 node
	node1, err := NewQlcService(cfgFile1, eventBus)
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile2, _ := config.DefaultConfig(dir2)
	cfgFile2.P2P.Listen = "/ip4/0.0.0.0/tcp/19742"
	cfgFile2.P2P.BootNodes = []string{b}
	cfgFile2.P2P.Discovery.MDNSEnabled = false
	cfgFile2.P2P.Discovery.DiscoveryInterval = 1

	//start node2
	node2, err := NewQlcService(cfgFile2, eventBus)
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

	var peerID string
	ticker1 := time.NewTicker(60 * time.Second)
	for {

		select {
		case <-ticker1.C:
			t.Fatal("connect peer timeout")
			return
		default:
			time.Sleep(5 * time.Millisecond)
		}
		peerID, err = node1.node.streamManager.RandomPeer()
		if err != nil {
			continue
		}
		break
	}
	node2.msgService.Stop()
	blk := mock.StateBlock()
	//test send message to peers
	node1.SendMessageToPeer(common.PublishReq, blk, peerID)
	time.Sleep(1 * time.Second)
	if len(node2.msgService.publishMessageCh) != 1 {
		t.Fatal("Send Message To Peer error")
	}
	msg := <-node2.msgService.publishMessageCh
	if msg.MessageType() != MessageType(common.PublishReq) {
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
	node1.SendMessageToPeers(common.PublishReq, blk, peerID)
	time.Sleep(1 * time.Second)
	if len(node2.msgService.publishMessageCh) != 0 {
		t.Fatal("Send Message To Peers error")
	}

	//test broadcast message
	node1.Broadcast(common.PublishReq, blk)
	time.Sleep(1 * time.Second)
	if len(node2.msgService.publishMessageCh) != 1 {
		t.Fatal("broadcast error")
	}
	msg = <-node2.msgService.publishMessageCh
	if msg.MessageType() != MessageType(common.PublishReq) {
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
	dir := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile, _ := config.DefaultConfig(dir)
	cfgFile.P2P.Listen = "/ip4/0.0.0.0/tcp/19743"
	cfgFile.P2P.Discovery.MDNSEnabled = false
	cfgFile.P2P.BootNodes = []string{}
	b := "/ip4/0.0.0.0/tcp/19743/ipfs/" + cfgFile.P2P.ID.PeerID

	eventBus := event.New()
	//start bootNode
	node, err := NewQlcService(cfgFile, eventBus)
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node1 config
	dir1 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile1, _ := config.DefaultConfig(dir1)
	cfgFile1.P2P.Listen = "/ip4/0.0.0.0/tcp/19744"
	cfgFile1.P2P.BootNodes = []string{b}
	cfgFile1.P2P.Discovery.MDNSEnabled = false
	cfgFile1.P2P.Discovery.DiscoveryInterval = 1

	//start1 node
	node1, err := NewQlcService(cfgFile1, eventBus)
	err = node1.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node2 config
	dir2 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile2, _ := config.DefaultConfig(dir2)
	cfgFile2.P2P.Listen = "/ip4/0.0.0.0/tcp/19745"
	cfgFile2.P2P.BootNodes = []string{b}
	cfgFile2.P2P.Discovery.MDNSEnabled = false
	cfgFile2.P2P.Discovery.DiscoveryInterval = 1

	//start node2
	node2, err := NewQlcService(cfgFile2, eventBus)
	err = node2.Start()
	if err != nil {
		t.Fatal(err)
	}

	//node3 config
	dir3 := filepath.Join(config.QlcTestDataDir(), "p2p", uuid.New().String())
	cfgFile3, _ := config.DefaultConfig(dir3)
	cfgFile3.P2P.Listen = "/ip4/0.0.0.0/tcp/19746"
	cfgFile3.P2P.BootNodes = []string{b}
	cfgFile3.P2P.Discovery.MDNSEnabled = false
	cfgFile3.P2P.Discovery.DiscoveryInterval = 1

	//start node2
	node3, err := NewQlcService(cfgFile3, eventBus)
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
		}
		counts = node1.node.streamManager.PeerCounts()
		if counts == 2 {
			break
		}
	}

	node2.msgService.Stop()
	node3.msgService.Stop()
	blk := mock.StateBlock()
	//test send message to peers
	node1.Broadcast(common.PublishReq, blk)
	time.Sleep(1 * time.Second)

	msg := <-node2.msgService.publishMessageCh

	//test message cache
	if node1.msgService.cache.Len() != 1 {
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
		t.Fatal("message cache value lens error")
	}

	if c[0].peerID != node2.node.ID.Pretty() && c[1].peerID != node2.node.ID.Pretty() {
		t.Fatal("message cache peer ID error")
	}
	if c[0].resendTimes != 0 || c[1].resendTimes != 0 {
		t.Fatal("message cache resendTimes error")
	}
	time.Sleep(10 * time.Second)
	node1.msgService.checkMessageCache()
	if c[0].resendTimes != 1 || c[1].resendTimes != 1 {
		t.Fatal("message cache resendTimes error")
	}
	for i := 0; i < 20; i++ {
		node1.msgService.checkMessageCache()
	}
	if node1.msgService.cache.Has(msg.Hash()) {
		t.Fatal("resendTimes error")
	}
}
