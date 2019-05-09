package api

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/ledger/relation"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type LedgerApi struct {
	ledger    *ledger.Ledger
	verifier  *process.LedgerVerifier
	vmContext *vmstore.VMContext
	eb        event.EventBus
	relation  *relation.Relation
	logger    *zap.SugaredLogger
}

type APIBlock struct {
	*types.StateBlock
	TokenName string        `json:"tokenName"`
	Amount    types.Balance `json:"amount"`
	Hash      types.Hash    `json:"hash"`
}

type APIAccount struct {
	Address        types.Address   `json:"account"`
	CoinBalance    *types.Balance  `json:"coinBalance,omitempty"`
	CoinVote       *types.Balance  `json:"vote,omitempty"`
	CoinNetwork    *types.Balance  `json:"network,omitempty"`
	CoinStorage    *types.Balance  `json:"storage,omitempty"`
	CoinOracle     *types.Balance  `json:"oracle,omitempty"`
	Representative *types.Address  `json:"representative,omitempty"`
	Tokens         []*APITokenMeta `json:"tokens"`
}

type APITokenMeta struct {
	*types.TokenMeta
	TokenName string        `json:"tokenName"`
	Pending   types.Balance `json:"pending"`
}

type APIPending struct {
	*types.PendingKey
	*types.PendingInfo
	TokenName string `json:"tokenName"`
	Timestamp int64  `json:"timestamp"`
}

type ApiTokenInfo struct {
	types.TokenInfo
}

func NewLedgerApi(l *ledger.Ledger, relation *relation.Relation, eb event.EventBus) *LedgerApi {
	return &LedgerApi{ledger: l, eb: eb, verifier: process.NewLedgerVerifier(l), vmContext: vmstore.NewVMContext(l),
		relation: relation, logger: log.NewLogger("api_ledger")}
}

func (b *APIBlock) fromStateBlock(block *types.StateBlock) *APIBlock {
	b.StateBlock = block
	b.Hash = block.GetHash()
	return b
}

func (l *LedgerApi) AccountBlocksCount(addr types.Address) (int64, error) {
	am, err := l.ledger.GetAccountMeta(addr)
	if err != nil {
		if err == ledger.ErrAccountNotFound {
			return 0, nil
		}
		return -1, err
	}
	var count int64
	for _, t := range am.Tokens {
		count = count + t.BlockCount
	}
	return count, nil
}

func checkOffset(count int, offset *int) (int, int, error) {
	if count < 1 {
		return 0, 0, errors.New("err count")
	}
	o := 0
	if offset != nil {
		o = *offset
		if o < 0 {
			return 0, 0, errors.New("err offset")
		}
	}
	return count, o, nil
}

func generateAPIBlock(ctx *vmstore.VMContext, block *types.StateBlock) (*APIBlock, error) {
	ab := new(APIBlock)
	ab.StateBlock = block
	ab.Hash = block.GetHash()
	if amount, err := ctx.CalculateAmount(block); err != nil {
		return nil, fmt.Errorf("block:%s, type:%s err:%s", ab.Hash.String(), ab.Type.String(), err)
	} else {
		ab.Amount = amount
	}
	token, err := abi.GetTokenById(ctx, block.GetToken())
	if err != nil {
		return nil, err
	}
	ab.TokenName = token.TokenName
	return ab, nil
}

// AccountHistoryTopn returns blocks of the account, blocks count of each token in the account up to n
func (l *LedgerApi) AccountHistoryTopn(address types.Address, count int, offset *int) ([]*APIBlock, error) {
	c, o, err := checkOffset(count, offset)
	if err != nil {
		return nil, err
	}
	hashes, err := l.relation.AccountBlocks(address, c, o)
	if err != nil {
		l.logger.Error(err)
		return nil, err
	}
	bs := make([]*APIBlock, 0)
	for _, h := range hashes {
		block, _ := l.ledger.GetStateBlock(h)
		b, err := generateAPIBlock(l.vmContext, block)
		if err != nil {
			return nil, err
		}
		bs = append(bs, b)
	}
	return bs, nil
}

func (l *LedgerApi) AccountInfo(address types.Address) (*APIAccount, error) {
	aa := new(APIAccount)
	am, err := l.ledger.GetAccountMeta(address)
	if err != nil {
		return nil, err
	}
	for _, t := range am.Tokens {
		if t.Type == common.ChainToken() {
			aa.CoinBalance = &t.Balance
			aa.Representative = &t.Representative
			aa.CoinVote = &am.CoinVote
			aa.CoinNetwork = &am.CoinNetwork
			aa.CoinOracle = &am.CoinOracle
			aa.CoinStorage = &am.CoinStorage
		}
		info, err := abi.GetTokenById(l.vmContext, t.Type)
		if err != nil {
			return nil, err
		}
		pendingKeys, err := l.ledger.TokenPending(address, t.Type)
		if err != nil {
			return nil, err
		}
		pendingAmount := types.ZeroBalance
		for _, key := range pendingKeys {
			pendinginfo, err := l.ledger.GetPending(*key)
			if err != nil {
				return nil, err
			}
			pendingAmount = pendingAmount.Add(pendinginfo.Amount)
		}
		tm := APITokenMeta{
			TokenMeta: t,
			TokenName: info.TokenName,
			Pending:   pendingAmount,
		}
		aa.Tokens = append(aa.Tokens, &tm)

	}
	aa.Address = address
	return aa, nil
}

func (l *LedgerApi) AccountRepresentative(addr types.Address) (types.Address, error) {
	am, err := l.ledger.GetAccountMeta(addr)
	if err != nil {
		return types.ZeroAddress, err
	}
	for _, t := range am.Tokens {
		if t.Type == common.ChainToken() {
			return t.Representative, nil
		}
	}
	return types.ZeroAddress, errors.New("account has no representative")
}

func (l *LedgerApi) AccountVotingWeight(addr types.Address) (types.Balance, error) {
	b, err := l.ledger.GetRepresentation(addr)
	if err != nil {
		return types.ZeroBalance, nil
	}
	return b.Total, err
}

func (l *LedgerApi) AccountsBalance(addresses []types.Address) (map[types.Address]map[string]map[string]types.Balance, error) {
	as := make(map[types.Address]map[string]map[string]types.Balance)

	for _, addr := range addresses {
		ac, err := l.ledger.GetAccountMeta(addr)
		if err != nil {
			if err == ledger.ErrAccountNotFound {
				continue
			}
			return nil, err
		}
		ts := make(map[string]map[string]types.Balance)
		for _, t := range ac.Tokens {
			info, err := abi.GetTokenById(l.vmContext, t.Type)
			if err != nil {
				return nil, err
			}
			b := make(map[string]types.Balance)
			pendings, err := l.ledger.TokenPendingInfo(addr, t.Type)
			if err != nil {
				return nil, err
			}
			amount := types.ZeroBalance
			for _, pending := range pendings {
				amount = amount.Add(pending.Amount)
			}
			b["balance"] = t.Balance
			b["pending"] = amount
			if info.TokenId == common.ChainToken() {
				if !ac.CoinVote.IsZero() || !ac.CoinStorage.IsZero() || !ac.CoinOracle.IsZero() || !ac.CoinNetwork.IsZero() {
					b["vote"] = ac.CoinVote
					b["network"] = ac.CoinNetwork
					b["oracle"] = ac.CoinOracle
					b["storage"] = ac.CoinStorage
				}
			}
			ts[info.TokenName] = b
		}
		as[addr] = ts
	}
	return as, nil
}

func (l *LedgerApi) AccountsFrontiers(addresses []types.Address) (map[types.Address]map[string]types.Hash, error) {
	r := make(map[types.Address]map[string]types.Hash)
	for _, addr := range addresses {
		ac, err := l.ledger.GetAccountMeta(addr)
		if err != nil {
			if err == ledger.ErrAccountNotFound {
				continue
			}
			return nil, err
		}
		t := make(map[string]types.Hash)
		for _, token := range ac.Tokens {
			info, err := abi.GetTokenById(l.vmContext, token.Type)
			if err != nil {
				return nil, err
			}
			t[info.TokenName] = token.Header
		}
		r[addr] = t
	}
	return r, nil
}

func (l *LedgerApi) AccountsPending(addresses []types.Address, n int) (map[types.Address][]*APIPending, error) {
	apMap := make(map[types.Address][]*APIPending)
	for _, addr := range addresses {
		ps := make([]*APIPending, 0)
		err := l.ledger.SearchPending(addr, func(key *types.PendingKey, info *types.PendingInfo) error {
			token, err := abi.GetTokenById(l.vmContext, info.Type)
			if err != nil {
				return err
			}
			tokenName := token.TokenName
			blk, err := l.ledger.GetStateBlock(key.Hash)
			if err != nil {
				return err
			}
			ap := APIPending{
				PendingKey:  key,
				PendingInfo: info,
				TokenName:   tokenName,
				Timestamp:   blk.Timestamp,
			}
			ps = append(ps, &ap)
			return nil
		})

		if err != nil {
			fmt.Println(err)
		}
		if len(ps) > 0 {
			pt := ps
			if n > -1 && len(ps) > n {
				pt = ps[:n]
			}
			apMap[addr] = pt
		}
	}
	return apMap, nil
}

func (l *LedgerApi) AccountsCount() (uint64, error) {
	return l.ledger.CountAccountMetas()
}

func (l *LedgerApi) Accounts(count int, offset *int) ([]*types.Address, error) {
	c, o, err := checkOffset(count, offset)
	if err != nil {
		return nil, err
	}
	as := make([]*types.Address, 0)
	index := 0
	err = l.ledger.GetAccountMetas(func(am *types.AccountMeta) error {
		if index >= o && index < o+c {
			as = append(as, &am.Address)
		}
		index = index + 1
		return nil
	})
	if err != nil {
		return nil, err
	}
	return as, nil
}

func (l *LedgerApi) BlockAccount(hash types.Hash) (types.Address, error) {
	sb, err := l.ledger.GetStateBlock(hash)
	if err != nil {
		return types.ZeroAddress, err
	}
	return sb.GetAddress(), nil
}

func (l *LedgerApi) BlockHash(block types.StateBlock) types.Hash {
	return block.GetHash()
}

// BlocksCount returns the number of blocks (include smartcontrant block) in the ctx and unchecked synchronizing blocks
func (l *LedgerApi) BlocksCount() (map[string]uint64, error) {
	sbCount, err := l.relation.BlocksCount()
	if err != nil {
		return nil, err
	}
	scbCount, err := l.ledger.CountSmartContractBlocks()
	if err != nil {
		return nil, err
	}
	unCount, err := l.ledger.CountUncheckedBlocks()
	if err != nil {
		return nil, err
	}
	c := make(map[string]uint64)
	c["count"] = sbCount + scbCount
	c["unchecked"] = unCount
	return c, nil
}

func (l *LedgerApi) BlocksCount2() (map[string]uint64, error) {
	sbCount, err := l.ledger.CountStateBlocks()
	if err != nil {
		return nil, err
	}
	scbCount, err := l.ledger.CountSmartContractBlocks()
	if err != nil {
		return nil, err
	}
	unCount, err := l.ledger.CountUncheckedBlocks()
	if err != nil {
		return nil, err
	}
	c := make(map[string]uint64)
	c["count"] = sbCount + scbCount
	c["unchecked"] = unCount
	return c, nil
}

//BlocksCountByType reports the number of blocks in the ctx by type (send, receive, open, change)
func (l *LedgerApi) BlocksCountByType() (map[string]uint64, error) {
	c := make(map[string]uint64)
	c[types.Open.String()] = 0
	c[types.Send.String()] = 0
	c[types.Receive.String()] = 0
	c[types.Change.String()] = 0
	c[types.ContractReward.String()] = 0
	c[types.ContractSend.String()] = 0
	c[types.ContractRefund.String()] = 0
	ts, err := l.relation.BlocksCountByType()
	if err != nil {
		l.logger.Error(err)
	}
	for k, v := range ts {
		c[k] = v
	}
	return c, nil
}

func (l *LedgerApi) BlocksInfo(hash []types.Hash) ([]*APIBlock, error) {
	bs := make([]*APIBlock, 0)
	for _, h := range hash {
		block, err := l.ledger.GetStateBlock(h)
		if err != nil {
			if err == ledger.ErrBlockNotFound {
				continue
			}
			return nil, fmt.Errorf("%s, %s", h, err)
		}
		b, err := generateAPIBlock(l.vmContext, block)
		if err != nil {
			return nil, err
		}
		bs = append(bs, b)
	}
	return bs, nil
}

func (l *LedgerApi) Blocks(count int, offset *int) ([]*APIBlock, error) {
	c, o, err := checkOffset(count, offset)
	if err != nil {
		return nil, err
	}
	hashes, err := l.relation.Blocks(c, o)
	if err != nil {
		return nil, err
	}
	bs := make([]*APIBlock, 0)
	for _, h := range hashes {
		block, err := l.ledger.GetStateBlock(h)
		if err != nil && err != ledger.ErrBlockNotFound {
			return nil, err
		}
		if block != nil {
			b, err := generateAPIBlock(l.vmContext, block)
			if err != nil {
				return nil, err
			}
			bs = append(bs, b)
		}
	}
	return bs, nil
}

// Chain returns a consecutive list of block hashes in the account chain starting at block up to count
func (l *LedgerApi) Chain(hash types.Hash, n int) ([]types.Hash, error) {
	if n < -1 {
		return nil, errors.New("wrong count number")
	}
	r := make([]types.Hash, 0)
	count := 0
	for (n != -1 && count < n) || n == -1 {
		blk, err := l.ledger.GetStateBlock(hash)
		if err != nil {
			return nil, err
		}
		r = append(r, blk.GetHash())
		hash = blk.GetPrevious()
		if hash.IsZero() {
			break
		}
		count = count + 1
	}
	return r, nil
}

// Delegators returns a list of pairs of delegator names given account a representative and its balance
func (l *LedgerApi) Delegators(hash types.Address) ([]*APIAccountBalance, error) {
	abs := make([]*APIAccountBalance, 0)

	err := l.ledger.GetAccountMetas(func(am *types.AccountMeta) error {
		t := am.Token(common.ChainToken())
		if t != nil {
			if t.Representative == hash {
				ab := &APIAccountBalance{am.Address, am.VoteWeight()}
				abs = append(abs, ab)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return abs, nil
}

// DelegatorsCount gets number of delegators for a specific representative account
func (l *LedgerApi) DelegatorsCount(hash types.Address) (int64, error) {
	var count int64
	err := l.ledger.GetAccountMetas(func(am *types.AccountMeta) error {
		t := am.Token(common.ChainToken())
		if t != nil {
			if t.Representative == hash {
				count = count + 1
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

type APISendBlockPara struct {
	From      types.Address `json:"from"`
	TokenName string        `json:"tokenName"`
	To        types.Address `json:"to"`
	Amount    types.Balance `json:"amount"`
	Sender    string        `json:"sender"`
	Receiver  string        `json:"receiver"`
	Message   types.Hash    `json:"message"`
}

func (l *LedgerApi) GenerateSendBlock(para *APISendBlockPara, prkStr string) (*types.StateBlock, error) {
	if para.Amount.Int == nil || para.From.IsZero() || para.To.IsZero() || para.TokenName == "" {
		return nil, errors.New("invalid transaction parameter")
	}
	prk, err := hex.DecodeString(prkStr)
	if err != nil {
		return nil, err
	}
	info, err := abi.GetTokenByName(l.vmContext, para.TokenName)
	if err != nil {
		return nil, err
	}

	s, err := phoneNumberSeri(para.Sender)
	if err != nil {
		return nil, errors.New("error sender")
	}
	r, err := phoneNumberSeri(para.Receiver)
	if err != nil {
		return nil, errors.New("error receiver")
	}
	sb := types.StateBlock{
		Address:  para.From,
		Token:    info.TokenId,
		Link:     para.To.ToHash(),
		Sender:   s,
		Receiver: r,
		Message:  para.Message,
	}
	block, err := l.ledger.GenerateSendBlock(&sb, para.Amount, prk)
	if err != nil {
		return nil, err
	}
	l.logger.Debug(block)
	return block, nil
}

func (l *LedgerApi) GenerateReceiveBlock(sendBlock *types.StateBlock, prkStr string) (*types.StateBlock, error) {
	prk, err := hex.DecodeString(prkStr)
	if err != nil {
		return nil, err
	}

	block, err := l.ledger.GenerateReceiveBlock(sendBlock, prk)
	if err != nil {
		return nil, err
	}
	l.logger.Debug(block)
	return block, nil
}

func (l *LedgerApi) GenerateChangeBlock(account types.Address, representative types.Address, prkStr string) (*types.StateBlock, error) {
	prk, err := hex.DecodeString(prkStr)
	if err != nil {
		return nil, err
	}

	block, err := l.ledger.GenerateChangeBlock(account, representative, prk)
	if err != nil {
		return nil, err
	}

	l.logger.Debug(block)
	return block, nil
}

func (l *LedgerApi) Pendings() ([]*APIPending, error) {
	aps := make([]*APIPending, 0)
	err := l.ledger.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		token, err := abi.GetTokenById(l.vmContext, pendingInfo.Type)
		if err != nil {
			return err
		}
		tokenName := token.TokenName
		blk, err := l.ledger.GetStateBlock(pendingKey.Hash)
		if err != nil {
			return err
		}
		ap := APIPending{
			PendingKey:  pendingKey,
			PendingInfo: pendingInfo,
			TokenName:   tokenName,
			Timestamp:   blk.Timestamp,
		}
		aps = append(aps, &ap)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return aps, nil
}

func (l *LedgerApi) Process(block *types.StateBlock) (types.Hash, error) {
	flag, err := l.verifier.Process(block)
	if err != nil {
		l.logger.Error(err)
		return types.ZeroHash, err
	}

	l.logger.Debug("process result, ", flag)
	switch flag {
	case process.Progress:
		l.logger.Debug("broadcast block")
		//TODO: refine
		//l.dpos.GetP2PService().Broadcast(p2p.PublishReq, block)
		l.eb.Publish(string(common.EventBroadcast), p2p.PublishReq, block)
		return block.GetHash(), nil
	case process.BadWork:
		return types.ZeroHash, errors.New("bad work")
	case process.BadSignature:
		return types.ZeroHash, errors.New("bad signature")
	case process.Old:
		l.logger.Info("old block")
		//return block.GetHash(), nil
		return types.ZeroHash, errors.New("old block")
	case process.Fork:
		return types.ZeroHash, errors.New("fork")
	case process.GapSource:
		return types.ZeroHash, errors.New("gap source block")
	case process.GapPrevious:
		return types.ZeroHash, errors.New("gap previous block")
	case process.BalanceMismatch:
		return types.ZeroHash, errors.New("balance mismatch")
	case process.UnReceivable:
		return types.ZeroHash, errors.New("unReceivable")
	case process.GapSmartContract:
		return types.ZeroHash, errors.New("gap SmartContract")
	case process.InvalidData:
		return types.ZeroHash, errors.New("invalid data")
	default:
		return types.ZeroHash, errors.New("error processing block")
	}
}

func (l *LedgerApi) Performance() ([]*types.PerformanceTime, error) {
	pts := make([]*types.PerformanceTime, 0)
	err := l.ledger.PerformanceTimes(func(p *types.PerformanceTime) {
		pts = append(pts, p)
	})
	if err != nil {
		return nil, err
	}
	return pts, nil
}

type APIAccountBalance struct {
	Address types.Address `json:"address"`
	Balance types.Balance `json:"balance"`
}

type APIAccountBalances []APIAccountBalance

func (r APIAccountBalances) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r APIAccountBalances) Len() int {
	return len(r)
}

func (r APIAccountBalances) Less(i, j int) bool {
	if r[i].Balance.Compare(r[j].Balance) == types.BalanceCompSmaller {
		return false
	}
	return true
}

//Representatives returns a list of pairs of representative and its voting weight
func (l *LedgerApi) Representatives(sorting *bool) (*APIAccountBalances, error) {
	rs := make(APIAccountBalances, 0)
	err := l.ledger.GetRepresentations(func(address types.Address, benefit *types.Benefit) error {
		r := APIAccountBalance{address, benefit.Total}
		rs = append(rs, r)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if sorting != nil && *sorting {
		sort.Sort(rs)
	}
	return &rs, nil
}

func (l *LedgerApi) Tokens() ([]*types.TokenInfo, error) {
	return abi.ListTokens(l.vmContext)
}

// BlocksCount returns the number of blocks (not include smartcontrant block) in the ctx and unchecked synchronizing blocks
func (l *LedgerApi) TransactionsCount() (map[string]uint64, error) {
	sbCount, err := l.relation.BlocksCount()
	if err != nil {
		return nil, err
	}
	unCount, err := l.ledger.CountUncheckedBlocks()
	if err != nil {
		return nil, err
	}
	c := make(map[string]uint64)
	c["count"] = sbCount
	c["unchecked"] = unCount
	return c, nil
}

func (l *LedgerApi) TokenInfoById(tokenId types.Hash) (*ApiTokenInfo, error) {
	token, err := abi.GetTokenById(l.vmContext, tokenId)
	if err != nil {
		return nil, err
	}
	return &ApiTokenInfo{*token}, nil
}

func (l *LedgerApi) TokenInfoByName(tokenName string) (*ApiTokenInfo, error) {
	token, err := abi.GetTokenByName(l.vmContext, tokenName)
	if err != nil {
		return nil, err
	}
	return &ApiTokenInfo{*token}, nil
}
