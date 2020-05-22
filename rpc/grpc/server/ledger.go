package grpcServer

import (
	"context"
	"fmt"
	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/mock"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/ledger"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	"go.uber.org/zap"
)

type LedgerAPI struct {
	ledger ledger.Store
	eb     event.EventBus
	logger *zap.SugaredLogger
}

func NewLedgerApi(ctx context.Context, l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *LedgerAPI {
	api := LedgerAPI{
		ledger: l,
		eb:     eb,
		logger: log.NewLogger("api_ledger"),
	}
	return &api
}

func (l *LedgerAPI) Test(context.Context, *empty.Empty) (*pb.TestResponse, error) {
	return &pb.TestResponse{
		Version: "123",
	}, nil
}

func (l *LedgerAPI) AccountBlocksCount(ctx context.Context, addr *pbtypes.Address) (*pb.Int64, error) {
	fmt.Println(addr.Address)
	return &pb.Int64{
		I: 1,
	}, nil
}

func (l *LedgerAPI) AccountHistoryTopn(context.Context, *pb.AccountHistoryTopnRequest) (*pb.APIBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountInfo(context.Context, *pbtypes.Address) (*pb.APIAccount, error) {
	panic("implement me")
}

func (l *LedgerAPI) ConfirmedAccountInfo(context.Context, *pbtypes.Address) (*pb.APIAccount, error) {
	return &pb.APIAccount{
		Address: "qlc_1c6xxzrjgy7qp6oehkjf9gbbeonmupg68zqszfp1okfm3k5y5738hahaus8o",
		Tokens: []*pb.APITokenMeta{
			{
				TokenName: "QLC",
				TokenMeta: &pbtypes.TokenMeta{
					Type: "3f3ab7b5657ce9a7786feea51a0c7660f63bff2e7abb27c6c6ff6a8d3a1955b2",
				},
			},
		},
	}, nil
}

func (l *LedgerAPI) AccountRepresentative(context.Context, *pbtypes.Address) (*pbtypes.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountVotingWeight(context.Context, *pbtypes.Address) (*pbtypes.Balance, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountsCount(context.Context, *empty.Empty) (*pb.UInt64, error) {
	panic("implement me")
}

func (l *LedgerAPI) Accounts(context.Context, *pb.Offset) (*pbtypes.Addresses, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountsBalance(ctx context.Context, addresses *pbtypes.Addresses) (*pb.AccountsBalanceResponse, error) {
	return &pb.AccountsBalanceResponse{
		AccountsBalances: map[string]*pb.AccountsBalanceResponseBalances{
			"qlc_1c6xxzrjgy7qp6oehkjf9gbbeonmupg68zqszfp1okfm3k5y5738hahaus8o": {
				Balances: map[string]*pb.AccountsBalanceResponse_APIAccountsBalance{
					"3f3ab7b5657ce9a7786feea51a0c7660f63bff2e7abb27c6c6ff6a8d3a1955b2": {
						Balance: 100,
						Vote:    100,
					},
					"660f63bff2e7abb27c6c6ff6a8d3a1955b24f3ab7b5657ce9a7786feea51a0c7": {
						Balance: 200,
						Vote:    200,
					},
				},
			},
			"qlc_upg68zqszfp1okfm3k5y5738hahaus8o1c6xxzrjgy7qp6oehkjf9gbbeonm": {
				Balances: map[string]*pb.AccountsBalanceResponse_APIAccountsBalance{
					"3f3ab7b5657ce9a7786feea51a0c7660f63bff2e7abb27c6c6ff6a8d3a1955b2": {
						Balance: 300,
						Vote:    100,
					},
				},
			},
		},
	}, nil
}

func (l *LedgerAPI) AccountsFrontiers(context.Context, *pbtypes.Addresses) (*pb.AccountsFrontiersResponse, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountsPending(context.Context, *pb.AccountsPendingRequest) (*pb.AccountsPendingResponse, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlockAccount(context.Context, *pbtypes.Hash) (*pbtypes.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlockConfirmedStatus(context.Context, *pbtypes.Hash) (*pb.Boolean, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlockHash(context.Context, *pbtypes.StateBlock) (*pbtypes.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlocksInfo(context.Context, *pbtypes.Hashes) (*pb.APIBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) ConfirmedBlocksInfo(context.Context, *pbtypes.Hashes) (*pb.APIBlocks, error) {
	blk := mock.StateBlockWithoutWork()
	return &pb.APIBlocks{
		Blocks: []*pb.APIBlock{
			{
				Block: &pbtypes.StateBlock{
					Token:   blk.Token.String(),
					Address: blk.Address.String(),
				},
				TokenName: "QLC",
				Hash:      "bb3ecfb70bf8905120c5cf815ad5ecc1a7195845099aec9f49d2889ecba4243d",
			},
			{
				Block: &pbtypes.StateBlock{
					Token:   blk.Token.String(),
					Address: blk.Address.String(),
				},
				TokenName: "QLC",
				Hash:      "bb3ecfb70bf8905120c5cf815ad5ecc1a7195845099aec9f49d2889ecba4243d",
			},
		},
	}, nil
}

func (l *LedgerAPI) Blocks(context.Context, *pb.Offset) (*pb.APIBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlocksCount(context.Context, *empty.Empty) (*pb.BlocksCountResponse, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlocksCount2(context.Context, *empty.Empty) (*pb.BlocksCountResponse, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlocksCountByType(context.Context, *empty.Empty) (*pb.BlocksCountResponse, error) {
	panic("implement me")
}

func (l *LedgerAPI) Chain(context.Context, *pb.ChainRequest) (*pbtypes.Hashes, error) {
	panic("implement me")
}

func (l *LedgerAPI) Delegators(context.Context, *pbtypes.Address) (*pb.APIAccountBalances, error) {
	panic("implement me")
}

func (l *LedgerAPI) DelegatorsCount(context.Context, *pbtypes.Address) (*pb.Int64, error) {
	panic("implement me")
}

func (l *LedgerAPI) Pendings(context.Context, *empty.Empty) (*pb.APIPendings, error) {
	panic("implement me")
}

func (l *LedgerAPI) Representatives(context.Context, *pb.Boolean) (*pb.APIRepresentatives, error) {
	panic("implement me")
}

func (l *LedgerAPI) TransactionsCount(context.Context, *empty.Empty) (*pb.BlocksCountResponse, error) {
	panic("implement me")
}

func (l *LedgerAPI) Tokens(context.Context, *empty.Empty) (*pbtypes.TokenInfos, error) {
	panic("implement me")
}

func (l *LedgerAPI) TokenInfoById(context.Context, *pbtypes.Hash) (*pbtypes.TokenInfo, error) {
	panic("implement me")
}

func (l *LedgerAPI) TokenInfoByName(context.Context, *pb.String) (*pbtypes.TokenInfo, error) {
	panic("implement me")
}

func (l *LedgerAPI) GetAccountOnlineBlock(context.Context, *pbtypes.Address) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisAddress(context.Context, *empty.Empty) (*pbtypes.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasAddress(context.Context, *empty.Empty) (*pbtypes.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) ChainToken(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasToken(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisMintageBlock(context.Context, *empty.Empty) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisMintageHash(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisBlock(context.Context, *empty.Empty) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisBlockHash(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasBlockHash(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasMintageBlock(context.Context, *empty.Empty) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasBlock(context.Context, *empty.Empty) (*pbtypes.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) IsGenesisBlock(context.Context, *pbtypes.StateBlock) (*pb.Boolean, error) {
	panic("implement me")
}

func (l *LedgerAPI) IsGenesisToken(context.Context, *pbtypes.StateBlock) (*pb.Boolean, error) {
	panic("implement me")
}

func (l *LedgerAPI) AllGenesisBlocks(context.Context, *pbtypes.StateBlock) (*pbtypes.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenerateSendBlock(context.Context, *pb.GenerateSendBlockRequest) (*pbtypes.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenerateReceiveBlock(context.Context, *pb.GenerateReceiveBlockRequest) (*pbtypes.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenerateReceiveBlockByHash(context.Context, *pb.GenerateReceiveBlockByHashRequest) (*pbtypes.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenerateChangeBlock(context.Context, *pb.GenerateChangeBlockRequest) (*pbtypes.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) Process(context.Context, *pbtypes.StateBlock) (*pbtypes.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) NewBlock(*empty.Empty, pb.LedgerAPI_NewBlockServer) error {
	panic("implement me")
}

func (l *LedgerAPI) NewAccountBlock(*pbtypes.Address, pb.LedgerAPI_NewAccountBlockServer) error {
	panic("implement me")
}

func (l *LedgerAPI) BalanceChange(*pbtypes.Address, pb.LedgerAPI_BalanceChangeServer) error {
	panic("implement me")
}

func (l *LedgerAPI) NewPending(*pbtypes.Address, pb.LedgerAPI_NewPendingServer) error {
	panic("implement me")
}
