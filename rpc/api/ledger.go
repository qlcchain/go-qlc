package api

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"
	"sort"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"go.uber.org/zap"
)

type LedgerApi struct {
	ledger *ledger.Ledger
	dpos   *consensus.DposService
	logger *zap.SugaredLogger
}

type APIBlock struct {
	*types.StateBlock
	TokenName string        `json:"tokenName"`
	Amount    types.Balance `json:"amount"`
	Hash      types.Hash    `json:"hash"`
}

type APIAccount struct {
	Address        types.Address   `json:"account"`
	CoinBalance    types.Balance   `json:"coinBalance"`
	Representative types.Address   `json:"representative"`
	Tokens         []*APITokenMeta `json:"tokens"`
}

type APITokenMeta struct {
	*types.TokenMeta
	TokenName string        `json:"tokenName"`
	Pending   types.Balance `json:"pending"`
}

type APIPending struct {
	*types.PendingInfo
	TokenName string     `json:"tokenName"`
	Hash      types.Hash `json:"hash"`
}

func NewLedgerApi(l *ledger.Ledger, dpos *consensus.DposService) *LedgerApi {
	return &LedgerApi{ledger: l, dpos: dpos, logger: log.NewLogger("rpc/ledger")}
}

func (b *APIBlock) fromStateBlock(block *types.StateBlock) *APIBlock {
	b.StateBlock = block
	b.Hash = block.GetHash()
	return b
}

func (l *LedgerApi) AccountBlocksCount(addr types.Address) (int64, error) {
	am, err := l.ledger.GetAccountMeta(addr)
	if err != nil {
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

func (l *LedgerApi) generateAPIBlock(block *types.StateBlock) (*APIBlock, error) {
	ab := new(APIBlock)
	ab.StateBlock = block
	ab.Hash = block.GetHash()
	_, ab.Amount = l.ledger.CalculateAmount(block)
	token, err := l.ledger.GetTokenById(block.GetToken())
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
	bs := make([]*APIBlock, 0)
	ac, err := l.ledger.GetAccountMeta(address)
	if err != nil {
		return nil, err
	}
	for _, token := range ac.Tokens {
		h := token.Header
		for {
			block, err := l.ledger.GetStateBlock(h)
			if err != nil {
				return nil, err
			}
			b, err := l.generateAPIBlock(block)
			if err != nil {
				return nil, err
			}
			bs = append(bs, b)
			h = b.GetPrevious()
			if h.IsZero() {
				break
			}
		}
	}
	l.logger.Debug("block count,", len(bs))
	if len(bs) > o {
		if len(bs) >= o+c {
			return bs[o : c+o], nil
		}
		return bs[o:], nil
	} else {
		return make([]*APIBlock, 0), nil
	}
}

func (l *LedgerApi) AccountInfo(address types.Address) (*APIAccount, error) {
	aa := new(APIAccount)
	am, err := l.ledger.GetAccountMeta(address)
	if err != nil {
		return nil, err
	}
	for _, t := range am.Tokens {
		if t.Type == common.QLCChainToken {
			aa.CoinBalance = t.Balance
			aa.Representative = t.Representative
		}
		info, err := l.ledger.GetTokenById(t.Type)
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
		if t.Type == common.QLCChainToken {
			return t.Representative, nil
		}
	}
	return types.ZeroAddress, err
}

func (l *LedgerApi) AccountVotingWeight(addr types.Address) (types.Balance, error) {
	return l.ledger.GetRepresentation(addr)
}

func (l *LedgerApi) AccountsBalances(addresses []types.Address) (map[types.Address]map[string]map[string]types.Balance, error) {
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
			info, err := l.ledger.GetTokenById(t.Type)
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
			info, err := l.ledger.GetTokenById(token.Type)
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
		pendingkeys, err := l.ledger.Pending(addr)
		if err != nil {
			return nil, err
		}

		ps := make([]*APIPending, 0)
		for _, pendingkey := range pendingkeys {
			if len(ps) >= n {
				break
			}
			pendinginfo, err := l.ledger.GetPending(*pendingkey)
			if err != nil {
				return nil, err
			}

			token, err := l.ledger.GetTokenById(pendinginfo.Type)
			if err != nil {
				return nil, err
			}
			tokenname := token.TokenName
			tp := APIPending{
				PendingInfo: pendinginfo,
				TokenName:   tokenname,
				Hash:        pendingkey.Hash,
			}
			ps = append(ps, &tp)
		}
		if len(ps) > 0 {
			apMap[addr] = ps
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

// BlocksCount returns the number of blocks (include smartcontrant block) in the ledger and unchecked synchronizing blocks
func (l *LedgerApi) BlocksCount() (map[string]uint64, error) {
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

//BlocksCountByType reports the number of blocks in the ledger by type (send, receive, open, change)
func (l *LedgerApi) BlocksCountByType() (map[string]uint64, error) {
	c := map[string]uint64{"open": 0, "send": 0, "receive": 0, "change": 0}
	err := l.ledger.GetStateBlocks(func(block *types.StateBlock) error {
		switch block.GetType() {
		case types.Open:
			c["open"] = c["open"] + 1
		case types.Send:
			c["send"] = c["send"] + 1
		case types.Receive:
			c["receive"] = c["receive"] + 1
		case types.Change:
			c["change"] = c["change"] + 1
		}
		return nil
	})
	if err != nil {
		return nil, err
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
		b, err := l.generateAPIBlock(block)
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
	ab := make([]*APIBlock, 0)
	index := 0
	err = l.ledger.GetStateBlocks(func(block *types.StateBlock) error {
		if index >= o && index < o+c {
			b, err := l.generateAPIBlock(block)
			if err != nil {
				return err
			}
			ab = append(ab, b)
		}
		index = index + 1
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ab, nil
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
		t := am.Token(common.QLCChainToken)
		if t != nil {
			if t.Representative == hash {
				ab := &APIAccountBalance{am.Address, t.Balance}
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
		t := am.Token(common.QLCChainToken)
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
	Send      types.Address `json:"send"`
	TokenName string        `json:"tokenName"`
	To        types.Address `json:"to"`
	Amount    types.Balance `json:"amount"`
}

func (l *LedgerApi) GenerateSendBlock(para APISendBlockPara, prkStr string) (*types.StateBlock, error) {
	if para.Amount.Int == nil || para.Send.IsZero() || para.To.IsZero() || para.TokenName == "" {
		return nil, errors.New("invalid send parameter")
	}
	prk, err := hex.DecodeString(prkStr)
	if err != nil {
		return nil, err
	}
	info, err := l.ledger.GetTokenByName(para.TokenName)
	if err != nil {
		return nil, err
	}
	block, err := l.ledger.GenerateSendBlock(para.Send, info.TokenId, para.To, para.Amount, prk)
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

func (l *LedgerApi) Process(block *types.StateBlock) (types.Hash, error) {
	flag, err := l.ledger.Process(block)
	if err != nil {
		l.logger.Error(err)
		return types.ZeroHash, err
	}

	l.logger.Debug("process result, ", flag)
	switch flag {
	case ledger.Progress:
		l.logger.Debug("broadcast block")
		l.dpos.GetP2PService().Broadcast(p2p.PublishReq, block)
		return block.GetHash(), nil
	case ledger.BadWork:
		return types.ZeroHash, errors.New("bad work")
	case ledger.BadSignature:
		return types.ZeroHash, errors.New("bad signature")
	case ledger.Old:
		l.logger.Info("old block")
		//return block.GetHash(), nil
		return types.ZeroHash, errors.New("old block")
	case ledger.Fork:
		return types.ZeroHash, errors.New("fork")
	case ledger.GapSource:
		return types.ZeroHash, errors.New("gap source block")
	case ledger.GapPrevious:
		return types.ZeroHash, errors.New("gap previous block")
	case ledger.BalanceMismatch:
		return types.ZeroHash, errors.New("balance mismatch")
	case ledger.UnReceivable:
		return types.ZeroHash, errors.New("unReceivable")
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
	err := l.ledger.GetRepresentations(func(address types.Address, balance types.Balance) error {
		r := APIAccountBalance{address, balance}
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
	return l.ledger.ListTokens()
}

// BlocksCount returns the number of blocks (not include smartcontrant block) in the ledger and unchecked synchronizing blocks
func (l *LedgerApi) TransactionsCount() (map[string]uint64, error) {
	sbCount, err := l.ledger.CountStateBlocks()
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
