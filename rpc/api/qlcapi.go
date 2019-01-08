package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/test/mock"
)

var logger = log.NewLogger("rpcapi")

type QlcApi struct {
	ledger *ledger.Ledger
	dpos   *consensus.DposService
}

type TokenPending struct {
	PendingInfo *types.PendingInfo `json:"pendingInfo"`
	TokenName   string             `json:"tokenName"`
	Hash        types.Hash         `json:"hash"`
}

type APIBlock struct {
	Block     *types.StateBlock `json:"block"`
	SubType   string            `json:"subType"`
	TokenName string            `json:"tokenName"`
}

func NewQlcApi(l *ledger.Ledger, dpos *consensus.DposService) *QlcApi {
	return &QlcApi{ledger: l, dpos: dpos}
}

func (q *QlcApi) AccountsBalances(addresses []types.Address) (map[types.Address]map[types.Hash][]types.Balance, error) {
	logger.Info("addresses", addresses)
	r := make(map[types.Address]map[types.Hash][]types.Balance)
	for _, addr := range addresses {
		ac, err := q.ledger.GetAccountMeta(addr)
		if err != nil {
			return nil, err
		}
		t := make(map[types.Hash][]types.Balance)
		for _, token := range ac.Tokens {
			t[token.Type] = append(t[token.Type], token.Balance)
			pendings, err := q.ledger.TokenPendingInfo(addr, token.Type)
			if err != nil {
				return nil, err
			}
			var b types.Balance
			for _, pending := range pendings {
				b = b.Add(pending.Amount)
			}
			t[token.Type] = append(t[token.Type], b)
		}
		r[addr] = t
	}
	return r, nil
}

func (q *QlcApi) AccountsFrontiers(addresses []types.Address) (map[types.Address]map[types.Hash]types.Hash, error) {
	logger.Info("addresses", addresses)
	r := make(map[types.Address]map[types.Hash]types.Hash)
	for _, addr := range addresses {
		ac, err := q.ledger.GetAccountMeta(addr)
		if err != nil {
			return nil, err
		}
		t := make(map[types.Hash]types.Hash)
		for _, token := range ac.Tokens {
			t[token.Type] = token.Header
		}
		r[addr] = t
	}
	return r, nil
}

func (q *QlcApi) AccountsPending(addresses []types.Address, n int) (map[types.Address][]*TokenPending, error) {
	logger.Info("addresses", addresses)
	apMap := make(map[types.Address][]*TokenPending)
	for _, addr := range addresses {
		pendingkeys, err := q.ledger.Pending(addr)
		if err != nil {
			return nil, err
		}
		var tps []*TokenPending

		for _, pendingkey := range pendingkeys {
			pendinginfo, err := q.ledger.GetPending(*pendingkey)
			if err != nil {
				return nil, err
			}

			token, err := mock.GetTokenById(pendinginfo.Type)
			if err != nil {
				return nil, err
			}
			tokenname := token.TokenName
			tp := TokenPending{
				PendingInfo: pendinginfo,
				TokenName:   tokenname,
				Hash:        pendingkey.Hash,
			}
			tps = append(tps, &tp)
			if len(tps) > n {
				break
			}
		}
		if tps == nil {
			tps = make([]*TokenPending, 0)
		}
		apMap[addr] = tps
	}
	return apMap, nil
}

func (q *QlcApi) GetOnlineRepresentatives() []types.Address {
	return q.dpos.GetOnlineRepresentatives()
}

func (q *QlcApi) BlocksInfo(hash []types.Hash) ([]*APIBlock, error) {
	var bs []*APIBlock
	for _, h := range hash {
		b := new(APIBlock)
		block, err := q.ledger.GetStateBlock(h)
		if err != nil {
			return nil, err
		}
		b.Block = block
		//b.SubType, err = q.judgeBlockKind(block.GetHash())
		//if err != nil {
		//	return nil, err
		//}
		b.SubType = "state"
		logger.Info("getToken,", block.GetToken())
		token, err := mock.GetTokenById(block.GetToken())
		if err != nil {
			return nil, err
		}
		b.TokenName = token.TokenName
		bs = append(bs, b)
	}
	if bs == nil {
		bs = make([]*APIBlock, 0)
	}
	return bs, nil
}

func (q *QlcApi) Process(block *types.StateBlock) (ledger.ProcessResult, error) {
	return q.ledger.Process(block)
}

func (q *QlcApi) judgeBlockKind(hash types.Hash) (string, error) {
	blkType, err := q.ledger.JudgeBlockKind(hash)
	if err != nil {
		return "", err
	}
	switch blkType {
	case ledger.Open:
		return "open", nil
	case ledger.Receive:
		return "receive", nil
	case ledger.Send:
		return "send", nil
	case ledger.Change:
		return "change", nil
	default:
		return "unknow", nil
	}
}

func (q *QlcApi) AccountHistoryTopn(address types.Address, n int) ([]*APIBlock, error) {
	logger.Info(address)
	blocks, err := q.ledger.GetStateBlocks()
	if err != nil {
		return nil, err
	}

	logger.Info(n)
	var bs []*APIBlock
	for _, block := range blocks {
		if block.GetAddress() == address {
			b := new(APIBlock)
			logger.Info(b)
			b.SubType, err = q.judgeBlockKind(block.GetHash())
			logger.Info(b.SubType)
			if err != nil {
				logger.Info(err)
				return nil, err
			}
			logger.Info("getToken,", block.GetToken())
			token, err := mock.GetTokenById(block.GetToken())
			if err != nil {
				logger.Info(err)
				return nil, err
			}
			b.TokenName = token.TokenName
			b.Block = block
			bs = append(bs, b)
		}
		if len(bs) > n {
			break
		}
	}

	if bs == nil {
		bs = make([]*APIBlock, 0)
	}
	return bs, nil
}

func (q *QlcApi) AccountInfo(addr types.Address) (*types.AccountMeta, error) {
	am, err := q.ledger.GetAccountMeta(addr)
	if err != nil {
		return nil, err
	}
	return am, nil
}

func (q *QlcApi) ValidateAccount(addr string) bool {
	return types.IsValidHexAddress(addr)
}

func (q *QlcApi) Tokens() ([]*mock.TokenInfo, error) {
	var tis []*mock.TokenInfo
	scs := mock.GetSmartContracts()
	for _, sc := range scs {
		hash := sc.GetHash()
		ti, err := mock.GetTokenById(hash)
		if err != nil {
			return nil, err
		}
		tis = append(tis, &ti)
	}
	return tis, nil
}
