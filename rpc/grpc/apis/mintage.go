package apis

import (
	"context"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type MintageAPI struct {
	mintage *api.MintageAPI
	logger  *zap.SugaredLogger
}

func NewMintageAPI(cfgFile string, l ledger.Store) *MintageAPI {
	return &MintageAPI{
		mintage: api.NewMintageApi(cfgFile, l),
		logger:  log.NewLogger("grpc_mintage"),
	}
}

func (m *MintageAPI) GetMintageData(ctx context.Context, param *pb.MintageParams) (*pb.Bytes, error) {
	p, err := toOriginMintageParams(param)
	if err != nil {
		return nil, err
	}
	r, err := m.mintage.GetMintageData(p)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (m *MintageAPI) GetMintageBlock(ctx context.Context, param *pb.MintageParams) (*pbtypes.StateBlock, error) {
	p, err := toOriginMintageParams(param)
	if err != nil {
		return nil, err
	}
	r, err := m.mintage.GetMintageBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (m *MintageAPI) GetRewardBlock(ctx context.Context, param *pbtypes.StateBlock) (*pbtypes.StateBlock, error) {
	block, err := toOriginStateBlock(param)
	if err != nil {
		return nil, err
	}
	r, err := m.mintage.GetRewardBlock(block)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (m *MintageAPI) GetWithdrawMintageData(ctx context.Context, param *pbtypes.Hash) (*pb.Bytes, error) {
	hash, err := toOriginHash(param)
	if err != nil {
		return nil, err
	}
	r, err := m.mintage.GetWithdrawMintageData(hash)
	if err != nil {
		return nil, err
	}
	return &pb.Bytes{
		Value: r,
	}, nil
}

func (m *MintageAPI) ParseTokenInfo(ctx context.Context, param *pb.Bytes) (*pbtypes.TokenInfo, error) {
	r, err := m.mintage.ParseTokenInfo(param.GetValue())
	if err != nil {
		return nil, err
	}
	return toTokenInfo(*r), nil
}

func (m *MintageAPI) GetWithdrawMintageBlock(ctx context.Context, param *pb.WithdrawParams) (*pbtypes.StateBlock, error) {
	p, err := toOriginWithdrawParams(param)
	if err != nil {
		return nil, err
	}
	r, err := m.mintage.GetWithdrawMintageBlock(p)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (m *MintageAPI) GetWithdrawRewardBlock(ctx context.Context, param *pbtypes.StateBlock) (*pbtypes.StateBlock, error) {
	blk, err := toOriginStateBlock(param)
	if err != nil {
		return nil, err
	}
	r, err := m.mintage.GetWithdrawRewardBlock(blk)
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func toOriginMintageParams(param *pb.MintageParams) (*api.MintageParams, error) {
	selfAddr, err := toOriginAddressByValue(param.GetSelfAddr())
	if err != nil {
		return nil, err
	}
	prevHash, err := toOriginHashByValue(param.GetPrevHash())
	if err != nil {
		return nil, err
	}
	benefit, err := toOriginAddressByValue(param.GetBeneficial())
	if err != nil {
		return nil, err
	}
	return &api.MintageParams{
		SelfAddr:    selfAddr,
		PrevHash:    prevHash,
		TokenName:   param.GetTokenName(),
		TokenSymbol: param.GetTokenSymbol(),
		TotalSupply: param.GetTotalSupply(),
		Decimals:    uint8(param.GetDecimals()),
		Beneficial:  benefit,
		NEP5TxId:    param.GetNEP5TxId(),
	}, nil
}

func toOriginWithdrawParams(param *pb.WithdrawParams) (*api.WithdrawParams, error) {
	selfAddr, err := toOriginAddressByValue(param.GetSelfAddr())
	if err != nil {
		return nil, err
	}
	tokenId, err := toOriginHashByValue(param.GetTokenId())
	if err != nil {
		return nil, err
	}
	return &api.WithdrawParams{
		SelfAddr: selfAddr,
		TokenId:  tokenId,
	}, nil
}
