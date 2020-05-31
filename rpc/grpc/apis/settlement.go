package apis

import (
	"context"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type SettlementAPI struct {
	settlement *api.SettlementAPI
	logger     *zap.SugaredLogger
}

func NewSettlementAPI(l ledger.Store, cc *chainctx.ChainContext) *SettlementAPI {
	return &SettlementAPI{
		settlement: api.NewSettlement(l, cc),
		logger:     log.NewLogger("grpc_settlement"),
	}
}

func (s *SettlementAPI) ToAddress(ctx context.Context, param *pb.CreateContractParam) (*pbtypes.Address, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetSettlementRewardsBlock(ctx context.Context, param *pbtypes.Hash) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetCreateContractBlock(ctx context.Context, param *pb.CreateContractParam) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetSignContractBlock(ctx context.Context, param *pb.SignContractParam) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetAddPreStopBlock(ctx context.Context, param *pb.StopParam) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetRemovePreStopBlock(ctx context.Context, param *pb.StopParam) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetUpdatePreStopBlock(ctx context.Context, param *pb.UpdateStopParam) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetAddNextStopBlock(ctx context.Context, param *pb.StopParam) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetRemoveNextStopBlock(ctx context.Context, param *pb.StopParam) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetUpdateNextStopBlock(ctx context.Context, param *pb.UpdateStopParam) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetAllContracts(ctx context.Context, param *pb.Offset) (*pb.SettlementContracts, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetContractsByAddress(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.SettlementContracts, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetContractsByStatus(ctx context.Context, param *pb.ContractsByStatusRequest) (*pb.SettlementContracts, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetExpiredContracts(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.SettlementContracts, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetContractsAsPartyA(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.SettlementContracts, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetContractsAsPartyB(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.SettlementContracts, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetContractAddressByPartyANextStop(ctx context.Context, param *pb.ContractAddressByPartyRequest) (*pbtypes.Address, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetContractAddressByPartyBPreStop(ctx context.Context, param *pb.ContractAddressByPartyRequest) (*pbtypes.Address, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetProcessCDRBlock(ctx context.Context, param *pb.ProcessCDRBlockRequest) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetTerminateContractBlock(ctx context.Context, param *pb.TerminateParam) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetCDRStatus(ctx context.Context, param *pb.CDRStatusRequest) (*pb.CDRStatus, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetCDRStatusByCdrData(ctx context.Context, param *pb.CDRStatusByCdrDataRequest) (*pb.CDRStatus, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetCDRStatusByDate(ctx context.Context, param *pb.CDRStatusByDateRequest) (*pb.CDRStatuses, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetAllCDRStatus(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.CDRStatuses, error) {
	panic("implement me")
}

func (s *SettlementAPI) GetMultiPartyCDRStatus(ctx context.Context, param *pb.MultiPartyCDRStatusRequest) (*pb.CDRStatuses, error) {
	panic("implement me")
}
