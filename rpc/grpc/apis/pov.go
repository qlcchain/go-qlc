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
	r, err := p.pov.GetPovStatus()
	if err != nil {
		return nil, err
	}
	return toPovStatus(r), nil
}

func (p *PovAPI) GetHeaderByHeight(ctx context.Context, param *pb.UInt64) (*pb.PovApiHeader, error) {
	v := toOriginUInt64(param)
	r, err := p.pov.GetHeaderByHeight(v)
	if err != nil {
		return nil, err
	}
	return toPovApiHeader(r), nil
}

func (p *PovAPI) GetHeaderByHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiHeader, error) {
	h, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetHeaderByHash(h)
	if err != nil {
		return nil, err
	}
	return toPovApiHeader(r), nil
}

func (p *PovAPI) GetLatestHeader(ctx context.Context, param *empty.Empty) (*pb.PovApiHeader, error) {
	r, err := p.pov.GetLatestHeader()
	if err != nil {
		return nil, err
	}
	return toPovApiHeader(r), nil
}

func (p *PovAPI) GetFittestHeader(ctx context.Context, param *pb.UInt64) (*pb.PovApiHeader, error) {
	v := toOriginUInt64(param)
	r, err := p.pov.GetFittestHeader(v)
	if err != nil {
		return nil, err
	}
	return toPovApiHeader(r), nil
}

func (p *PovAPI) BatchGetHeadersByHeight(ctx context.Context, param *pb.HeadersByHeightRequest) (*pb.PovApiBatchHeader, error) {
	height := param.GetHeight()
	count := param.GetCount()
	asc := param.GetAsc()
	r, err := p.pov.BatchGetHeadersByHeight(height, count, asc)
	if err != nil {
		return nil, err
	}
	return toPovApiBatchHeader(r), nil
}

func (p *PovAPI) GetBlockByHeight(ctx context.Context, param *pb.BlockByHeightRequest) (*pb.PovApiBlock, error) {
	height := param.GetHeight()
	txOffset := param.GetTxOffset()
	txLimit := param.GetTxLimit()
	r, err := p.pov.GetBlockByHeight(height, txOffset, txLimit)
	if err != nil {
		return nil, err
	}
	return toPovApiBlock(r), nil
}

func (p *PovAPI) GetBlockByHash(ctx context.Context, param *pb.BlockByHashRequest) (*pb.PovApiBlock, error) {
	blockHash, err := toOriginHashByValue(param.GetBlockHash())
	if err != nil {
		return nil, err
	}
	txOffset := param.GetTxOffset()
	txLimit := param.GetTxLimit()
	r, err := p.pov.GetBlockByHash(blockHash, txOffset, txLimit)
	if err != nil {
		return nil, err
	}
	return toPovApiBlock(r), nil
}

func (p *PovAPI) GetLatestBlock(ctx context.Context, param *pb.LatestBlockRequest) (*pb.PovApiBlock, error) {
	txOffset := param.GetTxOffset()
	txLimit := param.GetTxLimit()
	r, err := p.pov.GetLatestBlock(txOffset, txLimit)
	if err != nil {
		return nil, err
	}
	return toPovApiBlock(r), nil
}

func (p *PovAPI) GetTransaction(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiTxLookup, error) {
	blockHash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetTransaction(blockHash)
	if err != nil {
		return nil, err
	}
	return toPovApiTxLookup(r), nil
}

func (p *PovAPI) GetTransactionByBlockHashAndIndex(ctx context.Context, param *pb.TransactionByBlockHashRequest) (*pb.PovApiTxLookup, error) {
	blockHash, err := toOriginHashByValue(param.GetBlockHash())
	if err != nil {
		return nil, err
	}
	index := param.GetIndex()
	r, err := p.pov.GetTransactionByBlockHashAndIndex(blockHash, index)
	if err != nil {
		return nil, err
	}
	return toPovApiTxLookup(r), nil
}

func (p *PovAPI) GetTransactionByBlockHeightAndIndex(ctx context.Context, param *pb.TransactionByBlockHeightRequest) (*pb.PovApiTxLookup, error) {
	height := param.GetHeight()
	index := param.GetIndex()
	r, err := p.pov.GetTransactionByBlockHeightAndIndex(height, index)
	if err != nil {
		return nil, err
	}
	return toPovApiTxLookup(r), nil
}

func (p *PovAPI) GetAccountState(ctx context.Context, param *pb.AccountStateRequest) (*pb.PovApiState, error) {
	address, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	stateHash, err := toOriginHashByValue(param.GetStateHash())
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetAccountState(address, stateHash)
	if err != nil {
		return nil, err
	}
	return toPovApiState(r), nil
}

func (p *PovAPI) GetLatestAccountState(ctx context.Context, param *pbtypes.Address) (*pb.PovApiState, error) {
	address, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetLatestAccountState(address)
	if err != nil {
		return nil, err
	}
	return toPovApiState(r), nil
}

func (p *PovAPI) GetAccountStateByBlockHash(ctx context.Context, param *pb.AccountStateByHashRequest) (*pb.PovApiState, error) {
	address, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	hash, err := toOriginHashByValue(param.GetBlockHash())
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetAccountStateByBlockHash(address, hash)
	if err != nil {
		return nil, err
	}
	return toPovApiState(r), nil
}

func (p *PovAPI) GetAccountStateByBlockHeight(ctx context.Context, param *pb.AccountStateByHeightRequest) (*pb.PovApiState, error) {
	address, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	height := param.GetHeight()
	r, err := p.pov.GetAccountStateByBlockHeight(address, height)
	if err != nil {
		return nil, err
	}
	return toPovApiState(r), nil
}

func (p *PovAPI) GetAllRepStatesByStateHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiRepState, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetAllRepStatesByStateHash(hash)
	if err != nil {
		return nil, err
	}
	return toPovApiRepState(r), nil
}

func (p *PovAPI) GetAllRepStatesByBlockHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiRepState, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetAllRepStatesByBlockHash(hash)
	if err != nil {
		return nil, err
	}
	return toPovApiRepState(r), nil
}

func (p *PovAPI) GetAllRepStatesByBlockHeight(ctx context.Context, param *pb.UInt64) (*pb.PovApiRepState, error) {
	v := toOriginUInt64(param)
	r, err := p.pov.GetAllRepStatesByBlockHeight(v)
	if err != nil {
		return nil, err
	}
	return toPovApiRepState(r), nil

}

func (p *PovAPI) GetLedgerStats(ctx context.Context, param *empty.Empty) (*pb.PovLedgerStats, error) {
	r, err := p.pov.GetLedgerStats()
	if err != nil {
		return nil, err
	}
	return toPovLedgerStats(r), nil
}

func (p *PovAPI) GetBlockTDByHash(ctx context.Context, param *pbtypes.Hash) (*pb.PovApiTD, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetBlockTDByHash(hash)
	if err != nil {
		return nil, err
	}
	return toPovApiTD(r), nil
}

func (p *PovAPI) GetBlockTDByHeight(ctx context.Context, param *pb.UInt64) (*pb.PovApiTD, error) {
	v := toOriginUInt64(param)
	r, err := p.pov.GetBlockTDByHeight(v)
	if err != nil {
		return nil, err
	}
	return toPovApiTD(r), nil
}

func (p *PovAPI) GetMinerStats(ctx context.Context, param *pbtypes.Addresses) (*pb.PovMinerStats, error) {
	address, err := toOriginAddresses(param)
	if err != nil {
		return nil, err
	}
	r, err := p.pov.GetMinerStats(address)
	if err != nil {
		return nil, err
	}
	return toPovMinerStats(r), nil
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

func toPovStatus(s *api.PovStatus) *pb.PovStatus {
	return &pb.PovStatus{}
}

func toPovApiHeader(s *api.PovApiHeader) *pb.PovApiHeader {
	return &pb.PovApiHeader{}
}

func toPovApiBatchHeader(s *api.PovApiBatchHeader) *pb.PovApiBatchHeader {
	return &pb.PovApiBatchHeader{}
}

func toPovApiBlock(s *api.PovApiBlock) *pb.PovApiBlock {
	return &pb.PovApiBlock{}
}

func toPovApiTxLookup(s *api.PovApiTxLookup) *pb.PovApiTxLookup {
	return &pb.PovApiTxLookup{}
}

func toPovApiState(s *api.PovApiState) *pb.PovApiState {
	return &pb.PovApiState{}
}

func toPovApiRepState(s *api.PovApiRepState) *pb.PovApiRepState {
	return &pb.PovApiRepState{}
}

func toPovLedgerStats(s *api.PovLedgerStats) *pb.PovLedgerStats {
	return &pb.PovLedgerStats{}
}

func toPovApiTD(s *api.PovApiTD) *pb.PovApiTD {
	return &pb.PovApiTD{}
}

func toPovMinerStats(s *api.PovMinerStats) *pb.PovMinerStats {
	return &pb.PovMinerStats{}
}
