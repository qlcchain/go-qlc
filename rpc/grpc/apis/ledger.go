package apis

import (
	"context"
	"github.com/golang/protobuf/ptypes/wrappers"
	"math/big"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	pbtypes "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
)

type LedgerAPI struct {
	ledger *api.LedgerAPI
	logger *zap.SugaredLogger
}

func NewLedgerApi(ctx context.Context, l ledger.Store, eb event.EventBus, cc *chainctx.ChainContext) *LedgerAPI {
	ledgerApi := LedgerAPI{
		ledger: api.NewLedgerApi(ctx, l, eb, cc),
		logger: log.NewLogger("grpc_ledger"),
	}

	return &ledgerApi
}

func (l *LedgerAPI) AccountBlocksCount(ctx context.Context, addr *pbtypes.Address) (*wrappers.Int64Value, error) {
	address, err := types.HexToAddress(addr.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.AccountBlocksCount(address)
	if err != nil {
		return nil, err
	}
	return &wrappers.Int64Value{Value: r}, nil
}

func (l *LedgerAPI) AccountHistoryTopn(ctx context.Context, para *pb.AccountHistoryTopnReq) (*pb.APIBlocks, error) {
	address, err := types.HexToAddress(para.GetAddress())
	if err != nil {
		return nil, err
	}
	count := int(para.GetCount())
	offset := int(para.GetOffset())
	r, err := l.ledger.AccountHistoryTopn(address, count, &offset)
	if err != nil {
		return nil, err
	}
	return toAPIBlocks(r), nil
}

func (l *LedgerAPI) AccountInfo(ctx context.Context, addr *pbtypes.Address) (*pb.APIAccount, error) {
	address, err := types.HexToAddress(addr.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.AccountInfo(address)
	if err != nil {
		return nil, err
	}
	return toAPIAccount(r), nil
}

func (l *LedgerAPI) ConfirmedAccountInfo(ctx context.Context, addr *pbtypes.Address) (*pb.APIAccount, error) {
	//return &pb.APIAccount{
	//	Address: "qlc_1c6xxzrjgy7qp6oehkjf9gbbeonmupg68zqszfp1okfm3k5y5738hahaus8o",
	//	Tokens: []*pb.APITokenMeta{
	//		{
	//			TokenName: "QLC",
	//			TokenMeta: &pbtypes.TokenMeta{
	//				Type: "3f3ab7b5657ce9a7786feea51a0c7660f63bff2e7abb27c6c6ff6a8d3a1955b2",
	//			},
	//		},
	//	},
	//}, nil
	address, err := types.HexToAddress(addr.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.ConfirmedAccountInfo(address)
	if err != nil {
		return nil, err
	}
	return toAPIAccount(r), nil
}

func (l *LedgerAPI) AccountRepresentative(ctx context.Context, addr *pbtypes.Address) (*pbtypes.Address, error) {
	address, err := types.HexToAddress(addr.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.AccountRepresentative(address)
	if err != nil {
		return nil, err
	}
	return &pbtypes.Address{
		Address: r.String(),
	}, nil
}

func (l *LedgerAPI) AccountVotingWeight(ctx context.Context, addr *pbtypes.Address) (*pbtypes.Balance, error) {
	address, err := types.HexToAddress(addr.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.AccountVotingWeight(address)
	if err != nil {
		return nil, err
	}
	return &pbtypes.Balance{
		Balance: r.Int64(),
	}, nil
}

func (l *LedgerAPI) AccountsCount(context.Context, *empty.Empty) (*wrappers.UInt64Value, error) {
	r, err := l.ledger.AccountsCount()
	if err != nil {
		return nil, err
	}
	return &wrappers.UInt64Value{
		Value: r,
	}, nil
}

func (l *LedgerAPI) Accounts(ctx context.Context, para *pb.Offset) (*pbtypes.Addresses, error) {
	count := int(para.GetCount())
	offset := int(para.GetOffset())
	r, err := l.ledger.Accounts(count, &offset)
	if err != nil {
		return nil, err
	}
	return toAddresses(r), nil
}

func (l *LedgerAPI) AccountsBalance(ctx context.Context, addresses *pbtypes.Addresses) (*pb.AccountsBalanceRsp, error) {
	addrs, err := toOriginAddresses(addresses)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.AccountsBalance(addrs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*pb.AccountsBalanceRspBalances)
	for addr, info := range r {
		fi := make(map[string]*pb.AccountsBalanceRsp_APIAccountsBalance)
		for tokenName, ba := range info {
			pt := &pb.AccountsBalanceRsp_APIAccountsBalance{
				Balance: ba.Balance.Int64(),
				Pending: ba.Pending.Int64(),
			}
			if ba.Vote != nil {
				pt.Vote = ba.Vote.Int64()
			}
			if ba.Network != nil {
				pt.Network = ba.Network.Int64()
			}
			if ba.Storage != nil {
				pt.Storage = ba.Storage.Int64()
			}
			if ba.Oracle != nil {
				pt.Oracle = ba.Oracle.Int64()
			}
			fi[tokenName] = pt
		}
		aff := &pb.AccountsBalanceRspBalances{Balances: fi}
		result[addr.String()] = aff
	}
	return &pb.AccountsBalanceRsp{AccountsBalances: result}, nil
	//return &pb.AccountsBalanceRsp{
	//	AccountsBalances: map[string]*pb.AccountsBalanceRspBalances{
	//		"qlc_1c6xxzrjgy7qp6oehkjf9gbbeonmupg68zqszfp1okfm3k5y5738hahaus8o": {
	//			Balances: map[string]*pb.AccountsBalanceRsp_APIAccountsBalance{
	//				"3f3ab7b5657ce9a7786feea51a0c7660f63bff2e7abb27c6c6ff6a8d3a1955b2": {
	//					Balance: 100,
	//					Vote:    100,
	//				},
	//				"660f63bff2e7abb27c6c6ff6a8d3a1955b24f3ab7b5657ce9a7786feea51a0c7": {
	//					Balance: 200,
	//					Vote:    200,
	//				},
	//			},
	//		},
	//		"qlc_upg68zqszfp1okfm3k5y5738hahaus8o1c6xxzrjgy7qp6oehkjf9gbbeonm": {
	//			Balances: map[string]*pb.AccountsBalanceRsp_APIAccountsBalance{
	//				"3f3ab7b5657ce9a7786feea51a0c7660f63bff2e7abb27c6c6ff6a8d3a1955b2": {
	//					Balance: 300,
	//					Vote:    100,
	//				},
	//			},
	//		},
	//	},
	//}, nil
}

func (l *LedgerAPI) AccountsFrontiers(ctx context.Context, addresses *pbtypes.Addresses) (*pb.AccountsFrontiersRsp, error) {
	addrs, err := toOriginAddresses(addresses)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.AccountsFrontiers(addrs)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*pb.AccountsFrontiersRspFrontier)
	for addr, frontier := range r {
		fi := make(map[string]string)
		for tokenName, header := range frontier {
			fi[tokenName] = header.String()
		}
		aff := &pb.AccountsFrontiersRspFrontier{Frontier: fi}
		result[addr.String()] = aff
	}
	return &pb.AccountsFrontiersRsp{AccountsFrontiers: result}, nil
}

func (l *LedgerAPI) AccountsPending(ctx context.Context, ap *pb.AccountsPendingReq) (*pb.AccountsPendingRsp, error) {
	addrs, err := toOriginAddresses(ap.GetAddresses())
	if err != nil {
		return nil, err
	}
	n := int(ap.GetCount())
	r, err := l.ledger.AccountsPending(addrs, n)
	if err != nil {
		return nil, err
	}
	aps := &pb.AccountsPendingRsp{}
	aps.AccountsPendings = make(map[string]*pb.APIPendings)
	for addr, pendings := range r {
		aps.AccountsPendings[addr.String()] = toAPIPendings(pendings)
	}
	return aps, nil
}

func (l *LedgerAPI) BlockAccount(ctx context.Context, hash *pbtypes.Hash) (*pbtypes.Address, error) {
	h, err := toOriginHash(hash)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.BlockAccount(h)
	if err != nil {
		return nil, err
	}
	return toAddress(r), nil
}

func (l *LedgerAPI) BlockConfirmedStatus(ctx context.Context, hash *pbtypes.Hash) (*wrappers.BoolValue, error) {
	h, err := toOriginHash(hash)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.BlockConfirmedStatus(h)
	if err != nil {
		return nil, err
	}
	return &wrappers.BoolValue{
		Value: r,
	}, nil
}

func (l *LedgerAPI) BlockHash(ctx context.Context, block *pbtypes.StateBlock) (*pbtypes.Hash, error) {
	blk, err := toOriginStateBlock(block)
	if err != nil {
		return nil, err
	}
	r := l.ledger.BlockHash(*blk)
	return &pbtypes.Hash{
		Hash: r.String(),
	}, nil
}

func (l *LedgerAPI) BlocksInfo(ctx context.Context, hashes *pbtypes.Hashes) (*pb.APIBlocks, error) {
	hs, err := toOriginHashes(hashes)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.BlocksInfo(hs)
	if err != nil {
		return nil, err
	}
	return toAPIBlocks(r), nil
}

func (l *LedgerAPI) ConfirmedBlocksInfo(ctx context.Context, hashes *pbtypes.Hashes) (*pb.APIBlocks, error) {
	hs, err := toOriginHashes(hashes)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.ConfirmedBlocksInfo(hs)
	if err != nil {
		return nil, err
	}
	return toAPIBlocks(r), nil
	//blk := mock.StateBlockWithoutWork()
	//return &pb.APIBlocks{
	//	Blocks: []*pb.APIBlock{
	//		{
	//			Block: &pbtypes.StateBlock{
	//				Token:   blk.Token.String(),
	//				Address: blk.Address.String(),
	//			},
	//			TokenName: "QLC",
	//			Hash:      "bb3ecfb70bf8905120c5cf815ad5ecc1a7195845099aec9f49d2889ecba4243d",
	//		},
	//		{
	//			Block: &pbtypes.StateBlock{
	//				Token:   blk.Token.String(),
	//				Address: blk.Address.String(),
	//			},
	//			TokenName: "QLC",
	//			Hash:      "bb3ecfb70bf8905120c5cf815ad5ecc1a7195845099aec9f49d2889ecba4243d",
	//		},
	//	},
	//}, nil
}

func (l *LedgerAPI) Blocks(ctx context.Context, offset *pb.Offset) (*pb.APIBlocks, error) {
	count := int(offset.GetCount())
	os := int(offset.GetOffset())
	r, err := l.ledger.Blocks(count, &os)
	if err != nil {
		return nil, err
	}
	return toAPIBlocks(r), nil
}

func (l *LedgerAPI) BlocksCount(context.Context, *empty.Empty) (*pb.BlocksCountRsp, error) {
	r, err := l.ledger.BlocksCount()
	if err != nil {
		return nil, err
	}
	return &pb.BlocksCountRsp{
		Count: r,
	}, nil
}

func (l *LedgerAPI) BlocksCount2(context.Context, *empty.Empty) (*pb.BlocksCountRsp, error) {
	r, err := l.ledger.BlocksCount2()
	if err != nil {
		return nil, err
	}
	return &pb.BlocksCountRsp{
		Count: r,
	}, nil
}

func (l *LedgerAPI) BlocksCountByType(context.Context, *empty.Empty) (*pb.BlocksCountRsp, error) {
	r, err := l.ledger.BlocksCountByType()
	if err != nil {
		return nil, err
	}
	return &pb.BlocksCountRsp{
		Count: r,
	}, nil
}

func (l *LedgerAPI) Chain(ctx context.Context, para *pb.ChainReq) (*pbtypes.Hashes, error) {
	hash, err := toOriginHash(para.GetHash())
	if err != nil {
		return nil, err
	}
	count := int(para.GetCount())
	r, err := l.ledger.Chain(hash, count)
	if err != nil {
		return nil, err
	}
	return toHashes(r), nil
}

func (l *LedgerAPI) Delegators(ctx context.Context, addr *pbtypes.Address) (*pb.APIAccountBalances, error) {
	address, err := toOriginAddress(addr)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.Delegators(address)
	if err != nil {
		return nil, err
	}
	result := make([]*pb.APIAccountBalances_APIAccountBalance, 0)
	for _, b := range r {
		bt := &pb.APIAccountBalances_APIAccountBalance{
			Address: b.Address.String(),
			Balance: b.Balance.Int64(),
		}
		result = append(result, bt)
	}
	return &pb.APIAccountBalances{Balances: result}, nil
}

func (l *LedgerAPI) DelegatorsCount(ctx context.Context, addr *pbtypes.Address) (*wrappers.Int64Value, error) {
	address, err := toOriginAddress(addr)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.DelegatorsCount(address)
	if err != nil {
		return nil, err
	}
	return &wrappers.Int64Value{
		Value: r,
	}, nil

}

func (l *LedgerAPI) Pendings(context.Context, *empty.Empty) (*pb.APIPendings, error) {
	r, err := l.ledger.Pendings()
	if err != nil {
		return nil, err
	}
	return toAPIPendings(r), nil
}

func (l *LedgerAPI) Representatives(ctx context.Context, b *wrappers.BoolValue) (*pb.APIRepresentatives, error) {
	t := b.GetValue()
	r, err := l.ledger.Representatives(&t)
	if err != nil {
		return nil, err
	}
	return toAPIRepresentatives(r), nil
}

func (l *LedgerAPI) TransactionsCount(context.Context, *empty.Empty) (*pb.BlocksCountRsp, error) {
	r, err := l.ledger.TransactionsCount()
	if err != nil {
		return nil, err
	}
	return &pb.BlocksCountRsp{
		Count: r,
	}, nil
}

func (l *LedgerAPI) Tokens(context.Context, *empty.Empty) (*pbtypes.TokenInfos, error) {
	r, err := l.ledger.Tokens()
	if err != nil {
		return nil, err
	}
	return toTokenInfos(r), nil
}

func (l *LedgerAPI) TokenInfoById(ctx context.Context, id *pbtypes.Hash) (*pbtypes.TokenInfo, error) {
	hash, err := toOriginHash(id)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.TokenInfoById(hash)
	if err != nil {
		return nil, err
	}
	return toTokenInfo(r.TokenInfo), nil
}

func (l *LedgerAPI) TokenInfoByName(ctx context.Context, name *wrappers.StringValue) (*pbtypes.TokenInfo, error) {
	n := name.GetValue()
	r, err := l.ledger.TokenInfoByName(n)
	if err != nil {
		return nil, err
	}
	return toTokenInfo(r.TokenInfo), nil
}

func (l *LedgerAPI) GetAccountOnlineBlock(ctx context.Context, addr *pbtypes.Address) (*pbtypes.StateBlocks, error) {
	address, err := toOriginAddress(addr)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.GetAccountOnlineBlock(address)
	if err != nil {
		return nil, err
	}
	return toStateBlocks(r), nil
}

func (l *LedgerAPI) GenesisAddress(context.Context, *empty.Empty) (*pbtypes.Address, error) {
	r := l.ledger.GenesisAddress()
	return toAddress(r), nil
}

func (l *LedgerAPI) GasAddress(context.Context, *empty.Empty) (*pbtypes.Address, error) {
	r := l.ledger.GasAddress()
	return toAddress(r), nil
}

func (l *LedgerAPI) ChainToken(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	r := l.ledger.ChainToken()
	return toHash(r), nil
}

func (l *LedgerAPI) GasToken(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	r := l.ledger.GasToken()
	return toHash(r), nil
}

func (l *LedgerAPI) GenesisMintageBlock(context.Context, *empty.Empty) (*pbtypes.StateBlock, error) {
	r := l.ledger.GenesisMintageBlock()
	return toStateBlock(&r), nil
}

func (l *LedgerAPI) GenesisMintageHash(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	r := l.ledger.GenesisMintageHash()
	return toHash(r), nil
}

func (l *LedgerAPI) GenesisBlock(context.Context, *empty.Empty) (*pbtypes.StateBlock, error) {
	r := l.ledger.GenesisBlock()
	return toStateBlock(&r), nil
}

func (l *LedgerAPI) GenesisBlockHash(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	r := l.ledger.GenesisBlockHash()
	return toHash(r), nil
}

func (l *LedgerAPI) GasBlockHash(context.Context, *empty.Empty) (*pbtypes.Hash, error) {
	r := l.ledger.GasBlockHash()
	return toHash(r), nil
}

func (l *LedgerAPI) GasMintageBlock(context.Context, *empty.Empty) (*pbtypes.StateBlock, error) {
	r := l.ledger.GasMintageBlock()
	return toStateBlock(&r), nil
}

func (l *LedgerAPI) GasBlock(context.Context, *empty.Empty) (*pbtypes.StateBlock, error) {
	r := l.ledger.GasBlock()
	return toStateBlock(&r), nil
}

func (l *LedgerAPI) IsGenesisBlock(ctx context.Context, block *pbtypes.StateBlock) (*wrappers.BoolValue, error) {
	blk, err := toOriginStateBlock(block)
	if err != nil {
		return nil, err
	}
	r := l.ledger.IsGenesisBlock(blk)
	return &wrappers.BoolValue{
		Value: r,
	}, nil
}

func (l *LedgerAPI) IsGenesisToken(ctx context.Context, hash *pbtypes.Hash) (*wrappers.BoolValue, error) {
	h, err := toOriginHash(hash)
	if err != nil {
		return nil, err
	}
	r := l.ledger.IsGenesisToken(h)
	return &wrappers.BoolValue{
		Value: r,
	}, nil
}

func (l *LedgerAPI) AllGenesisBlocks(context.Context, *empty.Empty) (*pbtypes.StateBlocks, error) {
	r := l.ledger.AllGenesisBlocks()
	blocks := make([]*types.StateBlock, 0)
	for _, b := range r {
		blocks = append(blocks, &b)
	}
	return toStateBlocks(blocks), nil
}

func (l *LedgerAPI) GenerateSendBlock(ctx context.Context, para *pb.GenerateSendBlockReq) (*pbtypes.StateBlock, error) {
	from, err := types.HexToAddress(para.GetPara().GetFrom())
	if err != nil {
		return nil, err
	}
	to, err := types.HexToAddress(para.GetPara().GetTo())
	if err != nil {
		return nil, err
	}
	message := types.ZeroHash
	if para.GetPara().GetMessage() != "" {
		message, err = types.NewHash(para.GetPara().GetMessage())
		if err != nil {
			return nil, err
		}
	}
	r, err := l.ledger.GenerateSendBlock(&api.APISendBlockPara{
		From:      from,
		TokenName: para.GetPara().GetTokenName(),
		To:        to,
		Amount:    toOriginBalance(para.GetPara().GetAmount()),
		Sender:    para.GetPara().GetSender(),
		Receiver:  para.GetPara().GetReceiver(),
		Message:   message,
	}, toStringPoint(para.GetPrkStr()))
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (l *LedgerAPI) GenerateReceiveBlock(ctx context.Context, para *pb.GenerateReceiveBlockReq) (*pbtypes.StateBlock, error) {
	blk, err := toOriginStateBlock(para.GetBlock())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.GenerateReceiveBlock(blk, toStringPoint(para.GetPrkStr()))
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (l *LedgerAPI) GenerateReceiveBlockByHash(ctx context.Context, para *pb.GenerateReceiveBlockByHashReq) (*pbtypes.StateBlock, error) {
	hash, err := types.NewHash(para.GetHash())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.GenerateReceiveBlockByHash(hash, toStringPoint(para.GetPrkStr()))
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (l *LedgerAPI) GenerateChangeBlock(ctx context.Context, para *pb.GenerateChangeBlockReq) (*pbtypes.StateBlock, error) {
	acc, err := types.HexToAddress(para.GetAccount())
	if err != nil {
		return nil, err
	}
	rep, err := types.HexToAddress(para.GetRepresentative())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.GenerateChangeBlock(acc, rep, toStringPoint(para.GetPrkStr()))
	if err != nil {
		return nil, err
	}
	return toStateBlock(r), nil
}

func (l *LedgerAPI) Process(ctx context.Context, block *pbtypes.StateBlock) (*pbtypes.Hash, error) {
	blk, err := toOriginStateBlock(block)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.Process(blk)
	if err != nil {
		return nil, err
	}
	return &pbtypes.Hash{
		Hash: r.String(),
	}, nil
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

// StateBlock

func toStateBlock(blk *types.StateBlock) *pbtypes.StateBlock {
	return &pbtypes.StateBlock{
		Type:           blk.GetType().String(),
		Token:          blk.GetToken().String(),
		Address:        blk.GetAddress().String(),
		Balance:        blk.GetBalance().Int64(),
		Vote:           blk.GetVote().Int64(),
		Network:        blk.GetNetwork().Int64(),
		Storage:        blk.GetStorage().Int64(),
		Oracle:         blk.GetOracle().Int64(),
		Previous:       blk.GetPrevious().String(),
		Link:           blk.GetLink().String(),
		Sender:         blk.GetSender(),
		Receiver:       blk.GetReceiver(),
		Message:        blk.GetMessage().String(),
		Data:           blk.GetData(),
		PoVHeight:      blk.PoVHeight,
		Timestamp:      blk.GetTimestamp(),
		Extra:          blk.GetExtra().String(),
		Representative: blk.GetRepresentative().String(),
		PrivateFrom:    blk.PrivateFrom,
		PrivateFor:     blk.PrivateFor,
		PrivateGroupID: blk.PrivateGroupID,
		Work:           uint64(blk.GetWork()),
		Signature:      blk.GetSignature().String(),
		Flag:           blk.Flag,
		PrivateRecvRsp: blk.PrivateRecvRsp,
		PrivatePayload: blk.PrivatePayload,
	}
}

func toStateBlocks(blocks []*types.StateBlock) *pbtypes.StateBlocks {
	blk := make([]*pbtypes.StateBlock, 0)
	for _, b := range blocks {
		blk = append(blk, toStateBlock(b))
	}
	return &pbtypes.StateBlocks{StateBlocks: blk}
}

func toOriginStateBlock(blk *pbtypes.StateBlock) (*types.StateBlock, error) {
	token, err := types.NewHash(blk.GetToken())
	if err != nil {
		return nil, err
	}
	addr, err := types.HexToAddress(blk.GetAddress())
	if err != nil {
		return nil, err
	}
	pre, err := types.NewHash(blk.GetPrevious())
	if err != nil {
		return nil, err
	}
	link, err := types.NewHash(blk.GetLink())
	if err != nil {
		return nil, err
	}
	message, err := types.NewHash(blk.GetMessage())
	if err != nil {
		return nil, err
	}
	extra, err := types.NewHash(blk.GetExtra())
	if err != nil {
		return nil, err
	}
	rep, err := types.HexToAddress(blk.GetRepresentative())
	if err != nil {
		return nil, err
	}
	sign, err := types.NewSignature(blk.GetSignature())
	if err != nil {
		return nil, err
	}
	return &types.StateBlock{
		Type:           types.BlockTypeFromStr(blk.GetType()),
		Token:          token,
		Address:        addr,
		Balance:        types.Balance{Int: big.NewInt(blk.GetBalance())},
		Vote:           types.Balance{Int: big.NewInt(blk.GetVote())},
		Network:        types.Balance{Int: big.NewInt(blk.GetNetwork())},
		Storage:        types.Balance{Int: big.NewInt(blk.GetStorage())},
		Oracle:         types.Balance{Int: big.NewInt(blk.GetOracle())},
		Previous:       pre,
		Link:           link,
		Sender:         blk.GetSender(),
		Receiver:       blk.GetReceiver(),
		Message:        message,
		Data:           blk.GetData(),
		PoVHeight:      blk.GetPoVHeight(),
		Timestamp:      blk.GetTimestamp(),
		Extra:          extra,
		Representative: rep,
		PrivateFrom:    blk.GetPrivateFrom(),
		PrivateFor:     blk.GetPrivateFor(),
		PrivateGroupID: blk.GetPrivateGroupID(),
		Work:           types.Work(blk.GetWork()),
		Signature:      sign,
		Flag:           blk.GetFlag(),
		PrivateRecvRsp: blk.GetPrivateRecvRsp(),
		PrivatePayload: blk.GetPrivatePayload(),
	}, nil
}

func toOriginStateBlocks(blks []*pbtypes.StateBlock) ([]*types.StateBlock, error) {
	blocks := make([]*types.StateBlock, 0)
	for _, b := range blks {
		bt, err := toOriginStateBlock(b)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, bt)
	}
	return blocks, nil
}

func toAPIBlock(blk *api.APIBlock) *pb.APIBlock {
	return &pb.APIBlock{
		Type:             blk.GetType().String(),
		Token:            blk.GetToken().String(),
		Address:          blk.GetAddress().String(),
		Balance:          blk.GetBalance().Int64(),
		Vote:             blk.GetVote().Int64(),
		Network:          blk.GetNetwork().Int64(),
		Storage:          blk.GetStorage().Int64(),
		Oracle:           blk.GetOracle().Int64(),
		Previous:         blk.GetPrevious().String(),
		Link:             blk.GetLink().String(),
		Sender:           blk.GetSender(),
		Receiver:         blk.GetReceiver(),
		Message:          blk.GetMessage().String(),
		Data:             blk.GetData(),
		PoVHeight:        blk.PoVHeight,
		Timestamp:        blk.GetTimestamp(),
		Extra:            blk.GetExtra().String(),
		Representative:   blk.GetRepresentative().String(),
		PrivateFrom:      blk.PrivateFrom,
		PrivateFor:       blk.PrivateFor,
		PrivateGroupID:   blk.PrivateGroupID,
		Work:             uint64(blk.GetWork()),
		Signature:        blk.GetSignature().String(),
		Flag:             blk.Flag,
		PrivateRecvRsp:   blk.PrivateRecvRsp,
		PrivatePayload:   blk.PrivatePayload,
		TokenName:        blk.TokenName,
		Amount:           blk.Amount.Int64(),
		Hash:             blk.Hash.String(),
		PovConfirmHeight: blk.PovConfirmHeight,
		PovConfirmCount:  blk.PovConfirmCount,
	}
}

func toAPIBlocks(blks []*api.APIBlock) *pb.APIBlocks {
	blk := make([]*pb.APIBlock, 0)
	for _, b := range blks {
		blk = append(blk, toAPIBlock(b))
	}
	return &pb.APIBlocks{Blocks: blk}
}

// Address

func toAddress(addr types.Address) *pbtypes.Address {
	return &pbtypes.Address{
		Address: addr.String(),
	}
}

func toOriginAddress(addr *pbtypes.Address) (types.Address, error) {
	address, err := types.HexToAddress(addr.GetAddress())
	if err != nil {
		return types.ZeroAddress, nil
	}
	return address, nil
}

func toAddresses(addrs []types.Address) *pbtypes.Addresses {
	as := make([]string, 0)
	for _, addr := range addrs {
		as = append(as, addr.String())
	}
	return &pbtypes.Addresses{Addresses: as}
}

func toOriginAddresses(addrs *pbtypes.Addresses) ([]types.Address, error) {
	as := make([]types.Address, 0)
	for _, addr := range addrs.GetAddresses() {
		a, err := types.HexToAddress(addr)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, nil
}

func toAPIAccount(acc *api.APIAccount) *pb.APIAccount {
	tms := make([]*pb.APITokenMeta, 0)
	for _, tm := range acc.Tokens {
		t := &pb.APITokenMeta{
			Type:           tm.Type.String(),
			Header:         tm.Header.String(),
			Representative: tm.Representative.String(),
			OpenBlock:      tm.OpenBlock.String(),
			Balance:        tm.Balance.Int64(),
			BelongTo:       tm.BelongTo.String(),
			Modified:       tm.Modified,
			BlockCount:     tm.BlockCount,
			TokenName:      tm.TokenName,
			Pending:        tm.Pending.Int64(),
		}
		tms = append(tms, t)
	}
	r := &pb.APIAccount{
		Address: acc.Address.String(),
		Tokens:  tms,
	}
	if acc.CoinBalance != nil {
		r.CoinBalance = acc.CoinBalance.Int64()
	}
	if acc.CoinVote != nil {
		r.CoinVote = acc.CoinVote.Int64()
	}
	if acc.CoinNetwork != nil {
		r.CoinNetwork = acc.CoinNetwork.Int64()
	}
	if acc.CoinStorage != nil {
		r.CoinStorage = acc.CoinStorage.Int64()
	}
	if acc.CoinOracle != nil {
		r.CoinOracle = acc.CoinOracle.Int64()
	}
	if acc.Representative != nil {
		r.Representative = acc.Representative.String()
	}
	return r
}

// Hash

func toHash(hash types.Hash) *pbtypes.Hash {
	return &pbtypes.Hash{Hash: hash.String()}
}

func toOriginHash(hash *pbtypes.Hash) (types.Hash, error) {
	h, err := types.NewHash(hash.GetHash())
	if err != nil {
		return types.ZeroHash, err
	}
	return h, nil
}

func toHashes(hashes []types.Hash) *pbtypes.Hashes {
	hs := make([]string, 0)
	for _, h := range hashes {
		hs = append(hs, h.String())
	}
	return &pbtypes.Hashes{Hashes: hs}
}

func toOriginHashes(hashes *pbtypes.Hashes) ([]types.Hash, error) {
	hs := make([]types.Hash, 0)
	for _, h := range hashes.GetHashes() {
		h, err := types.NewHash(h)
		if err != nil {
			return nil, err
		}
		hs = append(hs, h)
	}
	return hs, nil
}

// Pending

func toAPIPendings(pendings []*api.APIPending) *pb.APIPendings {
	ps := make([]*pb.APIPending, 0)
	for _, pending := range pendings {
		pt := &pb.APIPending{
			Address:   pending.Address.String(),
			Hash:      pending.Hash.String(),
			Source:    pending.Source.String(),
			Amount:    pending.Amount.Int64(),
			Type:      pending.Type.String(),
			TokenName: pending.TokenName,
			Timestamp: pending.Timestamp,
			BlockType: pending.BlockType.String(),
		}
		ps = append(ps, pt)
	}
	return &pb.APIPendings{Pendings: ps}
}

// Representative

func toAPIRepresentatives(reps []*api.APIRepresentative) *pb.APIRepresentatives {
	rs := make([]*pb.APIRepresentative, 0)
	for _, r := range reps {
		rt := &pb.APIRepresentative{
			Address: r.Address.String(),
			Balance: r.Balance.Int64(),
			Vote:    r.Vote.Int64(),
			Network: r.Network.Int64(),
			Storage: r.Storage.Int64(),
			Oracle:  r.Oracle.Int64(),
			Total:   r.Total.Int64(),
		}
		rs = append(rs, rt)
	}
	return &pb.APIRepresentatives{Representatives: rs}
}

// Tokens

func toTokenInfo(token types.TokenInfo) *pbtypes.TokenInfo {
	return &pbtypes.TokenInfo{
		TokenId:       token.TokenId.String(),
		TokenName:     token.TokenName,
		TokenSymbol:   token.TokenSymbol,
		TotalSupply:   token.TotalSupply.Int64(),
		Decimals:      int32(token.Decimals),
		Owner:         token.Owner.String(),
		PledgeAmount:  token.PledgeAmount.Int64(),
		WithdrawTime:  token.WithdrawTime,
		PledgeAddress: token.PledgeAddress.String(),
		NEP5TxId:      token.NEP5TxId,
	}
}

func toTokenInfos(tokens []*types.TokenInfo) *pbtypes.TokenInfos {
	ts := make([]*pbtypes.TokenInfo, 0)
	for _, token := range tokens {
		ts = append(ts, toTokenInfo(*token))
	}
	return &pbtypes.TokenInfos{TokenInfos: ts}
}

// balance

func toOriginBalance(b int64) types.Balance {
	return types.Balance{Int: big.NewInt(b)}
}
