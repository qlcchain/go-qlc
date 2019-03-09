package p2p

import (
	"errors"
	"time"

	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

//  Message Type
const (
	PublishReq      = "0" //PublishReq
	ConfirmReq      = "1" //ConfirmReq
	ConfirmAck      = "2" //ConfirmAck
	FrontierRequest = "3" //FrontierReq
	FrontierRsp     = "4" //FrontierRsp
	BulkPullRequest = "5" //BulkPullRequest
	BulkPullRsp     = "6" //BulkPullRsp
	BulkPushBlock   = "7" //BulkPushBlock
	MessageResponse = "8" //MessageResponse
)

const (
	msgCacheSize           = 65535
	checkCacheTimeInterval = 10 * time.Second
	msgResendMaxTimes      = 20
	msgNeedResendInterval  = 1 * time.Second
)

type cacheValue struct {
	peerID      string
	resendTimes uint32
	startTime   time.Time
	data        []byte
	t           string
}

type MessageService struct {
	netService          *QlcService
	quitCh              chan bool
	messageCh           chan Message
	publishMessageCh    chan Message
	confirmReqMessageCh chan Message
	confirmAckMessageCh chan Message
	rspMessageCh        chan Message
	ledger              *ledger.Ledger
	syncService         *ServiceSync
	cache               gcache.Cache
}

// NewService return new Service.
func NewMessageService(netService *QlcService, ledger *ledger.Ledger) *MessageService {
	ms := &MessageService{
		quitCh:              make(chan bool, 1),
		messageCh:           make(chan Message, 65535),
		publishMessageCh:    make(chan Message, 65535),
		confirmReqMessageCh: make(chan Message, 65535),
		confirmAckMessageCh: make(chan Message, 65535),
		rspMessageCh:        make(chan Message, 65535),
		ledger:              ledger,
		netService:          netService,
		cache:               gcache.New(msgCacheSize).LRU().Build(),
	}
	ms.syncService = NewSyncService(netService, ledger)
	return ms
}

// Start start message service.
func (ms *MessageService) Start() {
	// register the network handler.
	netService := ms.netService
	netService.Register(NewSubscriber(ms, ms.publishMessageCh, false, PublishReq))
	netService.Register(NewSubscriber(ms, ms.confirmReqMessageCh, false, ConfirmReq))
	netService.Register(NewSubscriber(ms, ms.confirmAckMessageCh, false, ConfirmAck))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, FrontierRequest))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, FrontierRsp))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, BulkPullRequest))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, BulkPullRsp))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, BulkPushBlock))
	netService.Register(NewSubscriber(ms, ms.rspMessageCh, false, MessageResponse))
	// start loop().
	go ms.startLoop()
	go ms.syncService.Start()
	go ms.checkMessageCacheLoop()
	go ms.messageResponseLoop()
	go ms.publishReqLoop()
	go ms.confirmReqLoop()
	go ms.confirmAckLoop()
}

func (ms *MessageService) startLoop() {
	ms.netService.node.logger.Info("Started Message Service.")
	for {
		select {
		case <-ms.quitCh:
			ms.netService.node.logger.Info("Stopped Message Service.")
			return
		case message := <-ms.messageCh:
			switch message.MessageType() {
			case FrontierRequest:
				ms.syncService.onFrontierReq(message)
			case FrontierRsp:
				ms.syncService.onFrontierRsp(message)
			case BulkPullRequest:
				ms.syncService.onBulkPullRequest(message)
			case BulkPullRsp:
				ms.syncService.onBulkPullRsp(message)
			case BulkPushBlock:
				ms.syncService.onBulkPushBlock(message)
			default:
				ms.netService.node.logger.Error("Received unknown message.")
				time.Sleep(5 * time.Millisecond)
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (ms *MessageService) messageResponseLoop() {
	for {
		select {
		case <-ms.quitCh:
			return
		case message := <-ms.rspMessageCh:
			switch message.MessageType() {
			case MessageResponse:
				ms.onMessageResponse(message)
			default:
				time.Sleep(5 * time.Millisecond)
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (ms *MessageService) publishReqLoop() {
	for {
		select {
		case <-ms.quitCh:
			return
		case message := <-ms.publishMessageCh:
			switch message.MessageType() {
			case PublishReq:
				ms.onPublishReq(message)
			default:
				time.Sleep(5 * time.Millisecond)
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (ms *MessageService) confirmReqLoop() {
	for {
		select {
		case <-ms.quitCh:
			return
		case message := <-ms.confirmReqMessageCh:
			switch message.MessageType() {
			case ConfirmReq:
				ms.onConfirmReq(message)
			default:
				time.Sleep(5 * time.Millisecond)
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (ms *MessageService) confirmAckLoop() {
	for {
		select {
		case <-ms.quitCh:
			return
		case message := <-ms.confirmAckMessageCh:
			switch message.MessageType() {
			case ConfirmAck:
				ms.onConfirmAck(message)
			default:
				time.Sleep(5 * time.Millisecond)
			}
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (ms *MessageService) checkMessageCacheLoop() {
	ticker := time.NewTicker(checkCacheTimeInterval)
	for {
		select {
		case <-ms.quitCh:
			ms.netService.node.logger.Info("Stopped Message Service.")
			return
		case <-ticker.C:
			ms.checkMessageCache()
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (ms *MessageService) checkMessageCache() {
	var cs []*cacheValue
	var csTemp []*cacheValue
	var hash types.Hash
	m := ms.cache.GetALL()
	for k, v := range m {
		hash = k.(types.Hash)
		cs = v.([]*cacheValue)
		for i, value := range cs {
			if value.resendTimes > msgResendMaxTimes {
				csTemp = append(csTemp, cs[i])
				continue
			}
			if time.Now().Sub(value.startTime) < msgNeedResendInterval {
				continue
			}
			stream := ms.netService.node.streamManager.FindByPeerID(value.peerID)
			if stream == nil {
				ms.netService.node.logger.Debug("Failed to locate peer's stream,maybe lost connect")
				csTemp = append(csTemp, cs[i])
				continue
			}
			stream.messageChan <- value.data
			value.resendTimes++
			if value.resendTimes > msgResendMaxTimes {
				csTemp = append(csTemp, cs[i])
				continue
			}
		}

		if len(csTemp) == len(cs) {
			t := ms.cache.Remove(hash)
			if t {
				ms.netService.node.logger.Debugf("remove message:[%s] success", hash.String())
			}
		} else {
			csDiff := sliceDiff(cs, csTemp)
			err := ms.cache.Set(hash, csDiff)
			if err != nil {
				ms.netService.node.logger.Error(err)
			}
		}
	}
}

func (ms *MessageService) onMessageResponse(message Message) {
	//ms.netService.node.logger.Info("receive MessageResponse")
	var hash types.Hash
	var cs []*cacheValue
	var csTemp []*cacheValue
	err := hash.UnmarshalText(message.Data())
	if err != nil {
		ms.netService.node.logger.Errorf("onMessageResponse err:[%s]", err)
		return
	}
	v, err := ms.cache.Get(hash)
	if err != nil {
		if err == gcache.KeyNotFoundError {
			ms.netService.node.logger.Debugf("this hash:[%s] is not in cache", hash)
		} else {
			ms.netService.node.logger.Errorf("Get cache err:[%s] for hash:[%s]", err, hash)
		}
		return
	}
	cs = v.([]*cacheValue)
	for k, v := range cs {
		if v.peerID == message.MessageFrom() {
			csTemp = append(csTemp, cs[k])
		}
	}

	if len(csTemp) == len(cs) {
		t := ms.cache.Remove(hash)
		if t {
			ms.netService.node.logger.Debugf("remove message cache for hash:[%s] success", hash)
		}
	} else {
		csDiff := sliceDiff(cs, csTemp)
		err := ms.cache.Set(hash, csDiff)
		if err != nil {
			ms.netService.node.logger.Error(err)
		}
	}
}

func (ms *MessageService) onPublishReq(message Message) {
	if ms.netService.node.cfg.PerformanceTest.Enabled {
		blk, err := protos.PublishBlockFromProto(message.Data())
		if err != nil {
			ms.netService.node.logger.Error(err)
			return
		}
		hash := blk.Blk.GetHash()
		ms.addPerformanceTime(hash)
	}
	ms.netService.node.logger.Infof("receive PublishReq from:[%s]", message.MessageFrom())
	err := ms.netService.SendMessageToPeer(MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Errorf("send Publish Response err:[%s] for message hash:[%s]", err, message.Hash().String())
	}

	ms.netService.msgEvent.GetEvent("consensus").Notify(EventPublish, message)
}

func (ms *MessageService) onConfirmReq(message Message) {
	if ms.netService.node.cfg.PerformanceTest.Enabled {
		blk, err := protos.ConfirmReqBlockFromProto(message.Data())
		if err != nil {
			ms.netService.node.logger.Error(err)
			return
		}
		hash := blk.Blk.GetHash()
		ms.addPerformanceTime(hash)
	}
	ms.netService.node.logger.Infof("receive ConfirmReq from :[%s]", message.MessageFrom())
	//ms.netService.node.logger.Info("message hash is:", message.Hash())
	err := ms.netService.SendMessageToPeer(MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Errorf("send ConfirmReq Response err:[%s] for message hash:[%s]", err, message.Hash().String())
	}

	ms.netService.msgEvent.GetEvent("consensus").Notify(EventConfirmReq, message)
}

func (ms *MessageService) onConfirmAck(message Message) {
	if ms.netService.node.cfg.PerformanceTest.Enabled {
		ack, err := protos.ConfirmAckBlockFromProto(message.Data())
		if err != nil {
			ms.netService.node.logger.Error(err)
			return
		}
		ms.addPerformanceTime(ack.Blk.GetHash())
	}
	ms.netService.node.logger.Infof("receive ConfirmAck from :[%s]", message.MessageFrom())
	err := ms.netService.SendMessageToPeer(MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Errorf("send ConfirmAck Response err:[%s] for message hash:[%s]", err, message.Hash().String())
	}

	ms.netService.msgEvent.GetEvent("consensus").Notify(EventConfirmAck, message)
}

func (ms *MessageService) Stop() {
	//ms.netService.node.logger.Info("stopped message monitor")
	// quit.
	ms.quitCh <- true
	ms.syncService.quitCh <- true
	ms.netService.Deregister(NewSubscriber(ms, ms.publishMessageCh, false, PublishReq))
	ms.netService.Deregister(NewSubscriber(ms, ms.confirmReqMessageCh, false, ConfirmReq))
	ms.netService.Deregister(NewSubscriber(ms, ms.confirmAckMessageCh, false, ConfirmAck))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, FrontierRequest))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, FrontierRsp))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, BulkPullRequest))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, BulkPullRsp))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, BulkPushBlock))
	ms.netService.Deregister(NewSubscriber(ms, ms.rspMessageCh, false, MessageResponse))
}

func marshalMessage(messageName string, value interface{}) ([]byte, error) {
	switch messageName {
	case PublishReq:
		packet := protos.PublishBlock{
			Blk: value.(*types.StateBlock),
		}
		data, err := protos.PublishBlockToProto(&packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case ConfirmReq:
		packet := &protos.ConfirmReqBlock{
			Blk: value.(*types.StateBlock),
		}
		data, err := protos.ConfirmReqBlockToProto(packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case ConfirmAck:
		data, err := protos.ConfirmAckBlockToProto(value.(*protos.ConfirmAckBlock))
		if err != nil {
			return nil, err
		}
		return data, nil
	case FrontierRequest:
		data, err := protos.FrontierReqToProto(value.(*protos.FrontierReq))
		if err != nil {
			return nil, err
		}
		return data, nil
	case FrontierRsp:
		packet := &protos.FrontierResponse{
			Frontier: value.(*types.Frontier),
		}
		data, err := protos.FrontierResponseToProto(packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case BulkPullRequest:
		data, err := protos.BulkPullReqPacketToProto(value.(*protos.BulkPullReqPacket))
		if err != nil {
			return nil, err
		}
		return data, nil
	case BulkPullRsp:
		PullRsp := &protos.BulkPullRspPacket{
			Blk: value.(*types.StateBlock),
		}
		data, err := protos.BulkPullRspPacketToProto(PullRsp)
		if err != nil {
			return nil, err
		}
		return data, err
	case BulkPushBlock:
		push := &protos.BulkPush{
			Blk: value.(*types.StateBlock),
		}
		data, err := protos.BulkPushBlockToProto(push)
		if err != nil {
			return nil, err
		}
		return data, nil
	case MessageResponse:
		hash := value.(types.Hash)
		data, _ := hash.MarshalText()
		return data, nil
	default:
		return nil, errors.New("unKnown Message Type")
	}
}

func (ms *MessageService) addPerformanceTime(hash types.Hash) {
	if exit, err := ms.ledger.IsPerformanceTimeExist(hash); !exit && err == nil {
		if b, err := ms.ledger.HasStateBlock(hash); !b && err == nil {
			t := &types.PerformanceTime{
				Hash: hash,
				T0:   time.Now().UnixNano(),
				T1:   0,
				T2:   0,
				T3:   0,
			}
			err = ms.ledger.AddOrUpdatePerformance(t)
			if err != nil {
				ms.netService.node.logger.Error("error when run AddOrUpdatePerformance in onConfirmAck func")
			}
		}
	}
}

// InSliceIface checks given interface in interface slice.
func inSliceIface(v interface{}, sl []*cacheValue) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// SliceDiff returns diff slice of slice1 - slice2.
func sliceDiff(slice1, slice2 []*cacheValue) (diffslice []*cacheValue) {
	for _, v := range slice1 {
		if !inSliceIface(v, slice2) {
			diffslice = append(diffslice, v)
		}
	}
	return
}
