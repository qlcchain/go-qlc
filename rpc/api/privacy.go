package api

import (
	"errors"
	"time"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
)

type PrivacyApi struct {
	cfg    *config.Config
	l      ledger.Store
	logger *zap.SugaredLogger
	eb     event.EventBus
	feb    *event.FeedEventBus
	cc     *chainctx.ChainContext
}

func NewPrivacyApi(cfg *config.Config, l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *PrivacyApi {
	api := &PrivacyApi{
		cfg:    cfg,
		l:      l,
		eb:     eb,
		feb:    cc.FeedEventBus(),
		logger: log.NewLogger("rpc/privacy"),
		cc:     cc,
	}
	return api
}

func privacyDistributeRawPayload(eb event.EventBus, msgReq *topic.EventPrivacySendReqMsg) ([]byte, error) {
	eb.Publish(topic.EventPrivacySendReq, msgReq)

	select {
	case msgRsp := <-msgReq.RspChan:
		if msgRsp.Err != nil {
			return nil, msgRsp.Err
		}
		return msgRsp.EnclaveKey, nil
	case <-time.After(1 * time.Minute):
		return nil, errors.New("processing timeout")
	}
}

func privacyGetRawPayload(eb event.EventBus, msgReq *topic.EventPrivacyRecvReqMsg) ([]byte, error) {
	eb.Publish(topic.EventPrivacySendReq, msgReq)

	select {
	case msgRsp := <-msgReq.RspChan:
		if msgRsp.Err != nil {
			return nil, msgRsp.Err
		}
		return msgRsp.RawPayload, nil
	case <-time.After(1 * time.Minute):
		return nil, errors.New("processing timeout")
	}
}

type PrivacyDistributeParam struct {
	RawPayload     []byte   `json:"rawPayload"`
	PrivateFrom    string   `json:"privateFrom"`
	PrivateFor     []string `json:"privateFor"`
	PrivateGroupID string   `json:"privateGroupID"`
}

func (api *PrivacyApi) DistributeRawPayload(param *PrivacyDistributeParam) ([]byte, error) {
	msgReq := &topic.EventPrivacySendReqMsg{
		RawPayload:     param.RawPayload,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,

		RspChan: make(chan *topic.EventPrivacySendRspMsg, 1),
	}

	return privacyDistributeRawPayload(api.eb, msgReq)
}

type PrivacyPayloadResponse struct {
	RawPayload []byte `json:"rawPayload"`
}

func (api *PrivacyApi) GetRawPayload(enclaveKey []byte) ([]byte, error) {
	msgReq := &topic.EventPrivacyRecvReqMsg{
		EnclaveKey: enclaveKey,

		RspChan: make(chan *topic.EventPrivacyRecvRspMsg, 1),
	}

	return privacyGetRawPayload(api.eb, msgReq)
}
