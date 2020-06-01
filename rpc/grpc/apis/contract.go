package apis

import (
	"context"

	"math/big"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type ContractAPI struct {
	contract *api.ContractApi
	logger   *zap.SugaredLogger
}

func NewContractAPI(cc *chainctx.ChainContext, l ledger.Store) *ContractAPI {
	return &ContractAPI{
		contract: api.NewContractApi(cc, l),
		logger:   log.NewLogger("grpc_contract"),
	}
}

func (c *ContractAPI) GetAbiByContractAddress(ctx context.Context, param *pbtypes.Address) (*pb.String, error) {
	addr, err := toOriginAddress(param)
	if err != nil {
		return nil, err
	}
	r, err := c.contract.GetAbiByContractAddress(addr)
	if err != nil {
		return nil, err
	}
	return &pb.String{
		Value: r,
	}, nil
}

func (c *ContractAPI) PackContractData(ctx context.Context, param *pb.PackContractDataRequest) (*pb.Bytes, error) {
	abiStr := param.GetAbiStr()
	methodStr := param.GetMethodName()
	params := param.GetParams()
	r, err := c.contract.PackContractData(abiStr, methodStr, params)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (c *ContractAPI) PackChainContractData(ctx context.Context, param *pb.PackChainContractDataRequest) (*pb.Bytes, error) {
	addr := param.GetContractAddress()
	address, err := types.HexToAddress(addr)
	if err != nil {
		return nil, err
	}
	methodStr := param.GetMethodName()
	params := param.GetParams()
	r, err := c.contract.PackChainContractData(address, methodStr, params)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (c *ContractAPI) ContractAddressList(ctx context.Context, param *empty.Empty) (*pbtypes.Addresses, error) {
	r := c.contract.ContractAddressList()
	return toAddresses(r), nil
}

func (c *ContractAPI) GenerateSendBlock(ctx context.Context, param *pb.ContractSendBlockPara) (*pbtypes.StateBlock, error) {
	addr, err := types.HexToAddress(param.GetAddress())
	if err != nil {
		return nil, err
	}
	to, err := types.HexToAddress(param.GetTo())
	if err != nil {
		return nil, err
	}
	sendPara := &api.ContractSendBlockPara{
		Address:        addr,
		TokenName:      param.GetTokenName(),
		To:             to,
		Amount:         types.Balance{Int: big.NewInt(param.GetAmount())},
		Data:           param.GetData(),
		PrivateFrom:    param.GetPrivateFrom(),
		PrivateFor:     param.GetPrivateFor(),
		PrivateGroupID: param.GetPrivateGroupID(),
		EnclaveKey:     param.GetEnclaveKey(),
	}
	r, err := c.contract.GenerateSendBlock(sendPara)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (c *ContractAPI) GenerateRewardBlock(ctx context.Context, param *pb.ContractRewardBlockPara) (*pbtypes.StateBlock, error) {
	hash, err := types.NewHash(param.GetSendHash())
	if err != nil {
		return nil, err
	}
	rewardPara := &api.ContractRewardBlockPara{
		SendHash: hash,
	}
	r, err := c.contract.GenerateRewardBlock(rewardPara)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}
