package apis

import (
	"context"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

type RewardsAPI struct {
	reward *api.RewardsAPI
	logger *zap.SugaredLogger
}

func NewRewardsAPI(l ledger.Store, cc *chainctx.ChainContext) *RewardsAPI {
	return &RewardsAPI{
		reward: api.NewRewardsAPI(l, cc),
		logger: log.NewLogger("grpc_reward"),
	}
}

func (r *RewardsAPI) GetUnsignedRewardData(ctx context.Context, param *pb.RewardsParam) (*pbtypes.Hash, error) {
	p, err := toOriginRewardsParam(param)
	if err != nil {
		return nil, err
	}
	result, err := r.reward.GetUnsignedRewardData(p)
	if err != nil {
		return nil, err
	}
	return &pbtypes.Hash{
		Hash: result.String(),
	}, nil
}

func (r *RewardsAPI) GetUnsignedConfidantData(ctx context.Context, param *pb.RewardsParam) (*pbtypes.Hash, error) {
	p, err := toOriginRewardsParam(param)
	if err != nil {
		return nil, err
	}
	result, err := r.reward.GetUnsignedConfidantData(p)
	if err != nil {
		return nil, err
	}
	return &pbtypes.Hash{
		Hash: result.String(),
	}, nil
}

func (r *RewardsAPI) GetSendRewardBlock(ctx context.Context, param *pb.RewardsParamWithSign) (*pbtypes.StateBlock, error) {
	p, sign, err := toOriginRewardsParamWithSign(param)
	if err != nil {
		return nil, err
	}
	result, err := r.reward.GetSendRewardBlock(p, sign)
	if err != nil {
		return nil, err
	}
	return toStateBlock(result), nil
}

func (r *RewardsAPI) GetSendConfidantBlock(ctx context.Context, param *pb.RewardsParamWithSign) (*pbtypes.StateBlock, error) {
	p, sign, err := toOriginRewardsParamWithSign(param)
	if err != nil {
		return nil, err
	}
	result, err := r.reward.GetSendConfidantBlock(p, sign)
	if err != nil {
		return nil, err
	}
	return toStateBlock(result), nil
}

func (r *RewardsAPI) GetReceiveRewardBlock(ctx context.Context, param *pbtypes.Hash) (*pbtypes.StateBlock, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	result, err := r.reward.GetReceiveRewardBlock(&hash)
	if err != nil {
		return nil, err
	}
	return toStateBlock(result), nil
}

func (r *RewardsAPI) IsAirdropRewards(ctx context.Context, param *pb.Bytes) (*pb.Boolean, error) {
	result := r.reward.IsAirdropRewards(param.GetValue())
	return &pb.Boolean{
		Value: result,
	}, nil
}

func (r *RewardsAPI) GetTotalRewards(ctx context.Context, param *pb.String) (*pb.Int64, error) {
	result, err := r.reward.GetTotalRewards(param.GetValue())
	if err != nil {
		return nil, err
	}
	return &pb.Int64{
		Value: result.Int64(),
	}, nil
}

func (r *RewardsAPI) GetRewardsDetail(ctx context.Context, param *pb.String) (*pb.RewardsInfos, error) {
	result, err := r.reward.GetRewardsDetail(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toRewardsInfos(result), nil
}

func (r *RewardsAPI) GetConfidantRewards(ctx context.Context, param *pbtypes.Address) (*pb.ConfidantRewardsResponse, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	result, err := r.reward.GetConfidantRewards(addr)
	if err != nil {
		return nil, err
	}

	rep := &pb.ConfidantRewardsResponse{}
	rep.Rewards = make(map[string]int64)
	for k, v := range result {
		rep.Rewards[k] = v.Int64()
	}
	return rep, nil
}

func (r *RewardsAPI) GetConfidantRewordsDetail(ctx context.Context, param *pbtypes.Address) (*pb.RewardsInfosByAddress, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	result, err := r.reward.GetConfidantRewordsDetail(addr)
	if err != nil {
		return nil, err
	}

	rep := &pb.RewardsInfosByAddress{}
	rep.Infos = make(map[string]*pb.RewardsInfos)
	for k, v := range result {
		rep.Infos[k] = toRewardsInfos(v)
	}
	return rep, nil
}

func toOriginRewardsParam(param *pb.RewardsParam) (*api.RewardsParam, error) {
	return &api.RewardsParam{}, nil
}

func toOriginRewardsParamWithSign(param *pb.RewardsParamWithSign) (*api.RewardsParam, *types.Signature, error) {
	return &api.RewardsParam{}, &types.Signature{}, nil
}

func toRewardsInfos(infos []*abi.RewardsInfo) *pb.RewardsInfos {
	return &pb.RewardsInfos{}
}
