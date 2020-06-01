package apis

import (
	"context"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type RepAPI struct {
	rep    *api.RepApi
	logger *zap.SugaredLogger
}

func NewRepAPI(cfg *config.Config, ledger ledger.Store) *RepAPI {
	return &RepAPI{
		rep:    api.NewRepApi(cfg, ledger),
		logger: log.NewLogger("grpc_rep"),
	}
}

func (p *RepAPI) GetRewardData(ctx context.Context, param *pb.RepRewardParam) (*pb.Bytes, error) {
	rewardPara, err := toOriginRepRewardParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.rep.GetRewardData(rewardPara)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (p *RepAPI) UnpackRewardData(ctx context.Context, param *pb.Bytes) (*pb.RepRewardParam, error) {
	r, err := p.rep.UnpackRewardData(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toRepRewardParam(r), nil
}

func (p *RepAPI) GetAvailRewardInfo(ctx context.Context, param *pbtypes.Address) (*pb.RepAvailRewardInfo, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := p.rep.GetAvailRewardInfo(addr)
	if err != nil {
		return nil, err
	}
	return toRepAvailRewardInfo(r), nil
}

func (p *RepAPI) GetRewardSendBlock(ctx context.Context, param *pb.RepRewardParam) (*pbtypes.StateBlock, error) {
	rewardPara, err := toOriginRepRewardParam(param)
	if err != nil {
		return nil, err
	}
	r, err := p.rep.GetRewardSendBlock(rewardPara)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *RepAPI) GetRewardRecvBlock(ctx context.Context, param *pbtypes.StateBlock) (*pbtypes.StateBlock, error) {
	blk, err := toOriginStateBlock(param)
	if err != nil {
		return nil, err
	}
	r, err := p.rep.GetRewardRecvBlock(blk)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *RepAPI) GetRewardRecvBlockBySendHash(ctx context.Context, param *pbtypes.Hash) (*pbtypes.StateBlock, error) {
	h, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.rep.GetRewardRecvBlockBySendHash(h)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (p *RepAPI) GetRepStateWithHeight(ctx context.Context, param *pb.RepStateParams) (*pbtypes.PovRepState, error) {
	account, err := types.HexToAddress(param.GetAccount())
	if err != nil {
		return nil, err
	}
	height := param.GetHeight()
	r, err := p.rep.GetRepStateWithHeight(&api.RepStateParams{
		Account: account,
		Height:  height,
	})
	if err != nil {
		return nil, err
	}
	return &pbtypes.PovRepState{
		Account: r.Account.String(),
		Balance: r.Balance.Int64(),
		Vote:    r.Vote.Int64(),
		Network: r.Network.Int64(),
		Storage: r.Storage.Int64(),
		Oracle:  r.Oracle.Int64(),
		Total:   r.Total.Int64(),
		Status:  r.Status,
		Height:  r.Height,
	}, nil
}

func (p *RepAPI) GetRewardHistory(ctx context.Context, param *pbtypes.Address) (*pb.RepHistoryRewardInfo, error) {
	account, err := types.HexToAddress(param.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := p.rep.GetRewardHistory(account)
	if err != nil {
		return nil, err
	}
	return &pb.RepHistoryRewardInfo{
		LastEndHeight:  r.LastEndHeight,
		RewardBlocks:   r.RewardBlocks,
		LastRewardTime: r.LastRewardTime,
		RewardAmount:   r.RewardAmount.Int64(),
	}, nil
}

func toOriginRepRewardParam(param *pb.RepRewardParam) (*api.RepRewardParam, error) {
	return &api.RepRewardParam{}, nil
}

func toRepRewardParam(param *api.RepRewardParam) *pb.RepRewardParam {
	return &pb.RepRewardParam{}
}

func toRepAvailRewardInfo(r *api.RepAvailRewardInfo) *pb.RepAvailRewardInfo {
	return &pb.RepAvailRewardInfo{}
}
