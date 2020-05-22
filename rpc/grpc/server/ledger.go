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

func (l *LedgerAPI) AccountBlocksCount(ctx context.Context, addr *pb.Address) (*pb.Int64, error) {
	fmt.Println(addr.Address)
	return &pb.Int64{
		I: 1,
	}, nil
}

func (l *LedgerAPI) AccountHistoryTopn(context.Context, *pb.AccountHistoryTopnRequest) (*pb.APIBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountInfo(context.Context, *pb.Address) (*pb.APIAccount, error) {
	panic("implement me")
}

func (l *LedgerAPI) ConfirmedAccountInfo(context.Context, *pb.Address) (*pb.APIAccount, error) {
	return &pb.APIAccount{
		Address: "qlc_1c6xxzrjgy7qp6oehkjf9gbbeonmupg68zqszfp1okfm3k5y5738hahaus8o",
		Tokens: []*pb.APITokenMeta{
			{
				TokenName: "QLC",
				TokenMeta: &pb.TokenMeta{
					Type: "3f3ab7b5657ce9a7786feea51a0c7660f63bff2e7abb27c6c6ff6a8d3a1955b2",
				},
			},
		},
	}, nil
}

func (l *LedgerAPI) AccountRepresentative(context.Context, *pb.Address) (*pb.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountVotingWeight(context.Context, *pb.Address) (*pb.Balance, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountsCount(context.Context, *empty.Empty) (*pb.UInt64, error) {
	panic("implement me")
}

func (l *LedgerAPI) Accounts(context.Context, *pb.Offset) (*pb.Addresses, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountsBalance(ctx context.Context, addresses *pb.Addresses) (*pb.AccountsBalanceResponse, error) {
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

func (l *LedgerAPI) AccountsFrontiers(context.Context, *pb.Addresses) (*pb.AccountsFrontiersResponse, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountsPending(context.Context, *pb.AccountsPendingRequest) (*pb.AccountsPendingResponse, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlockAccount(context.Context, *pb.Hash) (*pb.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlockConfirmedStatus(context.Context, *pb.Hash) (*pb.Boolean, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlockHash(context.Context, *pb.StateBlock) (*pb.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlocksInfo(context.Context, *pb.Hashes) (*pb.APIBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) ConfirmedBlocksInfo(context.Context, *pb.Hashes) (*pb.APIBlocks, error) {
	blk := mock.StateBlockWithoutWork()
	return &pb.APIBlocks{
		Blocks: []*pb.APIBlock{
			{
				Block: &pb.StateBlock{
					Token:   blk.Token.String(),
					Address: blk.Address.String(),
				},
				TokenName: "QLC",
				Hash:      "bb3ecfb70bf8905120c5cf815ad5ecc1a7195845099aec9f49d2889ecba4243d",
			},
			{
				Block: &pb.StateBlock{
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

func (l *LedgerAPI) Chain(context.Context, *pb.ChainRequest) (*pb.Hashes, error) {
	panic("implement me")
}

func (l *LedgerAPI) Delegators(context.Context, *pb.Address) (*pb.APIAccountBalances, error) {
	panic("implement me")
}

func (l *LedgerAPI) DelegatorsCount(context.Context, *pb.Address) (*pb.Int64, error) {
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

func (l *LedgerAPI) Tokens(context.Context, *empty.Empty) (*pb.TokenInfos, error) {
	panic("implement me")
}

func (l *LedgerAPI) TokenInfoById(context.Context, *pb.Hash) (*pb.TokenInfo, error) {
	panic("implement me")
}

func (l *LedgerAPI) TokenInfoByName(context.Context, *pb.String) (*pb.TokenInfo, error) {
	panic("implement me")
}

func (l *LedgerAPI) GetAccountOnlineBlock(context.Context, *pb.Address) (*pb.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisAddress(context.Context, *empty.Empty) (*pb.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasAddress(context.Context, *empty.Empty) (*pb.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) ChainToken(context.Context, *empty.Empty) (*pb.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasToken(context.Context, *empty.Empty) (*pb.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisMintageBlock(context.Context, *empty.Empty) (*pb.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisMintageHash(context.Context, *empty.Empty) (*pb.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisBlock(context.Context, *empty.Empty) (*pb.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisBlockHash(context.Context, *empty.Empty) (*pb.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasBlockHash(context.Context, *empty.Empty) (*pb.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasMintageBlock(context.Context, *empty.Empty) (*pb.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasBlock(context.Context, *empty.Empty) (*pb.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) IsGenesisBlock(context.Context, *pb.StateBlock) (*pb.Boolean, error) {
	panic("implement me")
}

func (l *LedgerAPI) IsGenesisToken(context.Context, *pb.StateBlock) (*pb.Boolean, error) {
	panic("implement me")
}

func (l *LedgerAPI) AllGenesisBlocks(context.Context, *pb.StateBlock) (*pb.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenerateSendBlock(context.Context, *pb.GenerateSendBlockRequest) (*pb.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenerateReceiveBlock(context.Context, *pb.GenerateReceiveBlockRequest) (*pb.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenerateReceiveBlockByHash(context.Context, *pb.GenerateReceiveBlockByHashRequest) (*pb.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenerateChangeBlock(context.Context, *pb.GenerateChangeBlockRequest) (*pb.StateBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) Process(context.Context, *pb.StateBlock) (*pb.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) NewBlock(*empty.Empty, pb.LedgerAPI_NewBlockServer) error {
	panic("implement me")
}

func (l *LedgerAPI) NewAccountBlock(*pb.Address, pb.LedgerAPI_NewAccountBlockServer) error {
	panic("implement me")
}

func (l *LedgerAPI) BalanceChange(*pb.Address, pb.LedgerAPI_BalanceChangeServer) error {
	panic("implement me")
}

func (l *LedgerAPI) NewPending(*pb.Address, pb.LedgerAPI_NewPendingServer) error {
	panic("implement me")
}
