package apis

import (
	"context"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type PrivacyAPI struct {
	privacy *api.PrivacyApi
	logger  *zap.SugaredLogger
}

func NewPrivacyAPI(cfg *config.Config, l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *PrivacyAPI {
	return &PrivacyAPI{
		privacy: api.NewPrivacyApi(cfg, l, eb, cc),
		logger:  log.NewLogger("grpc_privacy"),
	}
}

func (p *PrivacyAPI) DistributeRawPayload(ctx context.Context, param *pb.PrivacyDistributeParam) (*pb.Bytes, error) {
	r, err := p.privacy.DistributeRawPayload(&api.PrivacyDistributeParam{
		RawPayload:     param.GetRawPayload(),
		PrivateFrom:    param.GetPrivateFrom(),
		PrivateFor:     param.GetPrivateFor(),
		PrivateGroupID: param.GetPrivateGroupID(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (p *PrivacyAPI) GetRawPayload(ctx context.Context, param *pb.Bytes) (*pb.Bytes, error) {
	r, err := p.privacy.GetRawPayload(param.GetValue())
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (p *PrivacyAPI) GetBlockPrivatePayload(ctx context.Context, param *pbtypes.Hash) (*pb.Bytes, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.privacy.GetBlockPrivatePayload(hash)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (p *PrivacyAPI) GetDemoKV(ctx context.Context, param *pb.Bytes) (*pb.Bytes, error) {
	r, err := p.privacy.GetDemoKV(param.GetValue())
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}
