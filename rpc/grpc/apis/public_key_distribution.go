package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type PublicKeyDistributionAPI struct {
	pkd    *api.PublicKeyDistributionApi
	logger *zap.SugaredLogger
}

func NewPublicKeyDistributionAPI(cfgFile string, l ledger.Store) *PublicKeyDistributionAPI {
	return &PublicKeyDistributionAPI{
		pkd:    api.NewPublicKeyDistributionApi(cfgFile, l),
		logger: log.NewLogger("grpc_pkd"),
	}
}

func (p *PublicKeyDistributionAPI) GetVerifierRegisterBlock(ctx context.Context, param *pb.VerifierRegParam) (*pbtypes.StateBlock, error) {
	verPara, err := toOriginVerifierRegParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifierRegisterBlock(verPara)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetVerifierUnregisterBlock(ctx context.Context, param *pb.VerifierUnRegParam) (*pbtypes.StateBlock, error) {
	verPara, err := toOriginVerifierUnRegParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifierUnregisterBlock(verPara)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetAllVerifiers(ctx context.Context, param *empty.Empty) (*pb.VerifierRegParams, error) {
	r, err := p.pkd.GetAllVerifiers()
	if err != nil {
		return nil, err
	}
	return toVerifierRegParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetVerifiersByType(ctx context.Context, param *pb.String) (*pb.VerifierRegParams, error) {
	r, err := p.pkd.GetVerifiersByType(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toVerifierRegParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetActiveVerifiers(ctx context.Context, param *pb.String) (*pb.VerifierRegParams, error) {
	r, err := p.pkd.GetActiveVerifiers(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toVerifierRegParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetVerifiersByAccount(ctx context.Context, param *pbtypes.Address) (*pb.VerifierRegParams, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifiersByAccount(addr)
	if err != nil {
		return nil, err
	}
	return toVerifierRegParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetVerifierStateByBlockHeight(ctx context.Context, param *pb.VerifierStateByBlockHeightRequest) (*pbtypes.PovVerifierState, error) {
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifierStateByBlockHeight(param.GetHeight(), addr)
	if err != nil {
		return nil, err
	}
	return toPovVerifierState(r), nil
}

func (p *PublicKeyDistributionAPI) GetAllVerifierStatesByBlockHeight(ctx context.Context, param *pb.UInt64) (*pb.PKDVerifierStateList, error) {
	r, err := p.pkd.GetAllVerifierStatesByBlockHeight(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toPKDVerifierStateList(r), nil
}

func (p *PublicKeyDistributionAPI) GetPublishBlock(ctx context.Context, param *pb.PublishParam) (*pb.PublishRet, error) {
	publishParam, err := toOriginPublishParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetPublishBlock(publishParam)
	if err != nil {
		return nil, err
	}
	return toPublishRet(r), nil
}

func (p *PublicKeyDistributionAPI) GetUnPublishBlock(ctx context.Context, param *pb.UnPublishParam) (*pbtypes.StateBlock, error) {
	publishParam, err := toOriginUnPublishParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetUnPublishBlock(publishParam)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetPubKeyByTypeAndID(ctx context.Context, param *pb.TypeAndIDParam) (*pb.PublishInfoStates, error) {
	r, err := p.pkd.GetPubKeyByTypeAndID(param.GetPType(), param.GetPID())
	if err != nil {
		return nil, err
	}
	return toPublishInfoStates(r), nil
}

func (p *PublicKeyDistributionAPI) GetRecommendPubKey(ctx context.Context, param *pb.TypeAndIDParam) (*pb.PublishInfoState, error) {
	r, err := p.pkd.GetRecommendPubKey(param.GetPType(), param.GetPID())
	if err != nil {
		return nil, err
	}
	return toPublishInfoState(r), nil
}

func (p *PublicKeyDistributionAPI) GetPublishInfosByType(ctx context.Context, param *pb.String) (*pb.PublishInfoStates, error) {
	r, err := p.pkd.GetPublishInfosByType(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toPublishInfoStates(r), nil
}

func (p *PublicKeyDistributionAPI) GetPublishInfosByAccountAndType(ctx context.Context, param *pb.AccountAndTypeParam) (*pb.PublishInfoStates, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetPublishInfosByAccountAndType(addr, param.GetPType())
	if err != nil {
		return nil, err
	}
	return toPublishInfoStates(r), nil
}

func (p *PublicKeyDistributionAPI) GetOracleBlock(ctx context.Context, param *pb.OracleParam) (*pbtypes.StateBlock, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetOracleBlock(&api.OracleParam{
		Account: addr,
		OType:   param.GetType(),
		OID:     param.GetId(),
		KeyType: param.GetKeyType(),
		PubKey:  param.GetPubKey(),
		Code:    param.GetCode(),
		Hash:    param.GetHash(),
	})
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetOracleInfosByType(ctx context.Context, param *pb.String) (*pb.OracleParams, error) {
	r, err := p.pkd.GetOracleInfosByType(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toOracleParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetOracleInfosByTypeAndID(ctx context.Context, param *pb.TypeAndIDParam) (*pb.OracleParams, error) {
	r, err := p.pkd.GetOracleInfosByTypeAndID(param.GetPType(), param.GetPID())
	if err != nil {
		return nil, err
	}
	return toOracleParams(r), nil
}

func (p *PublicKeyDistributionAPI) GetOracleInfosByHash(ctx context.Context, param *pbtypes.Hash) (*pb.OracleParams, error) {
	r, err := p.pkd.GetOracleInfosByHash(param.GetHash())
	if err != nil {
		return nil, err
	}
	return toOracleParams(r), nil
}

func (p *PublicKeyDistributionAPI) PackRewardData(ctx context.Context, param *pb.PKDRewardParam) (*pb.Bytes, error) {
	rewardPara, err := toOriginPackRewardData(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.PackRewardData(rewardPara)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{Value: r}, nil
}

func (p *PublicKeyDistributionAPI) UnpackRewardData(ctx context.Context, param *pb.Bytes) (*pb.PKDRewardParam, error) {
	r, err := p.pkd.UnpackRewardData(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toPackRewardData(r), nil
}

func (p *PublicKeyDistributionAPI) GetRewardSendBlock(ctx context.Context, param *pb.PKDRewardParam) (*pbtypes.StateBlock, error) {
	rewardPara, err := toOriginPackRewardData(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetRewardSendBlock(rewardPara)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetRewardRecvBlock(ctx context.Context, param *pbtypes.StateBlock) (*pbtypes.StateBlock, error) {
	blk, err := toOriginStateBlock(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetRewardRecvBlock(blk)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetRewardRecvBlockBySendHash(ctx context.Context, param *pbtypes.Hash) (*pbtypes.StateBlock, error) {
	h, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetRewardRecvBlockBySendHash(h)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *PublicKeyDistributionAPI) GetRewardHistory(ctx context.Context, param *pbtypes.Address) (*pb.PKDHistoryRewardInfo, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetRewardHistory(addr)
	if err != nil {
		return nil, err
	}
	return &pb.PKDHistoryRewardInfo{
		LastEndHeight:  r.LastEndHeight,
		LastBeneficial: toAddressValue(r.LastBeneficial),
		LastRewardTime: r.LastRewardTime,
		RewardAmount:   toBalanceValue(r.RewardAmount),
	}, nil
}

func (p *PublicKeyDistributionAPI) GetAvailRewardInfo(ctx context.Context, param *pbtypes.Address) (*pb.PKDAvailRewardInfo, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetAvailRewardInfo(addr)
	if err != nil {
		return nil, err
	}
	return &pb.PKDAvailRewardInfo{
		LastEndHeight:     r.LastEndHeight,
		LatestBlockHeight: r.LatestBlockHeight,
		NodeRewardHeight:  r.NodeRewardHeight,
		AvailEndHeight:    r.AvailEndHeight,
		AvailRewardAmount: toBalanceValue(r.AvailRewardAmount),
		NeedCallReward:    r.NeedCallReward,
	}, nil
}

func (p *PublicKeyDistributionAPI) GetVerifierHeartBlock(ctx context.Context, param *pb.VerifierHeartBlockRequest) (*pbtypes.StateBlock, error) {
	addr, err := toOriginAddressByValue(param.GetAccount())
	if err != nil {
		return nil, err
	}
	r, err := p.pkd.GetVerifierHeartBlock(addr, param.GetVTypes())
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func toOriginVerifierRegParam(para *pb.VerifierRegParam) (*api.VerifierRegParam, error) {
	return &api.VerifierRegParam{}, nil
}

func toOriginVerifierUnRegParam(para *pb.VerifierUnRegParam) (*api.VerifierUnRegParam, error) {
	return &api.VerifierUnRegParam{}, nil
}

func toVerifierRegParam(para *api.VerifierRegParam) *pb.VerifierRegParam {
	return &pb.VerifierRegParam{}
}

func toVerifierRegParams(para []*api.VerifierRegParam) *pb.VerifierRegParams {
	return &pb.VerifierRegParams{}
}

func toPovVerifierState(state *types.PovVerifierState) *pbtypes.PovVerifierState {
	return &pbtypes.PovVerifierState{}
}

func toPKDVerifierStateList(list *api.PKDVerifierStateList) *pb.PKDVerifierStateList {
	return &pb.PKDVerifierStateList{}
}

func toOriginPublishParam(param *pb.PublishParam) (*api.PublishParam, error) {
	return &api.PublishParam{}, nil
}

func toPublishRet(ret *api.PublishRet) *pb.PublishRet {
	return &pb.PublishRet{}
}

func toOracleParams(params []*api.OracleParam) *pb.OracleParams {
	return &pb.OracleParams{}
}

func toOriginUnPublishParam(param *pb.UnPublishParam) (*api.UnPublishParam, error) {
	return &api.UnPublishParam{}, nil
}

func toPublishInfoStates(states []*api.PublishInfoState) *pb.PublishInfoStates {
	return &pb.PublishInfoStates{}
}

func toPublishInfoState(state *api.PublishInfoState) *pb.PublishInfoState {
	return &pb.PublishInfoState{}
}

func toOriginPackRewardData(param *pb.PKDRewardParam) (*api.PKDRewardParam, error) {
	return &api.PKDRewardParam{}, nil
}

func toPackRewardData(param *api.PKDRewardParam) *pb.PKDRewardParam {
	return &pb.PKDRewardParam{}
}
