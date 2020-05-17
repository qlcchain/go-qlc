package grpcServer

import (
	"context"
	"fmt"
	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/log"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/rpc/grpc/proto"
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

func (l *LedgerAPI) NewBlock(*empty.Empty, proto.LedgerAPI_NewBlockServer) error {
	panic("implement me")
}

func (l *LedgerAPI) NewAccountBlock(*proto.Address, proto.LedgerAPI_NewAccountBlockServer) error {
	panic("implement me")
}

func (l *LedgerAPI) BalanceChange(*proto.Address, proto.LedgerAPI_BalanceChangeServer) error {
	panic("implement me")
}

func (l *LedgerAPI) NewPending(*proto.Address, proto.LedgerAPI_NewPendingServer) error {
	panic("implement me")
}

func (l *LedgerAPI) Test(context.Context, *empty.Empty) (*proto.TestResponse, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountBlocksCount(ctx context.Context, addr *proto.Address) (*proto.Int64, error) {
	fmt.Println(addr.Address)
	return &proto.Int64{
		I: 1,
	}, nil
}

func (l *LedgerAPI) AccountHistoryTopn(context.Context, *proto.AccountHistoryTopnRequest) (*proto.APIBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountInfo(context.Context, *proto.Address) (*proto.APIAccount, error) {
	panic("implement me")
}

func (l *LedgerAPI) ConfirmedAccountInfo(context.Context, *proto.Address) (*proto.APIAccount, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountRepresentative(context.Context, *proto.Address) (*proto.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountVotingWeight(context.Context, *proto.Address) (*proto.Balance, error) {
	panic("implement me")
}

func (l *LedgerAPI) AccountsCount(context.Context, *empty.Empty) (*proto.UInt64, error) {
	panic("implement me")
}

func (l *LedgerAPI) Accounts(context.Context, *proto.Offset) (*proto.Addresses, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlockAccount(context.Context, *proto.Hash) (*proto.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlockConfirmedStatus(context.Context, *proto.Hash) (*proto.Boolean, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlockHash(context.Context, *proto.StateBlock) (*proto.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) BlocksInfo(context.Context, *proto.Hashes) (*proto.APIBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) ConfirmedBlocksInfo(context.Context, *proto.Hashes) (*proto.APIBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) Blocks(context.Context, *proto.Offset) (*proto.APIBlocks, error) {
	panic("implement me")
}

func (l *LedgerAPI) Chain(context.Context, *proto.ChainRequest) (*proto.Hashes, error) {
	panic("implement me")
}

func (l *LedgerAPI) Delegators(context.Context, *proto.Address) (*proto.APIAccountBalances, error) {
	panic("implement me")
}

func (l *LedgerAPI) DelegatorsCount(context.Context, *proto.Address) (*proto.Int64, error) {
	panic("implement me")
}

func (l *LedgerAPI) Pendings(context.Context, *empty.Empty) (*proto.APIPendings, error) {
	panic("implement me")
}

func (l *LedgerAPI) Representatives(context.Context, *proto.Boolean) (*proto.APIRepresentatives, error) {
	panic("implement me")
}

func (l *LedgerAPI) Tokens(context.Context, *empty.Empty) (*proto.TokenInfos, error) {
	panic("implement me")
}

func (l *LedgerAPI) TokenInfoById(context.Context, *proto.Hash) (*proto.TokenInfo, error) {
	panic("implement me")
}

func (l *LedgerAPI) TokenInfoByName(context.Context, *proto.String) (*proto.TokenInfo, error) {
	panic("implement me")
}

func (l *LedgerAPI) GetAccountOnlineBlock(context.Context, *proto.Address) (*proto.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisAddress(context.Context, *empty.Empty) (*proto.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasAddress(context.Context, *empty.Empty) (*proto.Address, error) {
	panic("implement me")
}

func (l *LedgerAPI) ChainToken(context.Context, *empty.Empty) (*proto.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasToken(context.Context, *empty.Empty) (*proto.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisMintageBlock(context.Context, *empty.Empty) (*proto.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisMintageHash(context.Context, *empty.Empty) (*proto.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisBlock(context.Context, *empty.Empty) (*proto.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GenesisBlockHash(context.Context, *empty.Empty) (*proto.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasBlockHash(context.Context, *empty.Empty) (*proto.Hash, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasMintageBlock(context.Context, *empty.Empty) (*proto.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) GasBlock(context.Context, *empty.Empty) (*proto.StateBlock, error) {
	panic("implement me")
}

func (l *LedgerAPI) IsGenesisBlock(context.Context, *proto.StateBlock) (*proto.Boolean, error) {
	panic("implement me")
}

func (l *LedgerAPI) IsGenesisToken(context.Context, *proto.StateBlock) (*proto.Boolean, error) {
	panic("implement me")
}

func (l *LedgerAPI) AllGenesisBlocks(context.Context, *proto.StateBlock) (*proto.StateBlocks, error) {
	panic("implement me")
}
