package p2p

import (
	"errors"
	"time"

	"github.com/qlcchain/go-qlc/common"

	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	msgCacheSize           = 65535
	checkCacheTimeInterval = 30 * time.Second
	msgResendMaxTimes      = 10
	msgNeedResendInterval  = 10 * time.Second
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
	messageCh           chan *Message
	publishMessageCh    chan *Message
	confirmReqMessageCh chan *Message
	confirmAckMessageCh chan *Message
	rspMessageCh        chan *Message
	ledger              *ledger.Ledger
	syncService         *ServiceSync
	cache               gcache.Cache
}

// NewService return new Service.
func NewMessageService(netService *QlcService, ledger *ledger.Ledger) *MessageService {
	ms := &MessageService{
		quitCh:              make(chan bool, 6),
		messageCh:           make(chan *Message, 65535),
		publishMessageCh:    make(chan *Message, 65535),
		confirmReqMessageCh: make(chan *Message, 65535),
		confirmAckMessageCh: make(chan *Message, 65535),
		rspMessageCh:        make(chan *Message, 65535),
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
	netService.Register(NewSubscriber(ms, ms.publishMessageCh, false, common.PublishReq))
	netService.Register(NewSubscriber(ms, ms.confirmReqMessageCh, false, common.ConfirmReq))
	netService.Register(NewSubscriber(ms, ms.confirmAckMessageCh, false, common.ConfirmAck))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, common.FrontierRequest))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, common.FrontierRsp))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, common.BulkPullRequest))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, common.BulkPullRsp))
	netService.Register(NewSubscriber(ms, ms.messageCh, false, common.BulkPushBlock))
	netService.Register(NewSubscriber(ms, ms.rspMessageCh, false, common.MessageResponse))
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
			case common.FrontierRequest:
				ms.syncService.onFrontierReq(message)
			case common.FrontierRsp:
				ms.syncService.checkFrontier(message)
				//ms.syncService.onFrontierRsp(message)
			case common.BulkPullRequest:
				ms.syncService.onBulkPullRequest(message)
			case common.BulkPullRsp:
				ms.syncService.onBulkPullRsp(message)
			case common.BulkPushBlock:
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
			case common.MessageResponse:
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
			case common.PublishReq:
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
			case common.ConfirmReq:
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
			case common.ConfirmAck:
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
				//csTemp = append(csTemp, cs[i])
				value.resendTimes++
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

func (ms *MessageService) onMessageResponse(message *Message) {
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

func (ms *MessageService) onPublishReq(message *Message) {
	if ms.netService.node.cfg.PerformanceEnabled {
		blk, err := protos.PublishBlockFromProto(message.Data())
		if err != nil {
			ms.netService.node.logger.Error(err)
			return
		}
		hash := blk.Blk.GetHash()
		ms.addPerformanceTime(hash)
	}
	err := ms.netService.SendMessageToPeer(common.MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Errorf("send Publish Response err:[%s] for message hash:[%s]", err, message.Hash().String())
	}
	hash, err := types.HashBytes(message.Content())
	if err != nil {
		ms.netService.node.logger.Error(err)
		return
	}
	p, err := protos.PublishBlockFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}
	ms.netService.msgEvent.Publish(string(common.EventPublish), p.Blk, hash, message.MessageFrom())
}

func (ms *MessageService) onConfirmReq(message *Message) {
	if ms.netService.node.cfg.PerformanceEnabled {
		blk, err := protos.ConfirmReqBlockFromProto(message.Data())
		if err != nil {
			ms.netService.node.logger.Error(err)
			return
		}
		hash := blk.Blk.GetHash()
		ms.addPerformanceTime(hash)
	}
	err := ms.netService.SendMessageToPeer(common.MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Errorf("send ConfirmReq Response err:[%s] for message hash:[%s]", err, message.Hash().String())
	}
	hash, err := types.HashBytes(message.Content())
	if err != nil {
		ms.netService.node.logger.Error(err)
		return
	}
	r, err := protos.ConfirmReqBlockFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Error(err)
		return
	}
	ms.netService.msgEvent.Publish(string(common.EventConfirmReq), r.Blk, hash, message.MessageFrom())
}

func (ms *MessageService) onConfirmAck(message *Message) {
	if ms.netService.node.cfg.PerformanceEnabled {
		ack, err := protos.ConfirmAckBlockFromProto(message.Data())
		if err != nil {
			ms.netService.node.logger.Error(err)
			return
		}
		ms.addPerformanceTime(ack.Blk.GetHash())
	}
	err := ms.netService.SendMessageToPeer(common.MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Errorf("send ConfirmAck Response err:[%s] for message hash:[%s]", err, message.Hash().String())
	}

	hash, err := types.HashBytes(message.Content())
	if err != nil {
		ms.netService.node.logger.Error(err)
		return
	}
	ack, err := protos.ConfirmAckBlockFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}
	ms.netService.msgEvent.Publish(string(common.EventConfirmAck), ack, hash, message.MessageFrom())
}

func (ms *MessageService) Stop() {
	//ms.netService.node.logger.Info("stopped message monitor")
	// quit.
	for i := 0; i < 6; i++ {
		ms.quitCh <- true
	}
	ms.syncService.quitCh <- true
	ms.netService.Deregister(NewSubscriber(ms, ms.publishMessageCh, false, common.PublishReq))
	ms.netService.Deregister(NewSubscriber(ms, ms.confirmReqMessageCh, false, common.ConfirmReq))
	ms.netService.Deregister(NewSubscriber(ms, ms.confirmAckMessageCh, false, common.ConfirmAck))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, common.FrontierRequest))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, common.FrontierRsp))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, common.BulkPullRequest))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, common.BulkPullRsp))
	ms.netService.Deregister(NewSubscriber(ms, ms.messageCh, false, common.BulkPushBlock))
	ms.netService.Deregister(NewSubscriber(ms, ms.rspMessageCh, false, common.MessageResponse))
}

func marshalMessage(messageName string, value interface{}) ([]byte, error) {
	switch messageName {
	case common.PublishReq:
		packet := protos.PublishBlock{
			Blk: value.(*types.StateBlock),
		}
		data, err := protos.PublishBlockToProto(&packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case common.ConfirmReq:
		packet := &protos.ConfirmReqBlock{
			Blk: value.(*types.StateBlock),
		}
		data, err := protos.ConfirmReqBlockToProto(packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case common.ConfirmAck:
		data, err := protos.ConfirmAckBlockToProto(value.(*protos.ConfirmAckBlock))
		if err != nil {
			return nil, err
		}
		return data, nil
	case common.FrontierRequest:
		data, err := protos.FrontierReqToProto(value.(*protos.FrontierReq))
		if err != nil {
			return nil, err
		}
		return data, nil
	case common.FrontierRsp:
		packet := value.(*protos.FrontierResponse)
		data, err := protos.FrontierResponseToProto(packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case common.BulkPullRequest:
		data, err := protos.BulkPullReqPacketToProto(value.(*protos.BulkPullReqPacket))
		if err != nil {
			return nil, err
		}
		return data, nil
	case common.BulkPullRsp:
		PullRsp := &protos.BulkPullRspPacket{
			Blk: value.(*types.StateBlock),
		}
		data, err := protos.BulkPullRspPacketToProto(PullRsp)
		if err != nil {
			return nil, err
		}
		return data, err
	case common.BulkPushBlock:
		push := &protos.BulkPush{
			Blk: value.(*types.StateBlock),
		}
		data, err := protos.BulkPushBlockToProto(push)
		if err != nil {
			return nil, err
		}
		return data, nil
	case common.MessageResponse:
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
