package apis

import (
	"context"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type MinerAPI struct {
	miner  *api.MinerApi
	logger *zap.SugaredLogger
}

func NewMinerAPI(cfg *config.Config, ledger ledger.Store) *MinerAPI {
	return &MinerAPI{
		miner:  api.NewMinerApi(cfg, ledger),
		logger: log.NewLogger("grpc_miner"),
	}
}

func (m *MinerAPI) GetRewardData(ctx context.Context, params *pb.RewardParam) (*pb.Bytes, error) {
	rewardParam, err := toOriginRewardParam(params)
	if err != nil {
		return nil, err
	}
	r, err := m.miner.GetRewardData(rewardParam)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func toOriginRewardParam(param *pb.RewardParam) (*api.RewardParam, error) {
	return &api.RewardParam{}, nil
}

func toRewardParam(param *api.RewardParam) *pb.RewardParam {
	return &pb.RewardParam{}
}

func (m *MinerAPI) UnpackRewardData(ctx context.Context, params *pb.Bytes) (*pb.RewardParam, error) {
	r, err := m.miner.UnpackRewardData(params.GetValue())
	if err != nil {
		return nil, err
	}
	return toRewardParam(r), nil
}

func (m *MinerAPI) GetAvailRewardInfo(ctx context.Context, params *pbtypes.Address) (*pb.MinerAvailRewardInfo, error) {
	addr, err := toOriginAddress(params)
	if err != nil {
		return nil, err
	}
	r, err := m.miner.GetAvailRewardInfo(addr)
	if err != nil {
		return nil, err
	}
	return &pb.MinerAvailRewardInfo{
		LastEndHeight:     r.LastEndHeight,
		LatestBlockHeight: r.LatestBlockHeight,
		NodeRewardHeight:  r.NodeRewardHeight,
		AvailStartHeight:  r.AvailStartHeight,
		AvailEndHeight:    r.AvailEndHeight,
		AvailRewardBlocks: r.AvailRewardBlocks,
		AvailRewardAmount: toBalanceValue(r.AvailRewardAmount),
		NeedCallReward:    r.NeedCallReward,
	}, nil

}

func (m *MinerAPI) GetRewardSendBlock(ctx context.Context, params *pb.RewardParam) (*pbtypes.StateBlock, error) {
	rewardParam, err := toOriginRewardParam(params)
	if err != nil {
		return nil, err
	}
	r, err := m.miner.GetRewardSendBlock(rewardParam)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (m *MinerAPI) GetRewardRecvBlock(ctx context.Context, params *pbtypes.StateBlock) (*pbtypes.StateBlock, error) {
	blk, err := toOriginStateBlock(params)
	if err != nil {
		return nil, err
	}
	r, err := m.miner.GetRewardRecvBlock(blk)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (m *MinerAPI) GetRewardRecvBlockBySendHash(ctx context.Context, params *pbtypes.Hash) (*pbtypes.StateBlock, error) {
	hash, err := toOriginHash(params)
	if err != nil {
		return nil, err
	}
	r, err := m.miner.GetRewardRecvBlockBySendHash(hash)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (m *MinerAPI) GetRewardHistory(ctx context.Context, params *pbtypes.Address) (*pb.MinerHistoryRewardInfo, error) {
	addr, err := toOriginAddress(params)
	if err != nil {
		return nil, err
	}
	r, err := m.miner.GetRewardHistory(addr)
	if err != nil {
		return nil, err
	}
	return &pb.MinerHistoryRewardInfo{
		LastEndHeight:  r.LastEndHeight,
		RewardBlocks:   r.RewardBlocks,
		RewardAmount:   toBalanceValue(r.RewardAmount),
		LastRewardTime: r.LastRewardTime,
	}, nil
}
