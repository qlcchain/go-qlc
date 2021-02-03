package privacy

import (
	"encoding/hex"

	"github.com/AsynkronIT/protoactor-go/actor"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
)

type Controller struct {
	logger     *zap.SugaredLogger
	cfg        *config.Config
	eb         event.EventBus
	subscriber *event.ActorSubscriber
	ptm        *PTM
	quitCh     chan struct{}

	feb            *event.FeedEventBus
	febRpcMsgCh    chan *topic.EventRPCSyncCallMsg
	febRpcMsgSubID event.FeedSubscription
}

func NewController(cc *context.ChainContext) *Controller {
	c := new(Controller)
	c.logger = log.NewLogger("privacy")
	c.cfg, _ = cc.Config()
	c.eb = cc.EventBus()
	c.ptm = NewPTM(c.cfg)
	c.quitCh = make(chan struct{})

	c.feb = cc.FeedEventBus()
	c.febRpcMsgCh = make(chan *topic.EventRPCSyncCallMsg, 100)
	return c
}

func (c *Controller) Init() error {
	err := c.ptm.Init()
	if err != nil {
		return err
	}

	c.febRpcMsgSubID = c.feb.Subscribe(topic.EventRpcSyncCall, c.febRpcMsgCh)
	if c.febRpcMsgSubID == nil {
		c.logger.Error("failed to subscribe EventRpcSyncCall")
	}

	c.subscriber = event.NewActorSubscriber(event.SpawnWithPool(func(ctx actor.Context) {
		switch msg := ctx.Message().(type) {
		case *topic.EventPrivacySendReqMsg:
			c.onEventPrivacySendReqMsg(msg)
		case *topic.EventPrivacyRecvReqMsg:
			c.onEventPrivacyRecvReqMsg(msg)
		}
	}), c.eb)

	err = c.subscriber.Subscribe(topic.EventPrivacySendReq, topic.EventPrivacyRecvReq)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) Start() error {
	err := c.ptm.Start()
	if err != nil {
		return err
	}

	common.Go(c.mainLoop)

	return nil
}

func (c *Controller) Stop() error {
	c.febRpcMsgSubID.Unsubscribe()

	err := c.subscriber.UnsubscribeAll()
	if err != nil {
		return err
	}

	close(c.quitCh)

	err = c.ptm.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) mainLoop() {
	for {
		select {
		case <-c.quitCh:
			return
		case msg := <-c.febRpcMsgCh:
			c.onEventRPCSyncCall(msg)
		}
	}
}

func (c *Controller) onEventRPCSyncCall(msg *topic.EventRPCSyncCallMsg) {
	needRsp := false
	if msg.Name == "Debug.PrivacyInfo" {
		c.getDebugInfo(msg.In, msg.Out)
		needRsp = true
	}
	if needRsp && msg.ResponseChan != nil {
		msg.ResponseChan <- msg.Out
	}
}

func (c *Controller) onEventPrivacySendReqMsg(msg *topic.EventPrivacySendReqMsg) {
	rspMsg := &topic.EventPrivacySendRspMsg{ReqData: msg.ReqData}
	rspMsg.EnclaveKey, rspMsg.Err = c.ptm.Send(msg.RawPayload, msg.PrivateFrom, msg.PrivateFor)

	if rspMsg.Err != nil {
		c.logger.Errorf("PrivacySendReq Event, PrivateFrom:%s, return Err:%s", msg.PrivateFrom, rspMsg.Err)
	} else {
		c.logger.Debugf("PrivacySendReq Event, PrivateFrom:%s, return EnclaveKey:%s", msg.PrivateFrom, formatBytesPrefix(rspMsg.EnclaveKey))
	}

	msg.RspChan <- rspMsg
}

func (c *Controller) onEventPrivacyRecvReqMsg(msg *topic.EventPrivacyRecvReqMsg) {
	rspMsg := &topic.EventPrivacyRecvRspMsg{ReqData: msg.ReqData}
	rspMsg.RawPayload, rspMsg.Err = c.ptm.Receive(msg.EnclaveKey)

	if rspMsg.Err != nil {
		c.logger.Errorf("PrivacyRecvReq Event, EnclaveKey:%s, return Err:%s", formatBytesPrefix(msg.EnclaveKey), rspMsg.Err)
	} else {
		c.logger.Debugf("PrivacyRecvReq Event, EnclaveKey:%s, return RawPayload:%s", formatBytesPrefix(msg.EnclaveKey), formatBytesPrefix(rspMsg.RawPayload))
	}

	msg.RspChan <- rspMsg
}

func (c *Controller) getDebugInfo(in interface{}, out interface{}) {
	outArgs := out.(map[string]interface{})

	outArgs["err"] = nil

	if c.ptm != nil {
		outArgs["ptm"] = c.ptm.GetDebugInfo()
	}
}

func formatBytesPrefix(data []byte) string {
	if len(data) == 0 {
		return "<nil>"
	}
	if len(data) > 32 {
		return hex.EncodeToString(data[0:32])
	}
	return hex.EncodeToString(data[0:])
}
