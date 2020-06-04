package apis

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

type NEP5PledgeAPI struct {
	nep5   *api.NEP5PledgeAPI
	logger *zap.SugaredLogger
}

func NewNEP5PledgeAPI(cfgFile string, l ledger.Store) *NEP5PledgeAPI {
	return &NEP5PledgeAPI{
		nep5:   api.NewNEP5PledgeAPI(cfgFile, l),
		logger: log.NewLogger("grpc_nep5"),
	}
}

func (n *NEP5PledgeAPI) GetPledgeData(ctx context.Context, param *pb.PledgeParam) (*pb.Bytes, error) {
	block, err := toOriginPledgeParam(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetPledgeData(block)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (n *NEP5PledgeAPI) GetPledgeBlock(ctx context.Context, param *pb.PledgeParam) (*pbtypes.StateBlock, error) {
	block, err := toOriginPledgeParam(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetPledgeBlock(block)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (n *NEP5PledgeAPI) GetPledgeRewardBlock(ctx context.Context, param *pbtypes.StateBlock) (*pbtypes.StateBlock, error) {
	block, err := toOriginStateBlock(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetPledgeRewardBlock(block)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (n *NEP5PledgeAPI) GetPledgeRewardBlockBySendHash(ctx context.Context, param *pbtypes.Hash) (*pbtypes.StateBlock, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetPledgeRewardBlockBySendHash(hash)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (n *NEP5PledgeAPI) GetWithdrawPledgeData(ctx context.Context, param *pb.WithdrawPledgeParam) (*pb.Bytes, error) {
	p, err := toOriginWithdrawPledgeParam(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetWithdrawPledgeData(p)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (n *NEP5PledgeAPI) GetWithdrawPledgeBlock(ctx context.Context, param *pb.WithdrawPledgeParam) (*pbtypes.StateBlock, error) {
	p, err := toOriginWithdrawPledgeParam(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetWithdrawPledgeBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (n *NEP5PledgeAPI) GetWithdrawRewardBlock(ctx context.Context, param *pbtypes.StateBlock) (*pbtypes.StateBlock, error) {
	block, err := toOriginStateBlock(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetWithdrawRewardBlock(block)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (n *NEP5PledgeAPI) GetWithdrawRewardBlockBySendHash(ctx context.Context, param *pbtypes.Hash) (*pbtypes.StateBlock, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetWithdrawRewardBlockBySendHash(hash)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (n *NEP5PledgeAPI) ParsePledgeInfo(ctx context.Context, param *pb.Bytes) (*pbtypes.NEP5PledgeInfo, error) {
	r, err := n.nep5.ParsePledgeInfo(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toNEP5PledgeInfo2(r), nil
}

func (n *NEP5PledgeAPI) GetPledgeInfosByPledgeAddress(ctx context.Context, param *pbtypes.Address) (*pb.PledgeInfos, error) {
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	r := n.nep5.GetPledgeInfosByPledgeAddress(addr)
	return toPledgeInfos(r), nil
}

func (n *NEP5PledgeAPI) GetPledgeBeneficialTotalAmount(ctx context.Context, param *pbtypes.Address) (*pb.Int64, error) {
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetPledgeBeneficialTotalAmount(addr)
	if err != nil {
		return nil, err
	}
	return &pb.Int64{
		Value: r.Int64(),
	}, nil
}

func (n *NEP5PledgeAPI) GetBeneficialPledgeInfosByAddress(ctx context.Context, param *pbtypes.Address) (*pb.PledgeInfos, error) {
	addr, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	r := n.nep5.GetBeneficialPledgeInfosByAddress(addr)
	return toPledgeInfos(r), nil
}

func (n *NEP5PledgeAPI) GetBeneficialPledgeInfos(ctx context.Context, param *pb.BeneficialPledgeRequest) (*pb.PledgeInfos, error) {
	addr, err := toOriginAddressByValue(param.GetBeneficial())
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetBeneficialPledgeInfos(addr, param.GetPType())
	if err != nil {
		return nil, err
	}
	return toPledgeInfos(r), nil
}

func (n *NEP5PledgeAPI) GetPledgeBeneficialAmount(ctx context.Context, param *pb.BeneficialPledgeRequest) (*pb.Int64, error) {
	addr, err := toOriginAddressByValue(param.GetBeneficial())
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetPledgeBeneficialAmount(addr, param.GetPType())
	if err != nil {
		return nil, err
	}
	return &pb.Int64{
		Value: r.Int64(),
	}, nil
}

func (n *NEP5PledgeAPI) GetPledgeInfo(ctx context.Context, param *pb.WithdrawPledgeParam) (*pb.NEP5PledgeInfos, error) {
	p, err := toOriginWithdrawPledgeParam(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetPledgeInfo(p)
	if err != nil {
		return nil, err
	}
	return toNEP5PledgeInfos(r), nil
}

func (n *NEP5PledgeAPI) GetPledgeInfoWithNEP5TxId(ctx context.Context, param *pb.WithdrawPledgeParam) (*pb.NEP5PledgeInfo, error) {
	p, err := toOriginWithdrawPledgeParam(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetPledgeInfoWithNEP5TxId(p)
	if err != nil {
		return nil, err
	}
	return toNEP5PledgeInfo(r), nil
}

func (n *NEP5PledgeAPI) GetPledgeInfoWithTimeExpired(ctx context.Context, param *pb.WithdrawPledgeParam) (*pb.NEP5PledgeInfos, error) {
	p, err := toOriginWithdrawPledgeParam(param)
	if err != nil {
		return nil, err
	}
	r, err := n.nep5.GetPledgeInfo(p)
	if err != nil {
		return nil, err
	}
	return toNEP5PledgeInfos(r), nil
}

func (n *NEP5PledgeAPI) GetAllPledgeInfo(context.Context, *empty.Empty) (*pb.NEP5PledgeInfos, error) {
	r, err := n.nep5.GetAllPledgeInfo()
	if err != nil {
		return nil, err
	}
	return toNEP5PledgeInfos(r), nil
}

func (n *NEP5PledgeAPI) GetTotalPledgeAmount(context.Context, *empty.Empty) (*pb.Int64, error) {
	r, err := n.nep5.GetTotalPledgeAmount()
	if err != nil {
		return nil, err
	}
	return &pb.Int64{
		Value: r.Int64(),
	}, nil
}

func toPledgeInfos(info *api.PledgeInfos) *pb.PledgeInfos {
	r := &pb.PledgeInfos{
		PledgeInfos:  nil,
		TotalAmounts: info.TotalAmounts.Int64(),
	}
	if info.PledgeInfo != nil {
		infos := make([]*pb.NEP5PledgeInfo, 0)
		for _, i := range info.PledgeInfo {
			it := &pb.NEP5PledgeInfo{
				PType:         i.PType,
				Amount:        i.Amount.Int64(),
				WithdrawTime:  i.WithdrawTime,
				Beneficial:    toAddressValue(i.Beneficial),
				PledgeAddress: toAddressValue(i.PledgeAddress),
				NEP5TxId:      i.NEP5TxId,
			}
			infos = append(infos, it)
		}
		r.PledgeInfos = infos
	}
	return r
}

func toOriginPledgeParam(param *pb.PledgeParam) (*api.PledgeParam, error) {
	bene, err := toOriginAddressByValue(param.GetBeneficial())
	if err != nil {
		return nil, err
	}
	pledge, err := toOriginAddressByValue(param.GetPledgeAddress())
	if err != nil {
		return nil, err
	}
	amount := toOriginBalanceByValue(param.GetAmount())
	return &api.PledgeParam{
		Beneficial:    bene,
		PledgeAddress: pledge,
		Amount:        amount,
		PType:         param.GetPType(),
		NEP5TxId:      param.GetNep5TxId(),
	}, nil
}

func toOriginWithdrawPledgeParam(param *pb.WithdrawPledgeParam) (*api.WithdrawPledgeParam, error) {
	bene, err := toOriginAddressByValue(param.GetBeneficial())
	if err != nil {
		return nil, err
	}
	amount := toOriginBalanceByValue(param.GetAmount())
	return &api.WithdrawPledgeParam{
		Beneficial: bene,
		Amount:     amount,
		PType:      param.GetPType(),
		NEP5TxId:   param.GetNep5TxId(),
	}, nil
}

func toNEP5PledgeInfo(info *api.NEP5PledgeInfo) *pb.NEP5PledgeInfo {
	return &pb.NEP5PledgeInfo{
		PType:         info.PType,
		Amount:        info.Amount.Int64(),
		WithdrawTime:  info.WithdrawTime,
		Beneficial:    toAddressValue(info.Beneficial),
		PledgeAddress: toAddressValue(info.PledgeAddress),
		NEP5TxId:      info.NEP5TxId,
	}
}

func toNEP5PledgeInfo2(info *abi.NEP5PledgeInfo) *pbtypes.NEP5PledgeInfo {
	return &pbtypes.NEP5PledgeInfo{
		PType:         int32(info.PType),
		Amount:        info.Amount.Int64(),
		WithdrawTime:  info.WithdrawTime,
		Beneficial:    toAddressValue(info.Beneficial),
		PledgeAddress: toAddressValue(info.PledgeAddress),
		NEP5TxId:      info.NEP5TxId,
	}
}

func toNEP5PledgeInfos(infos []*api.NEP5PledgeInfo) *pb.NEP5PledgeInfos {
	ps := make([]*pb.NEP5PledgeInfo, 0)
	for _, p := range infos {
		pt := toNEP5PledgeInfo(p)
		ps = append(ps, pt)
	}
	return &pb.NEP5PledgeInfos{PledgeInfos: ps}
}
