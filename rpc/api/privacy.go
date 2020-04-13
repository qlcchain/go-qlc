package api

import (
	"errors"
	"time"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type PrivacyApi struct {
	cfg    *config.Config
	l      ledger.Store
	logger *zap.SugaredLogger
	eb     event.EventBus
	feb    *event.FeedEventBus
	cc     *chainctx.ChainContext
	demoKV *contract.PrivacyDemoKVSet
}

func NewPrivacyApi(cfg *config.Config, l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *PrivacyApi {
	api := &PrivacyApi{
		cfg:    cfg,
		l:      l,
		eb:     eb,
		feb:    cc.FeedEventBus(),
		logger: log.NewLogger("rpc/privacy"),
		cc:     cc,
		demoKV: new(contract.PrivacyDemoKVSet),
	}
	return api
}

func privacyDistributeRawPayload(cc *chainctx.ChainContext, msgReq *topic.EventPrivacySendReqMsg) ([]byte, error) {
	cfg, err := cc.Config()
	if err != nil {
		return nil, err
	}
	if !cfg.Privacy.Enable {
		return nil, errors.New("privacy is not supported")
	}

	eb := cc.EventBus()
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

func privacyGetRawPayload(cc *chainctx.ChainContext, msgReq *topic.EventPrivacyRecvReqMsg) ([]byte, error) {
	cfg, err := cc.Config()
	if err != nil {
		return nil, err
	}
	if !cfg.Privacy.Enable {
		return nil, errors.New("privacy is not supported")
	}

	eb := cc.EventBus()
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
	if !api.cfg.Privacy.Enable {
		return nil, errors.New("privacy is not supported")
	}

	msgReq := &topic.EventPrivacySendReqMsg{
		RawPayload:     param.RawPayload,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,

		RspChan: make(chan *topic.EventPrivacySendRspMsg, 1),
	}

	return privacyDistributeRawPayload(api.cc, msgReq)
}

func (api *PrivacyApi) GetRawPayload(enclaveKey []byte) ([]byte, error) {
	if !api.cfg.Privacy.Enable {
		return nil, errors.New("privacy is not supported")
	}

	msgReq := &topic.EventPrivacyRecvReqMsg{
		EnclaveKey: enclaveKey,

		RspChan: make(chan *topic.EventPrivacyRecvRspMsg, 1),
	}

	return privacyGetRawPayload(api.cc, msgReq)
}

func (api *PrivacyApi) GetBlockPrivatePayload(blockHash types.Hash) ([]byte, error) {
	if !api.cfg.Privacy.Enable {
		return nil, errors.New("privacy is not supported")
	}

	pl, err := api.l.GetBlockPrivatePayload(blockHash)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

// GetDemoKV returns KV in PrivacyKV contract (just for demo in testnet)
func (api *PrivacyApi) GetDemoKV(key []byte) ([]byte, error) {
	vmCtx := vmstore.NewVMContext(api.l)
	return abi.PrivacyKVGetValue(vmCtx, key)
}
