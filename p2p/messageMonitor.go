package p2p

import (
	"context"
	"errors"
	"github.com/qlcchain/go-qlc/ledger/process"
	"time"

	"github.com/qlcchain/go-qlc/common"

	"github.com/bluele/gcache"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

const (
	checkCacheTimeInterval  = 30 * time.Second
	checkBlockCacheInterval = 60 * time.Second
	msgResendMaxTimes       = 10
	msgNeedResendInterval   = 10 * time.Second
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

	PovStatus      = "20"
	PovPublishReq  = "21"
	PovBulkPullReq = "22"
	PovBulkPullRsp = "23"
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
	ctx                 context.Context
	cancel              context.CancelFunc
	messageCh           chan *Message
	publishMessageCh    chan *Message
	confirmReqMessageCh chan *Message
	confirmAckMessageCh chan *Message
	rspMessageCh        chan *Message
	povMessageCh        chan *Message
	ledger              *ledger.Ledger
	syncService         *ServiceSync
	cache               gcache.Cache
}

// NewService return new Service.
func NewMessageService(netService *QlcService, ledger *ledger.Ledger) *MessageService {
	ctx, cancel := context.WithCancel(context.Background())
	ms := &MessageService{
		ctx:                 ctx,
		cancel:              cancel,
		messageCh:           make(chan *Message, common.P2PMonitorMsgChanSize),
		publishMessageCh:    make(chan *Message, common.P2PMonitorMsgChanSize),
		confirmReqMessageCh: make(chan *Message, common.P2PMonitorMsgChanSize),
		confirmAckMessageCh: make(chan *Message, common.P2PMonitorMsgChanSize),
		rspMessageCh:        make(chan *Message, common.P2PMonitorMsgChanSize),
		povMessageCh:        make(chan *Message, common.P2PMonitorMsgChanSize),
		ledger:              ledger,
		netService:          netService,
		cache:               gcache.New(common.P2PMonitorMsgCacheSize).LRU().Build(),
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
	// PoV message handlers
	netService.Register(NewSubscriber(ms, ms.povMessageCh, false, PovStatus))
	netService.Register(NewSubscriber(ms, ms.povMessageCh, false, PovPublishReq))
	netService.Register(NewSubscriber(ms, ms.povMessageCh, false, PovBulkPullReq))
	netService.Register(NewSubscriber(ms, ms.povMessageCh, false, PovBulkPullRsp))
	// start loop().
	go ms.startLoop()
	go ms.syncService.Start()
	go ms.checkMessageCacheLoop()
	go ms.messageResponseLoop()
	go ms.publishReqLoop()
	go ms.confirmReqLoop()
	go ms.confirmAckLoop()
	go ms.povMessageLoop()
	go ms.processBlockCacheLoop()
}

func (ms *MessageService) processBlockCacheLoop() {
	ms.netService.node.logger.Info("Started process blockCache loop.")
	ticker := time.NewTicker(checkBlockCacheInterval)
	for {
		select {
		case <-ms.ctx.Done():
			return
		case <-ticker.C:
			ms.processBlockCache()
		}
	}
}

func (ms *MessageService) processBlockCache() {
	blocks := make([]*types.StateBlock, 0)
	err := ms.ledger.GetBlockCaches(func(block *types.StateBlock) error {
		blocks = append(blocks, block)
		return nil
	})
	if err != nil {
		ms.netService.node.logger.Error("get block cache error")
	}
	for _, blk := range blocks {
		if b, err := ms.ledger.HasStateBlockConfirmed(blk.GetHash()); b && err == nil {
			_ = ms.ledger.DeleteBlockCache(blk.GetHash())
		} else {
			verifier := process.NewLedgerVerifier(ms.ledger)
			flag, err := verifier.BlockCheck(blk)
			if err != nil {
				ms.netService.node.logger.Error(err)
				return
			}
			if flag == process.Fork {
				err = verifier.Rollback(blk.GetHash())
				if err != nil {
					ms.netService.node.logger.Error(err)
					continue
				}
			} else {
				ms.netService.msgEvent.Publish(common.EventBroadcast, PublishReq, blk)
				ms.netService.msgEvent.Publish(common.EventGenerateBlock, flag, blk)
			}
		}
	}
}

func (ms *MessageService) startLoop() {
	ms.netService.node.logger.Info("Started Message Service.")
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.messageCh:
			switch message.MessageType() {
			case FrontierRequest:
				if err := ms.syncService.onFrontierReq(message); err != nil {
					ms.netService.node.logger.Error(err)
				}
			case FrontierRsp:
				ms.syncService.checkFrontier(message)
			case BulkPullRequest:
				if err := ms.syncService.onBulkPullRequest(message); err != nil {
					ms.netService.node.logger.Error(err)
				}
			case BulkPullRsp:
				if err := ms.syncService.onBulkPullRsp(message); err != nil {
					ms.netService.node.logger.Error(err)
				}
			case BulkPushBlock:
				if err := ms.syncService.onBulkPushBlock(message); err != nil {
					ms.netService.node.logger.Error(err)
				}
			default:
				ms.netService.node.logger.Error("Received unknown message.")
			}
		}
	}
}

func (ms *MessageService) messageResponseLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.rspMessageCh:
			switch message.MessageType() {
			case MessageResponse:
				ms.onMessageResponse(message)
			}
		}
	}
}

func (ms *MessageService) publishReqLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.publishMessageCh:
			switch message.MessageType() {
			case PublishReq:
				ms.onPublishReq(message)
			}
		}
	}
}

func (ms *MessageService) confirmReqLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.confirmReqMessageCh:
			switch message.MessageType() {
			case ConfirmReq:
				ms.onConfirmReq(message)
			}
		}
	}
}

func (ms *MessageService) confirmAckLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.confirmAckMessageCh:
			switch message.MessageType() {
			case ConfirmAck:
				ms.onConfirmAck(message)
			}
		}
	}
}

func (ms *MessageService) checkMessageCacheLoop() {
	ticker := time.NewTicker(checkCacheTimeInterval)
	for {
		select {
		case <-ms.ctx.Done():
			return
		case <-ticker.C:
			ms.checkMessageCache()
		}
	}
}

func (ms *MessageService) checkMessageCache() {
	var cs []*cacheValue
	var csTemp []*cacheValue
	var hash types.Hash
	m := ms.cache.GetALL(false)
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

func (ms *MessageService) povMessageLoop() {
	for {
		select {
		case <-ms.ctx.Done():
			return
		case message := <-ms.povMessageCh:
			switch message.MessageType() {
			case PovStatus:
				ms.onPovStatus(message)
			case PovPublishReq:
				ms.onPovPublishReq(message)
			case PovBulkPullReq:
				ms.onPovBulkPullReq(message)
			case PovBulkPullRsp:
				ms.onPovBulkPullRsp(message)
			default:
				ms.netService.node.logger.Warn("Received unknown pov message.")
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
	err := ms.netService.SendMessageToPeer(MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Debugf("send Publish Response err:[%s] for message hash:[%s]", err, message.Hash().String())
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
	ms.netService.msgEvent.Publish(common.EventPublish, p.Blk, hash, message.MessageFrom())
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
	err := ms.netService.SendMessageToPeer(MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Debugf("send ConfirmReq Response err:[%s] for message hash:[%s]", err, message.Hash().String())
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
	ms.netService.msgEvent.Publish(common.EventConfirmReq, r.Blk, hash, message.MessageFrom())
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
	err := ms.netService.SendMessageToPeer(MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Debugf("send ConfirmAck Response err:[%s] for message hash:[%s]", err, message.Hash().String())
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
	ms.netService.msgEvent.Publish(common.EventConfirmAck, ack, hash, message.MessageFrom())
}

func (ms *MessageService) onPovStatus(message *Message) {
	status, err := protos.PovStatusFromProto(message.data)
	if err != nil {
		ms.netService.node.logger.Errorf("failed to decode PovStatus from peer %s", message.from)
		return
	}

	hash, err := types.HashBytes(message.Content())
	if err != nil {
		ms.netService.node.logger.Error(err)
		return
	}

	ms.netService.msgEvent.Publish(common.EventPovPeerStatus, status, hash, message.MessageFrom())
}

func (ms *MessageService) onPovPublishReq(message *Message) {
	err := ms.netService.SendMessageToPeer(MessageResponse, message.Hash(), message.MessageFrom())
	if err != nil {
		ms.netService.node.logger.Errorf("send PoV Publish Response err:[%s] for message hash:[%s]", err, message.Hash().String())
	}

	hash, err := types.HashBytes(message.Content())
	if err != nil {
		ms.netService.node.logger.Error(err)
		return
	}

	p, err := protos.PovPublishBlockFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}

	ms.netService.msgEvent.Publish(common.EventPovRecvBlock, p.Blk, hash, message.MessageFrom())
}

func (ms *MessageService) onPovBulkPullReq(message *Message) {
	hash, err := types.HashBytes(message.Content())
	if err != nil {
		ms.netService.node.logger.Error(err)
		return
	}

	req, err := protos.PovBulkPullReqFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}

	ms.netService.msgEvent.Publish(common.EventPovBulkPullReq, req, hash, message.MessageFrom())
}

func (ms *MessageService) onPovBulkPullRsp(message *Message) {
	hash, err := types.HashBytes(message.Content())
	if err != nil {
		ms.netService.node.logger.Error(err)
		return
	}

	rsp, err := protos.PovBulkPullRspFromProto(message.Data())
	if err != nil {
		ms.netService.node.logger.Info(err)
		return
	}

	ms.netService.msgEvent.Publish(common.EventPovBulkPullRsp, rsp, hash, message.MessageFrom())
}

func (ms *MessageService) Stop() {
	//ms.netService.node.logger.Info("stopped message monitor")
	// quit.
	ms.cancel()
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
	ms.netService.Deregister(NewSubscriber(ms, ms.povMessageCh, false, PovStatus))
	ms.netService.Deregister(NewSubscriber(ms, ms.povMessageCh, false, PovPublishReq))
	ms.netService.Deregister(NewSubscriber(ms, ms.povMessageCh, false, PovBulkPullReq))
	ms.netService.Deregister(NewSubscriber(ms, ms.povMessageCh, false, PovBulkPullRsp))
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
		packet := value.(*protos.FrontierResponse)
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
	case PovStatus:
		status := value.(*protos.PovStatus)
		data, err := protos.PovStatusToProto(status)
		if err != nil {
			return nil, err
		}
		return data, nil
	case PovPublishReq:
		packet := protos.PovPublishBlock{
			Blk: value.(*types.PovBlock),
		}
		data, err := protos.PovPublishBlockToProto(&packet)
		if err != nil {
			return nil, err
		}
		return data, nil
	case PovBulkPullReq:
		req := value.(*protos.PovBulkPullReq)
		data, err := protos.PovBulkPullReqToProto(req)
		if err != nil {
			return nil, err
		}
		return data, nil
	case PovBulkPullRsp:
		rsp := value.(*protos.PovBulkPullRsp)
		data, err := protos.PovBulkPullRspToProto(rsp)
		if err != nil {
			return nil, err
		}
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
