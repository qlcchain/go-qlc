package api

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	rpc "github.com/qlcchain/jsonrpc2"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/sync/hashmap"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/ledger/relation"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var (
	ErrParameterNil = errors.New("parameter is nil")
)

type lockStatus uint8

const (
	using lockStatus = iota
	idle
)

const defaultLockSize = 1000

type lockValue struct {
	lockStatus atomic.Value
	mutex      *sync.Mutex
}

type LedgerApi struct {
	ledger *ledger.Ledger
	//vmContext *vmstore.VMContext
	eb                event.EventBus
	relation          *relation.Relation
	logger            *zap.SugaredLogger
	blockSubscription *BlockSubscription
	processLock       *hashmap.HashMap
	syncState         atomic.Value
}

type APIBlock struct {
	*types.StateBlock
	TokenName string        `json:"tokenName"`
	Amount    types.Balance `json:"amount"`
	Hash      types.Hash    `json:"hash"`

	PovConfirmHeight uint64 `json:"povConfirmHeight"`
	PovConfirmCount  uint64 `json:"povConfirmCount"`
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

type APIAccountsBalance struct {
	Balance types.Balance  `json:"balance"`
	Vote    *types.Balance `json:"vote,omitempty"`
	Network *types.Balance `json:"network,omitempty"`
	Storage *types.Balance `json:"storage,omitempty"`
	Oracle  *types.Balance `json:"oracle,omitempty"`
	Pending types.Balance  `json:"pending"`
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

	BlockType types.BlockType `json:"blockType"`
}

type ApiTokenInfo struct {
	types.TokenInfo
}

func NewLedgerApi(l *ledger.Ledger, relation *relation.Relation, eb event.EventBus) *LedgerApi {
	api := LedgerApi{
		ledger:            l,
		eb:                eb,
		relation:          relation,
		logger:            log.NewLogger("api_ledger"),
		blockSubscription: NewBlockSubscription(eb),
		processLock:       hashmap.New(defaultLockSize),
	}
	api.syncState.Store(common.SyncNotStart)
	_, _ = eb.SubscribeSync(common.EventPovSyncState, api.OnPovSyncState)
	return &api
}

func (b *APIBlock) fromStateBlock(block *types.StateBlock) *APIBlock {
	b.StateBlock = block
	b.Hash = block.GetHash()
	return b
}

func (l *LedgerApi) OnPovSyncState(state common.SyncState) {
	l.logger.Infof("ledger receive pov sync state [%s]", state)
	l.syncState.Store(state)
}

func (l *LedgerApi) AccountBlocksCount(addr types.Address) (int64, error) {
	am, err := l.ledger.GetAccountMetaConfirmed(addr)
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

func generateAPIBlock(ctx *vmstore.VMContext, block *types.StateBlock, latestPov *types.PovHeader) (*APIBlock, error) {
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

	// pov tx lookup
	if latestPov != nil {
		povTxl, _ := ctx.GetPovTxLookup(ab.Hash)
		if povTxl != nil {
			ab.PovConfirmHeight = povTxl.BlockHeight
			if latestPov.GetHeight() > ab.PovConfirmHeight {
				ab.PovConfirmCount = latestPov.GetHeight() - ab.PovConfirmHeight
			}
		}
	}

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
	vmContext := vmstore.NewVMContext(l.ledger)

	latestPov, err := l.ledger.GetLatestPovHeader()

	for _, h := range hashes {
		block, _ := l.ledger.GetStateBlockConfirmed(h)
		b, err := generateAPIBlock(vmContext, block, latestPov)
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
	vmContext := vmstore.NewVMContext(l.ledger)
	for _, t := range am.Tokens {
		if t.Type == common.ChainToken() {
			aa.CoinBalance = &t.Balance
			aa.Representative = &t.Representative
			aa.CoinVote = &am.CoinVote
			aa.CoinNetwork = &am.CoinNetwork
			aa.CoinOracle = &am.CoinOracle
			aa.CoinStorage = &am.CoinStorage
		}
		info, err := abi.GetTokenById(vmContext, t.Type)
		if err != nil {
			return nil, err
		}
		amount, err := l.ledger.PendingAmount(address, t.Type)
		if err != nil {
			l.logger.Errorf("pending amount error: %s", err)
			return nil, err
		}

		tm := APITokenMeta{
			TokenMeta: t,
			TokenName: info.TokenName,
			Pending:   amount,
		}
		aa.Tokens = append(aa.Tokens, &tm)
	}
	aa.Address = address
	return aa, nil
}

func (l *LedgerApi) ConfirmedAccountInfo(address types.Address) (*APIAccount, error) {
	aa := new(APIAccount)
	am, err := l.ledger.GetAccountMetaConfirmed(address)
	if err != nil {
		return nil, err
	}
	vmContext := vmstore.NewVMContext(l.ledger)
	for _, t := range am.Tokens {
		if t.Type == common.ChainToken() {
			aa.CoinBalance = &t.Balance
			aa.Representative = &t.Representative
			aa.CoinVote = &am.CoinVote
			aa.CoinNetwork = &am.CoinNetwork
			aa.CoinOracle = &am.CoinOracle
			aa.CoinStorage = &am.CoinStorage
		}
		info, err := abi.GetTokenById(vmContext, t.Type)
		if err != nil {
			return nil, err
		}
		amount, err := l.ledger.PendingAmount(address, t.Type)
		if err != nil {
			l.logger.Errorf("pending amount error: %s", err)
			return nil, err
		}

		tm := APITokenMeta{
			TokenMeta: t,
			TokenName: info.TokenName,
			Pending:   amount,
		}
		aa.Tokens = append(aa.Tokens, &tm)
	}
	aa.Address = address
	return aa, nil
}

func (l *LedgerApi) AccountRepresentative(addr types.Address) (types.Address, error) {
	am, err := l.ledger.GetAccountMetaConfirmed(addr)
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

func (l *LedgerApi) AccountsBalance(addresses []types.Address) (map[types.Address]map[string]*APIAccountsBalance, error) {
	as := make(map[types.Address]map[string]*APIAccountsBalance)
	vmContext := vmstore.NewVMContext(l.ledger)
	for _, addr := range addresses {
		ac, err := l.ledger.GetAccountMetaConfirmed(addr)
		if err != nil {
			if err == ledger.ErrAccountNotFound {
				continue
			}
			return nil, err
		}
		ts := make(map[string]*APIAccountsBalance)
		for _, t := range ac.Tokens {
			info, err := abi.GetTokenById(vmContext, t.Type)
			if err != nil {
				return nil, err
			}
			b := new(APIAccountsBalance)
			pendings, err := l.ledger.TokenPendingInfo(addr, t.Type)
			if err != nil {
				return nil, err
			}
			amount := types.ZeroBalance
			for _, pending := range pendings {
				amount = amount.Add(pending.Amount)
			}
			b.Balance = t.Balance
			b.Pending = amount
			if info.TokenId == common.ChainToken() {
				b.Vote = &ac.CoinVote
				b.Network = &ac.CoinNetwork
				b.Oracle = &ac.CoinOracle
				b.Storage = &ac.CoinStorage
			}
			ts[info.TokenName] = b
		}
		as[addr] = ts
	}
	return as, nil
}

func (l *LedgerApi) AccountsFrontiers(addresses []types.Address) (map[types.Address]map[string]types.Hash, error) {
	r := make(map[types.Address]map[string]types.Hash)
	vmContext := vmstore.NewVMContext(l.ledger)
	for _, addr := range addresses {
		ac, err := l.ledger.GetAccountMetaConfirmed(addr)
		if err != nil {
			if err == ledger.ErrAccountNotFound {
				continue
			}
			return nil, err
		}
		t := make(map[string]types.Hash)
		for _, token := range ac.Tokens {
			info, err := abi.GetTokenById(vmContext, token.Type)
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
	vmContext := vmstore.NewVMContext(l.ledger)

	for _, addr := range addresses {
		ps := make([]*APIPending, 0)
		err := l.ledger.SearchPending(addr, func(key *types.PendingKey, info *types.PendingInfo) error {
			token, err := abi.GetTokenById(vmContext, info.Type)
			if err != nil {
				return err
			}
			tokenName := token.TokenName
			blk, err := l.ledger.GetStateBlockConfirmed(key.Hash)
			if err != nil {
				return err
			}
			ap := APIPending{
				PendingKey:  key,
				PendingInfo: info,

				TokenName: tokenName,
				Timestamp: blk.Timestamp,

				BlockType: blk.Type,
			}
			ps = append(ps, &ap)
			return nil
		})

		if err != nil {
			l.logger.Error(err)
		}
		if len(ps) > 0 {
			pt := ps
			if n > -1 && len(ps) > n {
				pt = ps[:n]
			}

			sort.Slice(pt, func(i, j int) bool {
				return pt[i].Timestamp < pt[j].Timestamp
			})
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
	sb, err := l.ledger.GetStateBlockConfirmed(hash)
	if err != nil {
		return types.ZeroAddress, err
	}
	return sb.GetAddress(), nil
}

func (l *LedgerApi) BlockConfirmedStatus(hash types.Hash) (bool, error) {
	b, err := l.ledger.HasStateBlockConfirmed(hash)
	if err != nil {
		return false, err
	}
	return b, nil
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
	c[types.Online.String()] = 0
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
	vmContext := vmstore.NewVMContext(l.ledger)

	latestPov, _ := l.ledger.GetLatestPovHeader()

	for _, h := range hash {
		block, err := l.ledger.GetStateBlock(h)
		if err != nil {
			if err == ledger.ErrBlockNotFound {
				continue
			}
			return nil, fmt.Errorf("%s, %s", h, err)
		}
		b, err := generateAPIBlock(vmContext, block, latestPov)
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
	vmContext := vmstore.NewVMContext(l.ledger)

	latestPov, err := l.ledger.GetLatestPovHeader()

	for _, h := range hashes {
		block, err := l.ledger.GetStateBlock(h)
		if err != nil && err != ledger.ErrBlockNotFound {
			return nil, err
		}
		if block != nil {
			b, err := generateAPIBlock(vmContext, block, latestPov)
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
		blk, err := l.ledger.GetStateBlockConfirmed(hash)
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

func (l *LedgerApi) GenerateSendBlock(para *APISendBlockPara, prkStr *string) (*types.StateBlock, error) {
	if para == nil {
		return nil, ErrParameterNil
	}
	if para.Amount.Int == nil || para.From.IsZero() || para.To.IsZero() || para.TokenName == "" {
		return nil, errors.New("invalid transaction parameter")
	}
	if ss := l.syncState.Load().(common.SyncState); ss != common.SyncDone {
		return nil, errors.New("pov sync is not finished, please check it")
	}
	var prk []byte
	if prkStr != nil {
		var err error
		if prk, err = hex.DecodeString(*prkStr); err != nil {
			return nil, err
		}
	}
	vmContext := vmstore.NewVMContext(l.ledger)
	info, err := abi.GetTokenByName(vmContext, para.TokenName)
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

func (l *LedgerApi) GenerateReceiveBlock(sendBlock *types.StateBlock, prkStr *string) (*types.StateBlock, error) {
	if sendBlock == nil {
		return nil, ErrParameterNil
	}
	if ss := l.syncState.Load().(common.SyncState); ss != common.SyncDone {
		return nil, errors.New("pov sync is not finished, please check it")
	}
	var prk []byte
	if prkStr != nil {
		var err error
		if prk, err = hex.DecodeString(*prkStr); err != nil {
			return nil, err
		}
	}
	block, err := l.ledger.GenerateReceiveBlock(sendBlock, prk)
	if err != nil {
		return nil, err
	}
	l.logger.Debug(block)
	return block, nil
}

func (l *LedgerApi) GenerateReceiveBlockByHash(sendHash types.Hash, prkStr *string) (*types.StateBlock, error) {
	if ss := l.syncState.Load().(common.SyncState); ss != common.SyncDone {
		return nil, errors.New("pov sync is not finished, please check it")
	}
	var prk []byte
	if prkStr != nil {
		var err error
		if prk, err = hex.DecodeString(*prkStr); err != nil {
			return nil, err
		}
	}

	sendBlock, err := l.ledger.GetStateBlock(sendHash)
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

func (l *LedgerApi) GenerateChangeBlock(account types.Address, representative types.Address, prkStr *string) (*types.StateBlock, error) {
	var prk []byte
	if prkStr != nil {
		var err error
		if prk, err = hex.DecodeString(*prkStr); err != nil {
			return nil, err
		}
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
	vmContext := vmstore.NewVMContext(l.ledger)
	err := l.ledger.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		token, err := abi.GetTokenById(vmContext, pendingInfo.Type)
		if err != nil {
			return err
		}
		tokenName := token.TokenName
		blk, err := l.ledger.GetStateBlockConfirmed(pendingKey.Hash)
		if err != nil {
			l.logger.Errorf("can not fetch block from %s, %s", util.ToString(pendingKey), util.ToString(pendingInfo))
		} else {
			ap := APIPending{
				PendingKey:  pendingKey,
				PendingInfo: pendingInfo,
				TokenName:   tokenName,
				Timestamp:   blk.Timestamp,
			}
			aps = append(aps, &ap)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(aps) > 0 {
		sort.Slice(aps, func(i, j int) bool {
			return aps[i].Timestamp < aps[j].Timestamp
		})
	}

	return aps, nil
}

func (l *LedgerApi) getLockKey(addr types.Address, token types.Hash) []byte {
	key := make([]byte, 0)
	key = append(key, addr.Bytes()...)
	key = append(key, token.Bytes()...)
	return key
}

func (l *LedgerApi) getProcessLock(addr types.Address, token types.Hash) *lockValue {
	key := l.getLockKey(addr, token)
	v, b := l.processLock.Get(key)
	if b {
		v.(*lockValue).lockStatus.Store(using)
		return v.(*lockValue)
	} else {
		lv := &lockValue{}
		lv.lockStatus.Store(using)
		lv.mutex = &sync.Mutex{}
		l.processLock.Set(key, lv)
		if l.processLock.Len() >= defaultLockSize {
			for key := range l.processLock.Iter() {
				s := (key.Value).(*lockValue).lockStatus.Load()
				if s.(lockStatus) == idle {
					l.processLock.Del(key.Key)
				}
			}
		}
		return lv
	}
}

func (l *LedgerApi) Process(block *types.StateBlock) (types.Hash, error) {
	if block == nil {
		return types.ZeroHash, ErrParameterNil
	}
	if ss := l.syncState.Load().(common.SyncState); ss != common.SyncDone {
		return types.ZeroHash, errors.New("pov sync is not finished, please check it")
	}
	p := make(map[string]string)
	l.eb.Publish(common.EventPeersInfo, p)
	if len(p) == 0 {
		return types.ZeroHash, errors.New("no peer connect,please check it")
	}
	lv := l.getProcessLock(block.Address, block.Token)
	lv.mutex.Lock()
	defer func() {
		lv.mutex.Unlock()
		lv.lockStatus.Store(idle)
	}()
	verifier := process.NewLedgerVerifier(l.ledger)
	flag, err := verifier.BlockCacheCheck(block)
	if err != nil {
		l.logger.Error(err)
		return types.ZeroHash, err
	}

	l.logger.Debug("process result, ", flag)
	switch flag {
	case process.Progress:
		hash := block.GetHash()
		verify := process.NewLedgerVerifier(l.ledger)
		err := verify.BlockCacheProcess(block)
		if err != nil {
			l.logger.Errorf("Block %s add to blockCache error[%s]", hash, err)
			return types.ZeroHash, err
		}

		l.logger.Debug("broadcast block")
		//TODO: refine
		l.eb.Publish(common.EventBroadcast, p2p.PublishReq, block)
		l.eb.Publish(common.EventGenerateBlock, flag, block)
		return hash, nil
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
	case process.ReceiveRepeated:
		return types.ZeroHash, errors.New("generate receive block repeatedly ")
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

type APIRepresentative struct {
	Address types.Address `json:"address"`
	Balance types.Balance `json:"balance"`
	Vote    types.Balance `json:"vote"`
	Network types.Balance `json:"network"`
	Storage types.Balance `json:"storage"`
	Oracle  types.Balance `json:"oracle"`
	Total   types.Balance `json:"total"`
}

//Representatives returns a list of pairs of representative and its voting weight
func (l *LedgerApi) Representatives(sorting *bool) ([]*APIRepresentative, error) {
	rs := make([]*APIRepresentative, 0)
	err := l.ledger.GetRepresentations(func(address types.Address, benefit *types.Benefit) error {
		r := &APIRepresentative{
			Address: address,
			Balance: benefit.Balance,
			Vote:    benefit.Vote,
			Network: benefit.Network,
			Storage: benefit.Storage,
			Oracle:  benefit.Oracle,
			Total:   benefit.Total,
		}
		rs = append(rs, r)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if sorting != nil && *sorting {
		sort.Slice(rs, func(i, j int) bool {
			return rs[i].Total.Compare(rs[j].Total) == types.BalanceCompBigger
		})
	}
	return rs, nil
}

func (l *LedgerApi) Tokens() ([]*types.TokenInfo, error) {
	vmContext := vmstore.NewVMContext(l.ledger)
	return abi.ListTokens(vmContext)
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
	vmContext := vmstore.NewVMContext(l.ledger)
	token, err := abi.GetTokenById(vmContext, tokenId)
	if err != nil {
		return nil, err
	}
	return &ApiTokenInfo{*token}, nil
}

func (l *LedgerApi) TokenInfoByName(tokenName string) (*ApiTokenInfo, error) {
	vmContext := vmstore.NewVMContext(l.ledger)
	token, err := abi.GetTokenByName(vmContext, tokenName)
	if err != nil {
		return nil, err
	}
	return &ApiTokenInfo{*token}, nil
}

func (l *LedgerApi) GetAccountOnlineBlock(account types.Address) ([]*types.StateBlock, error) {
	blocks := make([]*types.StateBlock, 0)
	hashes, err := l.relation.AccountBlocks(account, -1, -1)
	if err == nil {
		for _, hash := range hashes {
			if blk, err := l.ledger.GetStateBlockConfirmed(hash); err == nil {
				if blk.Type == types.Online {
					blocks = append(blocks, blk)
				}
			}
		}
		return blocks, nil
	} else {
		return nil, err
	}
}

func (l *LedgerApi) NewBlock(ctx context.Context) (*rpc.Subscription, error) {
	ch := make(chan *types.StateBlock)
	l.logger.Infof("blocks ctx: %p, ch %p", ctx, ch)
	l.blockSubscription.addChan(ch)
	sub, err := createSubscription(ctx, func(notifier *rpc.Notifier, subscription *rpc.Subscription) {
		go func() {
			for {
				select {
				case block := <-ch:
					if err := notifier.Notify(subscription.ID, block); err != nil {
						l.logger.Errorf("notify error: %s", err)
						return
					}
				case err := <-subscription.Err():
					l.logger.Infof("subscription exception %s", err)
					l.blockSubscription.removeChan(ch)
					return
				}
			}
		}()
	})
	if err != nil || sub == nil {
		l.logger.Errorf("create subscription error, %s", err)
		return nil, err
	}
	l.logger.Infof("blocks subscription: %s", sub.ID)
	return sub, nil
}

func (l *LedgerApi) BalanceChange(ctx context.Context, address types.Address) (*rpc.Subscription, error) {
	ch := make(chan *types.StateBlock)
	l.logger.Infof("amount ctx: %p, ch %p", ctx, ch)
	l.blockSubscription.addChan(ch)
	return createSubscription(ctx, func(notifier *rpc.Notifier, subscription *rpc.Subscription) {
		go func() {
			for {
				select {
				case block := <-ch:
					if block.GetAddress() == address {
						ac, err := l.ledger.GetAccountMeta(address)
						if err != nil {
							l.logger.Errorf("get account meta: %s", err)
							return
						}
						if err := notifier.Notify(subscription.ID, ac); err != nil {
							l.logger.Errorf("notify error: %s", err)
							return
						}
					}
				case err := <-subscription.Err():
					l.logger.Infof("subscription exception %s", err)
					l.blockSubscription.removeChan(ch)
					return
				}
			}
		}()
	})
}

func (l *LedgerApi) NewPending(ctx context.Context, address types.Address) (*rpc.Subscription, error) {
	ch := make(chan *types.StateBlock)
	l.blockSubscription.addChan(ch)
	return createSubscription(ctx, func(notifier *rpc.Notifier, subscription *rpc.Subscription) {
		go func() {
			for {
				select {
				case block := <-ch:
					if block.IsSendBlock() {
						pk := &types.PendingKey{
							Address: address,
							Hash:    block.GetHash(),
						}
						if pi, _ := l.ledger.GetPending(pk); pi != nil {
							vmContext := vmstore.NewVMContext(l.ledger)
							token, err := abi.GetTokenById(vmContext, pi.Type)
							if err != nil {
								l.logger.Errorf("get token info: %s", err)
								return
							}
							blk, err := l.ledger.GetStateBlockConfirmed(pk.Hash)
							if err != nil {
								l.logger.Errorf("get block info: %s", err)
								return
							}

							ap := APIPending{
								PendingKey:  pk,
								PendingInfo: pi,
								TokenName:   token.TokenName,
								Timestamp:   blk.Timestamp,
							}
							if err := notifier.Notify(subscription.ID, ap); err != nil {
								l.logger.Errorf("notify error: %s", err)
								return
							}
						}
					}
				case err := <-subscription.Err():
					l.logger.Infof("subscription exception %s", err)
					l.blockSubscription.removeChan(ch)
					return
				}
			}
		}()
	})
}
