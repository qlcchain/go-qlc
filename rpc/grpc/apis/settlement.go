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
	cabi "github.com/qlcchain/go-qlc/vm/contract/abi/settlement"
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

func (s *SettlementAPI) ToAddress(ctx context.Context, param *pbtypes.CreateContractParam) (*pbtypes.Address, error) {
	p, err := toOriginCreateContractParamOfAbi(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.ToAddress(p)
	if err != nil {
		return nil, err
	}
	return toAddress(r), nil
}

func (s *SettlementAPI) GetSettlementRewardsBlock(ctx context.Context, param *pbtypes.Hash) (*pbtypes.StateBlock, error) {
	p, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetSettlementRewardsBlock(&p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetCreateContractBlock(ctx context.Context, param *pb.CreateContractParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginCreateContractParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetCreateContractBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetSignContractBlock(ctx context.Context, param *pb.SignContractParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginSignContractParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetSignContractBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetAddPreStopBlock(ctx context.Context, param *pb.StopParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginStopParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetAddPreStopBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetRemovePreStopBlock(ctx context.Context, param *pb.StopParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginStopParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetRemovePreStopBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetAddNextStopBlock(ctx context.Context, param *pb.StopParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginStopParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetAddNextStopBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetRemoveNextStopBlock(ctx context.Context, param *pb.StopParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginStopParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetRemoveNextStopBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetUpdatePreStopBlock(ctx context.Context, param *pb.UpdateStopParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginUpdateStopParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetUpdatePreStopBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetUpdateNextStopBlock(ctx context.Context, param *pb.UpdateStopParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginUpdateStopParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetUpdateNextStopBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetAllContracts(ctx context.Context, param *pb.Offset) (*pb.SettlementContracts, error) {
	count, offset := toOffsetByProto(param)
	r, err := s.settlement.GetAllContracts(count, offset)
	if err != nil {
		return nil, err
	}
	return toSettlementContracts(r), nil
}

func (s *SettlementAPI) GetContractsByAddress(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.SettlementContracts, error) {
	addr, count, offset, err := toOriginContractsByAddressRequest(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetContractsByAddress(addr, count, offset)
	if err != nil {
		return nil, err
	}
	return toSettlementContracts(r), nil
}

func (s *SettlementAPI) GetContractsByStatus(ctx context.Context, param *pb.ContractsByStatusRequest) (*pb.SettlementContracts, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	status := param.GetStatus()
	count, offset := toOffsetByValue(param.GetCount(), param.GetOffset())

	r, err := s.settlement.GetContractsByStatus(&addr, status, count, offset)
	if err != nil {
		return nil, err
	}
	return toSettlementContracts(r), nil
}

func (s *SettlementAPI) GetExpiredContracts(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.SettlementContracts, error) {
	addr, count, offset, err := toOriginContractsByAddressRequest(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetExpiredContracts(addr, count, offset)
	if err != nil {
		return nil, err
	}
	return toSettlementContracts(r), nil
}

func (s *SettlementAPI) GetContractsAsPartyA(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.SettlementContracts, error) {
	addr, count, offset, err := toOriginContractsByAddressRequest(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetContractsAsPartyA(addr, count, offset)
	if err != nil {
		return nil, err
	}
	return toSettlementContracts(r), nil
}

func (s *SettlementAPI) GetContractsAsPartyB(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.SettlementContracts, error) {
	addr, count, offset, err := toOriginContractsByAddressRequest(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetContractsAsPartyB(addr, count, offset)
	if err != nil {
		return nil, err
	}
	return toSettlementContracts(r), nil
}

func (s *SettlementAPI) GetContractAddressByPartyANextStop(ctx context.Context, param *pb.ContractAddressByPartyRequest) (*pbtypes.Address, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	stop := param.GetStopName()
	r, err := s.settlement.GetContractAddressByPartyANextStop(&addr, stop)
	if err != nil {
		return nil, err
	}
	return toAddress(*r), nil
}

func (s *SettlementAPI) GetContractAddressByPartyBPreStop(ctx context.Context, param *pb.ContractAddressByPartyRequest) (*pbtypes.Address, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	stop := param.GetStopName()
	r, err := s.settlement.GetContractAddressByPartyBPreStop(&addr, stop)
	if err != nil {
		return nil, err
	}
	return toAddress(*r), nil
}

func (s *SettlementAPI) GetProcessCDRBlock(ctx context.Context, param *pb.ProcessCDRBlockRequest) (*pbtypes.StateBlock, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	p, err := toOriginCDRParams(param.GetParams())
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetProcessCDRBlock(&addr, p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetTerminateContractBlock(ctx context.Context, param *pb.TerminateParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginTerminateParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetTerminateContractBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetCDRStatus(ctx context.Context, param *pb.CDRStatusRequest) (*pb.CDRStatus, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	hash, err := toOriginHashByValue(param.GetHash())
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetCDRStatus(&addr, hash)
	if err != nil {
		return nil, err
	}
	return toCDRStatus(r), nil
}

func (s *SettlementAPI) GetCDRStatusByCdrData(ctx context.Context, param *pb.CDRStatusByCdrDataRequest) (*pb.CDRStatus, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	index := param.GetIndex()
	sender := param.GetSender()
	destination := param.GetDestination()
	r, err := s.settlement.GetCDRStatusByCdrData(&addr, index, sender, destination)
	if err != nil {
		return nil, err
	}
	return toCDRStatus(r), nil
}

func (s *SettlementAPI) GetCDRStatusByDate(ctx context.Context, param *pb.CDRStatusByDateRequest) (*pb.CDRStatuses, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	count, offset := toOffsetByValue(param.GetCount(), param.GetOffset())
	r, err := s.settlement.GetCDRStatusByDate(&addr, start, end, count, offset)
	if err != nil {
		return nil, err
	}
	return toCDRStatuses(r), nil
}

func (s *SettlementAPI) GetAllCDRStatus(ctx context.Context, param *pb.ContractsByAddressRequest) (*pb.CDRStatuses, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	count, offset := toOffsetByValue(param.GetCount(), param.GetOffset())
	r, err := s.settlement.GetAllCDRStatus(&addr, count, offset)
	if err != nil {
		return nil, err
	}
	return toCDRStatuses(r), nil
}

func (s *SettlementAPI) GetMultiPartyCDRStatus(ctx context.Context, param *pb.MultiPartyCDRStatusRequest) (*pb.CDRStatuses, error) {
	addr1, err := toOriginAddressByValue(param.GetFirstAddr())
	if err != nil {
		return nil, err
	}
	addr2, err := toOriginAddressByValue(param.GetSecondAddr())
	if err != nil {
		return nil, err
	}
	count, offset := toOffsetByValue(param.GetCount(), param.GetOffset())
	r, err := s.settlement.GetMultiPartyCDRStatus(&addr1, &addr2, count, offset)
	if err != nil {
		return nil, err
	}
	return toCDRStatuses(r), nil
}

func (s *SettlementAPI) GetSummaryReport(ctx context.Context, param *pb.ReportRequest) (*pbtypes.SummaryResult, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	r, err := s.settlement.GetSummaryReport(&addr, start, end)
	if err != nil {
		return nil, err
	}
	return toSummaryResult(r), nil
}

func (s *SettlementAPI) GetSummaryReportByAccount(ctx context.Context, param *pb.ReportByAccountRequest) (*pbtypes.SummaryResult, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	r, err := s.settlement.GetSummaryReportByAccount(&addr, param.GetAccount(), start, end)
	if err != nil {
		return nil, err
	}
	return toSummaryResult(r), nil
}

func (s *SettlementAPI) GetSummaryReportByCustomer(ctx context.Context, param *pb.ReportByCustomerRequest) (*pbtypes.SummaryResult, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	r, err := s.settlement.GetSummaryReportByCustomer(&addr, param.GetCustomer(), start, end)
	if err != nil {
		return nil, err
	}
	return toSummaryResult(r), nil
}

func (s *SettlementAPI) GenerateInvoices(ctx context.Context, param *pb.InvoicesRequest) (*pb.InvoiceRecords, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	r, err := s.settlement.GenerateInvoices(&addr, start, end)
	if err != nil {
		return nil, err
	}
	return toInvoiceRecords(r), nil
}

func (s *SettlementAPI) GenerateInvoicesByAccount(ctx context.Context, param *pb.InvoicesByAccountRequest) (*pb.InvoiceRecords, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	r, err := s.settlement.GenerateInvoicesByAccount(&addr, param.GetAccount(), start, end)
	if err != nil {
		return nil, err
	}
	return toInvoiceRecords(r), nil
}

func (s *SettlementAPI) GenerateInvoicesByCustomer(ctx context.Context, param *pb.InvoicesByCustomerRequest) (*pb.InvoiceRecords, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	r, err := s.settlement.GenerateInvoicesByCustomer(&addr, param.GetCustomer(), start, end)
	if err != nil {
		return nil, err
	}
	return toInvoiceRecords(r), nil
}

func (s *SettlementAPI) GenerateInvoicesByContract(ctx context.Context, param *pb.InvoicesRequest) (*pb.InvoiceRecords, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	r, err := s.settlement.GenerateInvoicesByContract(&addr, start, end)
	if err != nil {
		return nil, err
	}
	return toInvoiceRecords(r), nil
}

func (s *SettlementAPI) GenerateMultiPartyInvoice(ctx context.Context, param *pb.MultiPartyInvoiceRequest) (*pb.InvoiceRecords, error) {
	first, err := toOriginAddressByValue(param.GetFirstAddr())
	if err != nil {
		return nil, err
	}
	second, err := toOriginAddressByValue(param.GetSecondAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	r, err := s.settlement.GenerateMultiPartyInvoice(&first, &second, start, end)
	if err != nil {
		return nil, err
	}
	return toInvoiceRecords(r), nil
}

func (s *SettlementAPI) GenerateMultiPartySummaryReport(ctx context.Context, param *pb.MultiPartyInvoiceRequest) (*pbtypes.MultiPartySummaryResult, error) {
	first, err := toOriginAddressByValue(param.GetFirstAddr())
	if err != nil {
		return nil, err
	}
	second, err := toOriginAddressByValue(param.GetSecondAddr())
	if err != nil {
		return nil, err
	}
	start := param.GetStart()
	end := param.GetEnd()
	r, err := s.settlement.GenerateMultiPartySummaryReport(&first, &second, start, end)
	if err != nil {
		return nil, err
	}
	return toMultiPartySummaryResult(r), nil
}

func (s *SettlementAPI) GetPreStopNames(ctx context.Context, param *pbtypes.Address) (*pb.Strings, error) {
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}

	r, err := s.settlement.GetPreStopNames(&addr)
	if err != nil {
		return nil, err
	}
	return toStrings(r), nil
}

func (s *SettlementAPI) GetNextStopNames(ctx context.Context, param *pbtypes.Address) (*pb.Strings, error) {
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}

	r, err := s.settlement.GetNextStopNames(&addr)
	if err != nil {
		return nil, err
	}
	return toStrings(r), nil
}

func (s *SettlementAPI) GetRegisterAssetBlock(ctx context.Context, param *pb.RegisterAssetParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginRegisterAssetParam(param)
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetRegisterAssetBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (s *SettlementAPI) GetAllAssets(ctx context.Context, param *pb.Offset) (*pb.AssetParams, error) {
	c, o := toOffsetByProto(param)
	r, err := s.settlement.GetAllAssets(c, o)
	if err != nil {
		return nil, err
	}
	return toAssetParams(r), nil
}

func (s *SettlementAPI) GetAssetsByOwner(ctx context.Context, param *pb.AssetsByOwnerRequest) (*pb.AssetParams, error) {
	owner, err := toOriginAddressByValue(param.GetOwner())
	if err != nil {
		return nil, err
	}
	c, o := toOffsetByValue(param.GetCount(), param.GetOffset())
	r, err := s.settlement.GetAssetsByOwner(&owner, c, o)
	if err != nil {
		return nil, err
	}
	return toAssetParams(r), nil
}

func (s *SettlementAPI) GetAsset(ctx context.Context, param *pbtypes.Address) (*pb.AssetParam, error) {
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := s.settlement.GetAsset(addr)
	if err != nil {
		return nil, err
	}
	return toAssetParam(r), nil
}

func toOriginCreateContractParamOfAbi(param *pbtypes.CreateContractParam) (*cabi.CreateContractParam, error) {
	services := make([]cabi.ContractService, 0)
	for _, c := range param.GetServices() {
		ct := cabi.ContractService{
			ServiceId:   c.GetServiceId(),
			Mcc:         c.GetMcc(),
			Mnc:         c.GetMnc(),
			TotalAmount: c.GetTotalAmount(),
			UnitPrice:   c.GetUnitPrice(),
			Currency:    c.GetCurrency(),
		}
		services = append(services, ct)
	}

	paAddrA, err := toOriginAddressByValue(param.GetPartyA().GetAddress())
	if err != nil {
		return nil, err
	}
	paAddrB, err := toOriginAddressByValue(param.GetPartyB().GetAddress())
	if err != nil {
		return nil, err
	}
	prev, err := toOriginHashByValue(param.GetPrevious())
	if err != nil {
		return nil, err
	}
	return &cabi.CreateContractParam{
		PartyA: cabi.Contractor{
			Address: paAddrA,
			Name:    param.GetPartyA().GetName(),
		},
		PartyB: cabi.Contractor{
			Address: paAddrB,
			Name:    param.GetPartyB().GetName(),
		},
		Previous:  prev,
		Services:  services,
		SignDate:  param.GetSignDate(),
		StartDate: param.GetStartDate(),
		EndDate:   param.GetEndDate(),
	}, nil
}

func toOriginCreateContractParam(param *pb.CreateContractParam) (*api.CreateContractParam, error) {
	services := make([]cabi.ContractService, 0)
	for _, c := range param.GetServices() {
		ct := cabi.ContractService{
			ServiceId:   c.GetServiceId(),
			Mcc:         c.GetMcc(),
			Mnc:         c.GetMnc(),
			TotalAmount: c.GetTotalAmount(),
			UnitPrice:   c.GetUnitPrice(),
			Currency:    c.GetCurrency(),
		}
		services = append(services, ct)
	}

	paAddrA, err := toOriginAddressByValue(param.GetPartyA().GetAddress())
	if err != nil {
		return nil, err
	}
	paAddrB, err := toOriginAddressByValue(param.GetPartyB().GetAddress())
	if err != nil {
		return nil, err
	}

	return &api.CreateContractParam{
		PartyA: cabi.Contractor{
			Address: paAddrA,
			Name:    param.GetPartyA().GetName(),
		},
		PartyB: cabi.Contractor{
			Address: paAddrB,
			Name:    param.GetPartyB().GetName(),
		},
		Services:  services,
		StartDate: param.GetStartDate(),
		EndDate:   param.GetEndDate(),
	}, nil
}

func toOriginSignContractParam(param *pb.SignContractParam) (*api.SignContractParam, error) {
	ca, err := toOriginAddressByValue(param.GetContractAddress())
	if err != nil {
		return nil, err
	}
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	return &api.SignContractParam{
		ContractAddress: ca,
		Address:         addr,
	}, nil
}

func toOriginStopParam(param *pb.StopParam) (*api.StopParam, error) {
	ca, err := toOriginAddressByValue(param.GetStopParam().GetContractAddress())
	if err != nil {
		return nil, err
	}
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	return &api.StopParam{
		StopParam: cabi.StopParam{
			ContractAddress: ca,
			StopName:        param.GetStopParam().GetStopName(),
		},
		Address: addr,
	}, nil
}

func toOriginUpdateStopParam(param *pb.UpdateStopParam) (*api.UpdateStopParam, error) {
	ca, err := toOriginAddressByValue(param.GetUpdateStopParam().GetContractAddress())
	if err != nil {
		return nil, err
	}
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	return &api.UpdateStopParam{
		UpdateStopParam: cabi.UpdateStopParam{
			ContractAddress: ca,
			StopName:        param.GetUpdateStopParam().GetStopName(),
			New:             param.GetUpdateStopParam().GetNewName(),
		},
		Address: addr,
	}, nil
}

func toOriginContractsByAddressRequest(param *pb.ContractsByAddressRequest) (*types.Address, int, *int, error) {
	addr, err := toOriginAddressByValue(param.GetAddr())
	if err != nil {
		return nil, 0, nil, err
	}
	c, o := toOffsetByValue(param.GetCount(), param.GetOffset())
	return &addr, c, o, nil
}

func toSettlementContract(cs *api.SettlementContract) *pb.SettlementContract {
	r := &pb.SettlementContract{
		PartyA: &pbtypes.Contractor{
			Address: toAddressValue(cs.PartyA.Address),
			Name:    cs.PartyA.Name,
		},
		PartyB: &pbtypes.Contractor{
			Address: toAddressValue(cs.PartyB.Address),
			Name:    cs.PartyB.Name,
		},
		Previous:    toHashValue(cs.Previous),
		Services:    nil,
		SignDate:    cs.SignDate,
		StartDate:   cs.StartDate,
		EndDate:     cs.EndDate,
		PreStops:    cs.PreStops,
		NextStops:   cs.NextStops,
		ConfirmDate: cs.ConfirmDate,
		Status:      int32(cs.Status),
		Terminator:  nil,
		Address:     toAddressValue(cs.Address),
	}
	if cs.Terminator != nil {
		r.Terminator = &pbtypes.Terminator{
			Address: toAddressValue(cs.Terminator.Address),
			Request: cs.Terminator.Request,
		}
	}
	if len(cs.Services) > 0 {
		services := make([]*pbtypes.ContractService, 0)
		for _, s := range cs.Services {
			st := &pbtypes.ContractService{
				ServiceId:   s.ServiceId,
				Mcc:         s.Mcc,
				Mnc:         s.Mnc,
				TotalAmount: s.TotalAmount,
				UnitPrice:   s.UnitPrice,
				Currency:    s.Currency,
			}
			services = append(services, st)
		}
		r.Services = services
	}
	return r
}

func toSettlementContracts(cs []*api.SettlementContract) *pb.SettlementContracts {
	r := make([]*pb.SettlementContract, 0)
	for _, c := range cs {
		rt := toSettlementContract(c)
		r = append(r, rt)
	}
	return &pb.SettlementContracts{Contracts: r}
}

func toOriginCDRParams(params []*pbtypes.CDRParam) ([]*cabi.CDRParam, error) {
	rs := make([]*cabi.CDRParam, 0)
	for _, p := range params {
		r := &cabi.CDRParam{
			Index:         p.GetIndex(),
			SmsDt:         p.GetSmsDt(),
			Account:       p.GetAccount(),
			Sender:        p.GetSender(),
			Customer:      p.GetCustomer(),
			Destination:   p.GetDestination(),
			SendingStatus: cabi.SendingStatus(p.GetSendingStatus()),
			DlrStatus:     cabi.DLRStatus(p.GetDlrStatus()),
			PreStop:       p.GetPreStop(),
			NextStop:      p.GetNextStop(),
		}
		rs = append(rs, r)
	}
	return rs, nil
}

func toOriginTerminateParam(param *pb.TerminateParam) (*api.TerminateParam, error) {
	contractAddress, err := toOriginAddressByValue(param.GetTerminateParam().GetContractAddress())
	if err != nil {
		return nil, err
	}
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	return &api.TerminateParam{
		TerminateParam: cabi.TerminateParam{
			ContractAddress: contractAddress,
			Request:         param.GetTerminateParam().GetRequest(),
		},
		Address: addr,
	}, nil
}

func toCDRStatus(cs *api.CDRStatus) *pb.CDRStatus {
	addr := cs.Address
	r := &pb.CDRStatus{
		Address: toAddressValue(*addr),
		Params:  nil,
		Status:  int32(cs.Status),
	}
	if cs.Params != nil {
		pt := make(map[string]*pbtypes.CDRParams)
		for k, v := range cs.Params {
			pts := make([]*pbtypes.CDRParam, 0)
			for _, param := range v {
				pt := &pbtypes.CDRParam{
					Index:         param.Index,
					SmsDt:         param.SmsDt,
					Account:       param.Account,
					Sender:        param.Sender,
					Customer:      param.Customer,
					Destination:   param.Destination,
					SendingStatus: int32(param.SendingStatus),
					DlrStatus:     int32(param.DlrStatus),
					PreStop:       param.PreStop,
					NextStop:      param.NextStop,
				}
				pts = append(pts, pt)
			}
			pt[k] = &pbtypes.CDRParams{
				CdrParams: pts,
			}
		}
		r.Params = pt
	}
	return r
}

func toCDRStatuses(cs []*api.CDRStatus) *pb.CDRStatuses {
	cds := make([]*pb.CDRStatus, 0)
	for _, c := range cs {
		cd := toCDRStatus(c)
		cds = append(cds, cd)
	}
	return &pb.CDRStatuses{Statuses: cds}
}

func toSummaryRecord(s *cabi.SummaryRecord) *pbtypes.SummaryRecord {
	return &pbtypes.SummaryRecord{
		Total:   s.Total,
		Success: s.Success,
		Fail:    s.Fail,
		Result:  s.Result,
	}
}

func toCompareRecord(cr *cabi.CompareRecord) *pbtypes.CompareRecord {
	return &pbtypes.CompareRecord{}
}

func toSummaryResult(r *cabi.SummaryResult) *pbtypes.SummaryResult {
	return &pbtypes.SummaryResult{
		Contract: nil,
		Records:  nil,
		Total:    nil,
	}
}

func toInvoiceRecords(rs []*cabi.InvoiceRecord) *pb.InvoiceRecords {
	records := make([]*pbtypes.InvoiceRecord, 0)
	for _, r := range rs {
		rt := &pbtypes.InvoiceRecord{
			Address:                  toAddressValue(r.Address),
			StartDate:                r.StartDate,
			EndDate:                  r.EndDate,
			Customer:                 r.Customer,
			CustomerSr:               r.CustomerSr,
			Country:                  r.Country,
			Operator:                 r.Operator,
			ServiceId:                r.ServiceId,
			MCC:                      r.MCC,
			MNC:                      r.MNC,
			Currency:                 r.Currency,
			UnitPrice:                r.UnitPrice,
			SumOfBillableSMSCustomer: r.SumOfBillableSMSCustomer,
			SumOfTOTPrice:            r.SumOfTOTPrice,
		}
		records = append(records, rt)
	}
	return &pb.InvoiceRecords{
		Records: records,
	}
}

func toMultiPartySummaryResult(r *cabi.MultiPartySummaryResult) *pbtypes.MultiPartySummaryResult {
	return &pbtypes.MultiPartySummaryResult{
		Contracts: nil,
		Records:   nil,
		Total:     nil,
	}
}

func toRegisterAssetParam(param *api.RegisterAssetParam) *pb.RegisterAssetParam {
	r := &pb.RegisterAssetParam{
		Owner: &pbtypes.Contractor{
			Address: toAddressValue(param.Owner.Address),
			Name:    param.Owner.Name,
		},
		Assets:    nil,
		StartDate: param.StartDate,
		EndDate:   param.EndDate,
		Status:    param.Status,
	}
	if param.Assets != nil {
		ar := make([]*pbtypes.Asset, 0)
		for _, ap := range param.Assets {
			at := &pbtypes.Asset{
				Mcc:         ap.Mcc,
				Mnc:         ap.Mnc,
				TotalAmount: ap.TotalAmount,
				Sla:         nil,
			}
			if ap.SLAs != nil {
				sr := make([]*pbtypes.SLA, 0)
				for _, sp := range ap.SLAs {
					st := &pbtypes.SLA{
						Type:          int32(sp.SLAType),
						Priority:      int32(sp.Priority),
						Value:         sp.Value,
						Compensations: nil,
					}
					if sp.Compensations != nil {
						cr := make([]*pbtypes.Compensation, 0)
						for _, cp := range sp.Compensations {
							ct := &pbtypes.Compensation{
								Low:  cp.Low,
								High: cp.High,
								Rate: cp.Rate,
							}
							cr = append(cr, ct)
						}
						st.Compensations = cr
					}
					sr = append(sr, st)
				}
				at.Sla = sr
			}
			ar = append(ar, at)
		}
		r.Assets = ar
	}
	return r
}

func toOriginRegisterAssetParam(param *pb.RegisterAssetParam) (*api.RegisterAssetParam, error) {
	r := &api.RegisterAssetParam{
		Owner:     cabi.Contractor{},
		Assets:    nil,
		StartDate: param.StartDate,
		EndDate:   param.EndDate,
		Status:    param.Status,
	}
	if param.Owner != nil {
		addr, err := toOriginAddressByValue(param.Owner.Address)
		if err != nil {
			return nil, err
		}
		r.Owner = cabi.Contractor{
			Address: addr,
			Name:    param.Owner.Name,
		}
	}
	if param.Assets != nil {
		ar := make([]*cabi.Asset, 0)
		for _, ap := range param.Assets {
			at := &cabi.Asset{
				Mcc:         ap.GetMcc(),
				Mnc:         ap.GetMnc(),
				TotalAmount: ap.GetTotalAmount(),
				SLAs:        nil,
			}
			if ap.Sla != nil {
				sr := make([]*cabi.SLA, 0)
				for _, sp := range ap.Sla {
					st := &cabi.SLA{
						SLAType:       cabi.SLAType(sp.GetType()),
						Priority:      uint(sp.GetPriority()),
						Value:         sp.GetValue(),
						Compensations: nil,
					}
					if sp.Compensations != nil {
						cr := make([]*cabi.Compensation, 0)
						for _, cp := range sp.Compensations {
							ct := &cabi.Compensation{
								Low:  cp.Low,
								High: cp.High,
								Rate: cp.Rate,
							}
							cr = append(cr, ct)
						}
						st.Compensations = cr
					}
					sr = append(sr, st)
				}
				at.SLAs = sr
			}
			ar = append(ar, at)
		}
		r.Assets = ar
	}
	return r, nil
}

func toAssetParams(params []*api.AssetParam) *pb.AssetParams {
	ar := make([]*pb.AssetParam, 0)
	for _, p := range params {
		ar = append(ar, toAssetParam(p))
	}
	return &pb.AssetParams{Params: ar}
}

func toAssetParam(param *api.AssetParam) *pb.AssetParam {
	r := &pb.AssetParam{
		Owner: &pbtypes.Contractor{
			Address: toAddressValue(param.Owner.Address),
			Name:    param.Owner.Name,
		},
		Assets:    nil,
		SignDate:  param.SignDate,
		StartDate: param.StartDate,
		EndDate:   param.EndDate,
		Status:    int32(param.Status),
		Address:   toAddressValue(param.Address),
	}
	if param.Assets != nil {
		ar := make([]*pb.Asset, 0)
		for _, ap := range param.Assets {
			at := &pb.Asset{
				Mcc:         ap.Mcc,
				Mnc:         ap.Mnc,
				TotalAmount: ap.TotalAmount,
				Sla:         nil,
				AssetID:     toHashValue(ap.AssetID),
			}
			if ap.SLAs != nil {
				sr := make([]*pbtypes.SLA, 0)
				for _, sp := range ap.SLAs {
					st := &pbtypes.SLA{
						Type:          int32(sp.SLAType),
						Priority:      int32(sp.Priority),
						Value:         sp.Value,
						Compensations: nil,
					}
					if sp.Compensations != nil {
						cr := make([]*pbtypes.Compensation, 0)
						for _, cp := range sp.Compensations {
							ct := &pbtypes.Compensation{
								Low:  cp.Low,
								High: cp.High,
								Rate: cp.Rate,
							}
							cr = append(cr, ct)
						}
						st.Compensations = cr
					}
					sr = append(sr, st)
				}
				at.Sla = sr
			}
			ar = append(ar, at)
		}
		r.Assets = ar
	}
	return r
}
