package apis

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
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

type PovAPI struct {
	pov    *api.PovApi
	logger *zap.SugaredLogger
}

func NewPovAPI(ctx context.Context, cfg *config.Config, l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *PovAPI {
	return &PovAPI{
		pov:    api.NewPovApi(ctx, cfg, l, eb, cc),
		logger: log.NewLogger("grpc_pov"),
	}
}

func (p *PovAPI) GetPovStatus(ctx context.Context, param *empty.Empty) (*pb.PovStatus, error) {
	panic("implement me")
}

func (p *PovAPI) GetHeaderByHeight(ctx context.Context, param *pb.UInt64) (*pb.PovApiHeader, error) {
	panic("implement me")
}

func (p *PovAPI) GetHeaderByHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiHeader, error) {
	panic("implement me")
}

func (p *PovAPI) GetLatestHeader(ctx context.Context, param *empty.Empty) (*pb.PovApiHeader, error) {
	panic("implement me")
}

func (p *PovAPI) GetFittestHeader(ctx context.Context, param *pb.UInt64) (*pb.PovApiHeader, error) {
	panic("implement me")
}

func (p *PovAPI) BatchGetHeadersByHeight(ctx context.Context, param *pb.HeadersByHeightRequest) (*pb.PovApiBatchHeader, error) {
	panic("implement me")
}

func (p *PovAPI) GetBlockByHeight(ctx context.Context, param *pb.BlockByHeightRequest) (*pb.PovApiBlock, error) {
	panic("implement me")
}

func (p *PovAPI) GetBlockByHash(ctx context.Context, param *pb.BlockByHashRequest) (*pb.PovApiBlock, error) {
	panic("implement me")
}

func (p *PovAPI) GetLatestBlock(ctx context.Context, param *pb.LatestBlockRequest) (*pb.PovApiBlock, error) {
	panic("implement me")
}

func (p *PovAPI) GetTransaction(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiTxLookup, error) {
	panic("implement me")
}

func (p *PovAPI) GetTransactionByBlockHashAndIndex(ctx context.Context, param *pb.TransactionByBlockHashRequest) (*pb.PovApiTxLookup, error) {
	panic("implement me")
}

func (p *PovAPI) GetTransactionByBlockHeightAndIndex(ctx context.Context, param *pb.TransactionByBlockHeightRequest) (*pb.PovApiTxLookup, error) {
	panic("implement me")
}

func (p *PovAPI) GetAccountState(ctx context.Context, param *pb.AccountStateRequest) (*pb.PovApiState, error) {
	panic("implement me")
}

func (p *PovAPI) GetLatestAccountState(ctx context.Context, param *pbtypes.Address) (*pb.PovApiState, error) {
	panic("implement me")
}

func (p *PovAPI) GetAccountStateByBlockHash(ctx context.Context, param *pb.AccountStateByHashRequest) (*pb.PovApiState, error) {
	panic("implement me")
}

func (p *PovAPI) GetAccountStateByBlockHeight(ctx context.Context, param *pb.AccountStateByHeightRequest) (*pb.PovApiState, error) {
	panic("implement me")
}

func (p *PovAPI) GetAllRepStatesByStateHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiRepState, error) {
	panic("implement me")
}

func (p *PovAPI) GetAllRepStatesByBlockHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiRepState, error) {
	panic("implement me")
}

func (p *PovAPI) GetAllRepStatesByBlockHeight(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiRepState, error) {
	panic("implement me")
}

func (p *PovAPI) GetLedgerStats(ctx context.Context, param *empty.Empty) (*pb.PovLedgerStats, error) {
	panic("implement me")
}

func (p *PovAPI) GetBlockTDByHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiTD, error) {
	panic("implement me")
}

func (p *PovAPI) GetBlockTDByHeight(ctx context.Context, param *pb.UInt64) (*pb.PovApiTD, error) {
	panic("implement me")
}

func (p *PovAPI) GetMinerStats(ctx context.Context, param *pbtypes.Addresses) (*pb.PovMinerStats, error) {
	panic("implement me")
}

func (p *PovAPI) GetRepStats(ctx context.Context, param *pbtypes.Addresses) (*pb.PovRepStats, error) {
	panic("implement me")
}

func (p *PovAPI) GetMinerDayStat(ctx context.Context, param *pb.UInt32) (*pbtypes.PovMinerDayStat, error) {
	panic("implement me")
}

func (p *PovAPI) GetMinerDayStatByHeight(ctx context.Context, param *pb.UInt64) (*pbtypes.PovMinerDayStat, error) {
	panic("implement me")
}

func (p *PovAPI) GetDiffDayStat(ctx context.Context, param *pb.UInt32) (*pbtypes.PovDiffDayStat, error) {
	panic("implement me")
}

func (p *PovAPI) GetDiffDayStatByHeight(ctx context.Context, param *pb.UInt64) (*pbtypes.PovDiffDayStat, error) {
	panic("implement me")
}

func (p *PovAPI) GetHashInfo(ctx context.Context, param *pb.HashInfoRequest) (*pb.PovApiHashInfo, error) {
	panic("implement me")
}

func (p *PovAPI) StartMining(ctx context.Context, param *pb.StartMiningRequest) (*empty.Empty, error) {
	panic("implement me")
}

func (p *PovAPI) StopMining(ctx context.Context, param *empty.Empty) (*empty.Empty, error) {
	panic("implement me")
}

func (p *PovAPI) GetMiningInfo(ctx context.Context, param *empty.Empty) (*pb.PovApiGetMiningInfo, error) {
	panic("implement me")
}

func (p *PovAPI) GetWork(ctx context.Context, param *pb.WorkRequest) (*pb.PovApiGetWork, error) {
	panic("implement me")
}

func (p *PovAPI) SubmitWork(ctx context.Context, param *pb.PovApiSubmitWork) (*empty.Empty, error) {
	panic("implement me")
}

func (p *PovAPI) GetLastNHourInfo(ctx context.Context, param *pb.LastNHourInfoRequest) (*pb.PovApiGetLastNHourInfo, error) {
	panic("implement me")
}

func (p *PovAPI) GetAllOnlineRepStates(ctx context.Context, param *pbtypes.PovHeader) (*pbtypes.PovRepStates, error) {
	panic("implement me")
}

func (p *PovAPI) GetRepStatesByHeightAndAccount(ctx context.Context, param *pb.RepStatesByHeightRequset) (*pbtypes.PovRepState, error) {
	panic("implement me")
}

func (p *PovAPI) CheckAllAccountStates(ctx context.Context, param *empty.Empty) (*pb.PovApiCheckStateRsp, error) {
	panic("implement me")
}
