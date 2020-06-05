package apis

import (
	"context"
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

func (l *LedgerAPI) AccountBlocksCount(ctx context.Context, addr *pbtypes.Address) (*pb.Int64, error) {
	address, err := toOriginAddressByValue(addr.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.AccountBlocksCount(address)
	if err != nil {
		return nil, err
	}
	return &pb.Int64{Value: r}, nil
}

func (l *LedgerAPI) AccountHistoryTopn(ctx context.Context, param *pb.AccountHistoryTopnReq) (*pb.APIBlocks, error) {
	address, err := toOriginAddressByValue(param.GetAddress())
	if err != nil {
		return nil, err
	}
	count, offset := toOffsetByValue(param.GetCount(), param.GetOffset())
	r, err := l.ledger.AccountHistoryTopn(address, count, offset)
	if err != nil {
		return nil, err
	}
	return toAPIBlocks(r), nil
}

func (l *LedgerAPI) AccountInfo(ctx context.Context, addr *pbtypes.Address) (*pb.APIAccount, error) {
	address, err := toOriginAddressByValue(addr.GetAddress())
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
	address, err := toOriginAddressByValue(addr.GetAddress())
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
	address, err := toOriginAddressByValue(addr.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.AccountRepresentative(address)
	if err != nil {
		return nil, err
	}
	return toAddress(r), nil
}

func (l *LedgerAPI) AccountVotingWeight(ctx context.Context, addr *pbtypes.Address) (*pbtypes.Balance, error) {
	address, err := toOriginAddressByValue(addr.GetAddress())
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.AccountVotingWeight(address)
	if err != nil {
		return nil, err
	}
	return toBalance(r), nil
}

func (l *LedgerAPI) AccountsCount(context.Context, *empty.Empty) (*pb.UInt64, error) {
	r, err := l.ledger.AccountsCount()
	if err != nil {
		return nil, err
	}
	return &pb.UInt64{
		Value: r,
	}, nil
}

func (l *LedgerAPI) Accounts(ctx context.Context, para *pb.Offset) (*pbtypes.Addresses, error) {
	count, offset := toOffsetByProto(para)
	r, err := l.ledger.Accounts(count, offset)
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
				Balance: toBalanceValue(ba.Balance),
				Pending: toBalanceValue(ba.Pending),
			}
			if ba.Vote != nil {
				pt.Vote = toBalanceValue(*ba.Vote)
			}
			if ba.Network != nil {
				pt.Network = toBalanceValue(*ba.Network)
			}
			if ba.Storage != nil {
				pt.Storage = toBalanceValue(*ba.Storage)
			}
			if ba.Oracle != nil {
				pt.Oracle = toBalanceValue(*ba.Oracle)
			}
			fi[tokenName] = pt
		}
		aff := &pb.AccountsBalanceRspBalances{Balances: fi}
		result[toAddressValue(addr)] = aff
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
			fi[tokenName] = toHashValue(header)
		}
		aff := &pb.AccountsFrontiersRspFrontier{Frontier: fi}
		result[toAddressValue(addr)] = aff
	}
	return &pb.AccountsFrontiersRsp{AccountsFrontiers: result}, nil
}

func (l *LedgerAPI) AccountsPending(ctx context.Context, ap *pb.AccountsPendingReq) (*pb.AccountsPendingRsp, error) {
	addrs, err := toOriginAddressesByValues(ap.GetAddresses())
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
		aps.AccountsPendings[toAddressValue(addr)] = toAPIPendings(pendings)
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

func (l *LedgerAPI) BlockConfirmedStatus(ctx context.Context, hash *pbtypes.Hash) (*pb.Boolean, error) {
	h, err := toOriginHash(hash)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.BlockConfirmedStatus(h)
	if err != nil {
		return nil, err
	}
	return &pb.Boolean{
		Value: r,
	}, nil
}

func (l *LedgerAPI) BlockHash(ctx context.Context, block *pbtypes.StateBlock) (*pbtypes.Hash, error) {
	blk, err := toOriginStateBlock(block)
	if err != nil {
		return nil, err
	}
	r := l.ledger.BlockHash(*blk)
	return toHash(r), nil
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

func (l *LedgerAPI) Blocks(ctx context.Context, param *pb.Offset) (*pb.APIBlocks, error) {
	count, offset := toOffsetByProto(param)
	r, err := l.ledger.Blocks(count, offset)
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
	hash, err := toOriginHashByValue(para.GetHash())
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
			Address: toAddressValue(b.Address),
			Balance: toBalanceValue(b.Balance),
		}
		result = append(result, bt)
	}
	return &pb.APIAccountBalances{Balances: result}, nil
}

func (l *LedgerAPI) DelegatorsCount(ctx context.Context, addr *pbtypes.Address) (*pb.Int64, error) {
	address, err := toOriginAddress(addr)
	if err != nil {
		return nil, err
	}
	r, err := l.ledger.DelegatorsCount(address)
	if err != nil {
		return nil, err
	}
	return &pb.Int64{
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

func (l *LedgerAPI) Representatives(ctx context.Context, b *pb.Boolean) (*pb.APIRepresentatives, error) {
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

func (l *LedgerAPI) TokenInfoByName(ctx context.Context, name *pb.String) (*pbtypes.TokenInfo, error) {
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

func (l *LedgerAPI) IsGenesisBlock(ctx context.Context, block *pbtypes.StateBlock) (*pb.Boolean, error) {
	blk, err := toOriginStateBlock(block)
	if err != nil {
		return nil, err
	}
	r := l.ledger.IsGenesisBlock(blk)
	return &pb.Boolean{
		Value: r,
	}, nil
}

func (l *LedgerAPI) IsGenesisToken(ctx context.Context, hash *pbtypes.Hash) (*pb.Boolean, error) {
	h, err := toOriginHash(hash)
	if err != nil {
		return nil, err
	}
	r := l.ledger.IsGenesisToken(h)
	return &pb.Boolean{
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
	from, err := toOriginAddressByValue(para.GetParam().GetFrom())
	if err != nil {
		return nil, err
	}
	to, err := toOriginAddressByValue(para.GetParam().GetTo())
	if err != nil {
		return nil, err
	}
	message := types.ZeroHash
	if para.GetParam().GetMessage() != "" {
		message, err = toOriginHashByValue(para.GetParam().GetMessage())
		if err != nil {
			return nil, err
		}
	}
	r, err := l.ledger.GenerateSendBlock(&api.APISendBlockPara{
		From:      from,
		TokenName: para.GetParam().GetTokenName(),
		To:        to,
		Amount:    toOriginBalanceByValue(para.GetParam().GetAmount()),
		Sender:    para.GetParam().GetSender(),
		Receiver:  para.GetParam().GetReceiver(),
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
	hash, err := toOriginHashByValue(para.GetHash())
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
	acc, err := toOriginAddressByValue(para.GetAccount())
	if err != nil {
		return nil, err
	}
	rep, err := toOriginAddressByValue(para.GetRepresentative())
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
	return toHash(r), nil
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

func toAPIBlock(blk *api.APIBlock) *pb.APIBlock {
	return &pb.APIBlock{
		Type:             blk.GetType().String(),
		Token:            toHashValue(blk.GetToken()),
		Address:          toAddressValue(blk.GetAddress()),
		Balance:          toBalanceValue(blk.GetBalance()),
		Vote:             toBalanceValue(blk.GetVote()),
		Network:          toBalanceValue(blk.GetNetwork()),
		Storage:          toBalanceValue(blk.GetStorage()),
		Oracle:           toBalanceValue(blk.GetOracle()),
		Previous:         toHashValue(blk.GetPrevious()),
		Link:             toHashValue(blk.GetLink()),
		Sender:           blk.GetSender(),
		Receiver:         blk.GetReceiver(),
		Message:          toHashValue(blk.GetMessage()),
		Data:             blk.GetData(),
		PoVHeight:        blk.PoVHeight,
		Timestamp:        blk.GetTimestamp(),
		Extra:            toHashValue(blk.GetExtra()),
		Representative:   toAddressValue(blk.GetRepresentative()),
		PrivateFrom:      blk.PrivateFrom,
		PrivateFor:       blk.PrivateFor,
		PrivateGroupID:   blk.PrivateGroupID,
		Work:             toWorkValue(blk.GetWork()),
		Signature:        toSignatureValue(blk.GetSignature()),
		Flag:             blk.Flag,
		PrivateRecvRsp:   blk.PrivateRecvRsp,
		PrivatePayload:   blk.PrivatePayload,
		TokenName:        blk.TokenName,
		Amount:           toBalanceValue(blk.Amount),
		Hash:             toHashValue(blk.Hash),
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

func toAPIAccount(acc *api.APIAccount) *pb.APIAccount {
	tms := make([]*pb.APITokenMeta, 0)
	for _, tm := range acc.Tokens {
		t := &pb.APITokenMeta{
			Type:           toHashValue(tm.Type),
			Header:         toHashValue(tm.Header),
			Representative: toAddressValue(tm.Representative),
			OpenBlock:      toHashValue(tm.OpenBlock),
			Balance:        toBalanceValue(tm.Balance),
			BelongTo:       toAddressValue(tm.BelongTo),
			Modified:       tm.Modified,
			BlockCount:     tm.BlockCount,
			TokenName:      tm.TokenName,
			Pending:        toBalanceValue(tm.Pending),
		}
		tms = append(tms, t)
	}
	r := &pb.APIAccount{
		Address: toAddressValue(acc.Address),
		Tokens:  tms,
	}
	if acc.CoinBalance != nil {
		r.CoinBalance = toBalanceValue(*acc.CoinBalance)
	}
	if acc.CoinVote != nil {
		r.CoinVote = toBalanceValue(*acc.CoinVote)
	}
	if acc.CoinNetwork != nil {
		r.CoinNetwork = toBalanceValue(*acc.CoinNetwork)
	}
	if acc.CoinStorage != nil {
		r.CoinStorage = toBalanceValue(*acc.CoinStorage)
	}
	if acc.CoinOracle != nil {
		r.CoinOracle = toBalanceValue(*acc.CoinOracle)
	}
	if acc.Representative != nil {
		r.Representative = toAddressValue(*acc.Representative)
	}
	return r
}

// Pending

func toAPIPendings(pendings []*api.APIPending) *pb.APIPendings {
	ps := make([]*pb.APIPending, 0)
	for _, pending := range pendings {
		pt := &pb.APIPending{
			Address:   toAddressValue(pending.Address),
			Hash:      toHashValue(pending.Hash),
			Source:    toAddressValue(pending.Source),
			Amount:    pending.Amount.Int64(),
			Type:      toHashValue(pending.Type),
			TokenName: pending.TokenName,
			Timestamp: pending.Timestamp,
			BlockType: toBlockTypeValue(pending.BlockType),
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
			Address: toAddressValue(r.Address),
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
		TokenId:       toHashValue(token.TokenId),
		TokenName:     token.TokenName,
		TokenSymbol:   token.TokenSymbol,
		TotalSupply:   token.TotalSupply.Int64(),
		Decimals:      int32(token.Decimals),
		Owner:         toAddressValue(token.Owner),
		PledgeAmount:  token.PledgeAmount.Int64(),
		WithdrawTime:  token.WithdrawTime,
		PledgeAddress: toAddressValue(token.PledgeAddress),
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
