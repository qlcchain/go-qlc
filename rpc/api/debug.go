package api

import (
	"context"
	"errors"
	qctx "github.com/qlcchain/go-qlc/chain/context"
	"time"

	rpc "github.com/qlcchain/jsonrpc2"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus/dpos"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type DebugApi struct {
	ledger  *ledger.Ledger
	logger  *zap.SugaredLogger
	eb      event.EventBus
	feb     *event.FeedEventBus
	cfgFile string
}

func NewDebugApi(cfgFile string, eb event.EventBus) *DebugApi {
	cc := qctx.NewChainContext(cfgFile)
	return &DebugApi{
		ledger:  ledger.NewLedger(cfgFile),
		logger:  log.NewLogger("api_debug"),
		eb:      eb,
		feb:     cc.FeedEventBus(),
		cfgFile: cfgFile,
	}
}

type APIUncheckBlock struct {
	Block       *types.StateBlock      `json:"block"`
	Hash        types.Hash             `json:"hash"`
	Link        types.Hash             `json:"link"`
	UnCheckType string                 `json:"uncheckType"`
	SyncType    types.SynchronizedKind `json:"syncType"`
	Height      uint64                 `json:"povHeight"`
}

func (l *DebugApi) BlockCacheCount() (map[string]uint64, error) {
	unCount, err := l.ledger.CountBlockCache()
	if err != nil {
		return nil, err
	}
	c := make(map[string]uint64)
	c["blockCache"] = unCount
	return c, nil
}

func (l *DebugApi) BlockCaches() ([]types.Hash, error) {
	r := make([]types.Hash, 0)
	err := l.ledger.GetBlockCaches(func(block *types.StateBlock) error {
		r = append(r, block.GetHash())
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (l *DebugApi) UncheckBlocks() ([]*APIUncheckBlock, error) {
	unchecks := make([]*APIUncheckBlock, 0)
	err := l.ledger.WalkUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
		uncheck := new(APIUncheckBlock)
		uncheck.Block = block
		uncheck.Hash = block.GetHash()
		uncheck.Link = link

		switch unCheckType {
		case types.UncheckedKindPrevious:
			uncheck.UnCheckType = "GapPrevious"
		case types.UncheckedKindLink:
			uncheck.UnCheckType = "GapLink"
		case types.UncheckedKindTokenInfo:
			uncheck.UnCheckType = "GapTokenInfo"
		case types.UncheckedKindPublish:
			uncheck.UnCheckType = "GapPublish"
		}

		uncheck.SyncType = sync
		unchecks = append(unchecks, uncheck)
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = l.ledger.WalkGapPovBlocks(func(blk *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		uncheck := new(APIUncheckBlock)
		uncheck.Block = blk
		uncheck.Hash = blk.GetHash()
		uncheck.UnCheckType = "GapPovHeight"
		uncheck.SyncType = sync
		uncheck.Height = height
		unchecks = append(unchecks, uncheck)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return unchecks, nil
}

func (l *DebugApi) Action(t ledger.ActionType) (string, error) {
	return l.ledger.Action(t)
}

func (l *DebugApi) BlockLink(hash types.Hash) (map[string]types.Hash, error) {
	r := make(map[string]types.Hash)
	child, err := l.ledger.GetChild(hash)
	if err == nil {
		r["child"] = child
	}
	link, _ := l.ledger.GetLinkBlock(hash)
	if !link.IsZero() {
		r["receiver"] = link
	}
	return r, nil
}

func (l *DebugApi) BlocksCountByType(typ string) (map[string]int64, error) {
	r := make(map[string]int64)
	if err := l.ledger.GetStateBlocks(func(block *types.StateBlock) error {
		var t string
		switch typ {
		case "address":
			t = block.GetAddress().String()
		case "type":
			t = block.GetType().String()
		case "token":
			t = block.GetToken().String()
		}
		if v, ok := r[t]; ok {
			r[t] = v + 1
		} else {
			r[t] = 1
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return r, nil
}

func (l *DebugApi) GetSyncBlockNum() (map[string]uint64, error) {
	data := make(map[string]uint64)

	uncheckedSyncNum, err := l.ledger.CountUncheckedSyncBlocks()
	if err != nil {
		return nil, err
	}

	unconfirmedSyncNum, err := l.ledger.CountUnconfirmedSyncBlocks()
	if err != nil {
		return nil, err
	}

	data["uncheckedSync"] = uncheckedSyncNum
	data["unconfirmedSync"] = unconfirmedSyncNum
	return data, nil
}

func (l *DebugApi) SyncCacheBlocks() ([]types.Hash, error) {
	blocks := make([]types.Hash, 0)
	err := l.ledger.GetSyncCacheBlocks(func(block *types.StateBlock) error {
		blocks = append(blocks, block.GetHash())
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (l *DebugApi) SyncCacheBlocksCount() (int64, error) {
	var n int64
	err := l.ledger.GetSyncCacheBlocks(func(block *types.StateBlock) error {
		n++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (l *DebugApi) Representative(address types.Address) (*APIRepresentative, error) {
	balance := types.ZeroBalance
	vote := types.ZeroBalance
	network := types.ZeroBalance
	total := types.ZeroBalance
	err := l.ledger.GetAccountMetas(func(am *types.AccountMeta) error {
		t := am.Token(common.ChainToken())
		if t != nil {
			if t.Representative == address {
				balance = balance.Add(t.Balance)
				vote = vote.Add(am.CoinVote)
				network = network.Add(am.CoinNetwork)
				total = total.Add(am.VoteWeight())
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &APIRepresentative{
		Address: address,
		Balance: balance,
		Vote:    vote,
		Network: network,
		Total:   total,
	}, nil
}

func (l *DebugApi) Representatives(address *types.Address) (map[types.Address]map[string]*types.Benefit, error) {
	r := make(map[types.Address]map[string]*types.Benefit)
	if address == nil {
		err := l.ledger.GetRepresentationsCache(types.ZeroAddress, func(address types.Address, be *types.Benefit, beCache *types.Benefit) error {
			beInfo := make(map[string]*types.Benefit)
			beInfo["db"] = be
			beInfo["memory"] = beCache
			r[address] = beInfo
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		err := l.ledger.GetRepresentationsCache(*address, func(address types.Address, be *types.Benefit, beCache *types.Benefit) error {
			beInfo := make(map[string]*types.Benefit)
			beInfo["db"] = be
			beInfo["memory"] = beCache
			r[address] = beInfo
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

type APIPendingInfo struct {
	*types.PendingKey
	*types.PendingInfo
	TokenName string `json:"tokenName"`
	Timestamp int64  `json:"timestamp"`
	Used      bool   `json:"used"`
}

func (l *DebugApi) AccountPending(address types.Address, hash types.Hash) (*APIPendingInfo, error) {
	vmContext := vmstore.NewVMContext(l.ledger)
	ap := new(APIPendingInfo)
	key := &types.PendingKey{
		Address: address,
		Hash:    hash,
	}
	info, err := l.ledger.GetPending(key)
	if err != nil {
		return nil, err
	}

	token, err := abi.GetTokenById(vmContext, info.Type)
	if err != nil {
		return nil, err
	}
	tokenName := token.TokenName
	blk, err := l.ledger.GetStateBlockConfirmed(key.Hash)
	if err != nil {
		return nil, err
	}
	ap = &APIPendingInfo{
		PendingKey:  key,
		PendingInfo: info,
		TokenName:   tokenName,
		Timestamp:   blk.Timestamp,
	}

	return ap, nil
}

func (l *DebugApi) PendingsAmount() (map[types.Address]map[string]types.Balance, error) {
	abs := make(map[types.Address]map[string]types.Balance, 0)
	vmContext := vmstore.NewVMContext(l.ledger)
	err := l.ledger.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		token, err := abi.GetTokenById(vmContext, pendingInfo.Type)
		if err != nil {
			return err
		}
		tokenName := token.TokenName
		address := pendingKey.Address
		amount := pendingInfo.Amount
		if ab, ok := abs[address]; ok {
			if m, ok := ab[tokenName]; ok {
				abs[address][tokenName] = m.Add(amount)
			} else {
				abs[address][tokenName] = amount
			}
		} else {
			abs[address] = make(map[string]types.Balance)
			abs[address][tokenName] = amount
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return abs, nil
}

func (l *DebugApi) PendingsCount() (int, error) {
	n := 0
	if err := l.ledger.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		n++
		return nil
	}); err != nil {
		return 0, err
	}
	return n, nil
}

func (l *DebugApi) GetOnlineInfo() (map[uint64]*dpos.RepOnlinePeriod, error) {
	repOnline := make(map[uint64]*dpos.RepOnlinePeriod, 0)

	cc := qctx.NewChainContext(l.cfgFile)
	sv, err := cc.Service(qctx.ConsensusService)
	if err != nil {
		return nil, err
	}
	sv.RpcCall(common.RpcDPosOnlineInfo, nil, repOnline)

	return repOnline, nil
}

func (l *DebugApi) GetPovInfo() (map[string]interface{}, error) {
	inArgs := make(map[string]interface{})
	outArgs := make(map[string]interface{})

	l.feb.RpcSyncCall(&topic.EventRPCSyncCallMsg{Name: "Debug.PovInfo", In: inArgs, Out: outArgs})

	err, ok := outArgs["err"]
	if !ok {
		return nil, errors.New("api not support")
	}
	if err != nil {
		err := outArgs["err"].(error)
		return nil, err
	}
	delete(outArgs, "err")

	return outArgs, nil
}

func (l *DebugApi) NewBlock(ctx context.Context) (*rpc.Subscription, error) {
	l.logger.Infof("debug blocks ctx: %p", ctx)
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, rpc.ErrNotificationsUnsupported
	}

	// by explicitly creating an subscription we make sure that the subscription id is send back to the client
	// before the first subscription.Notify is called.
	subscription := notifier.CreateSubscription()
	go func() {
		t := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-t.C:
				if err := notifier.Notify(subscription.ID, mock.StateBlock()); err != nil {
					l.logger.Errorf("notify error: %s", err)
					return
				}
				l.logger.Info("notify success!")
			case err := <-subscription.Err():
				l.logger.Infof("subscription exception %s", err)
				return
			}
		}
	}()

	if subscription == nil {
		return nil, errors.New("create subscription error")
	}
	l.logger.Infof("blocks subscription: %s", subscription.ID)
	return subscription, nil
}

func (l *DebugApi) ContractCount() (map[string]int64, error) {
	r := make(map[string]int64)
	ctx := vmstore.NewVMContext(l.ledger)
	for _, addr := range types.ChainContractAddressList {
		var n int64 = 0
		if err := ctx.Iterator(addr[:], func(key []byte, value []byte) error {
			n++
			return nil
		}); err != nil {
			return nil, err
		}
		r[addr.String()] = n
	}
	return r, nil
}

func (l *DebugApi) GetConsInfo() (map[string]interface{}, error) {
	inArgs := make(map[string]interface{})
	outArgs := make(map[string]interface{})

	cc := qctx.NewChainContext(l.cfgFile)
	sv, err := cc.Service(qctx.ConsensusService)
	if err != nil {
		return nil, err
	}
	sv.RpcCall(common.RpcDPosConsInfo, inArgs, outArgs)

	er, ok := outArgs["err"]
	if !ok {
		return nil, errors.New("api not support")
	}
	if er != nil {
		err := outArgs["err"].(error)
		return nil, err
	}
	delete(outArgs, "err")

	return outArgs, nil
}

func (l *DebugApi) SetConsPerf(op int) (map[string]interface{}, error) {
	outArgs := make(map[string]interface{})

	cc := qctx.NewChainContext(l.cfgFile)
	sv, err := cc.Service(qctx.ConsensusService)
	if err != nil {
		return nil, err
	}
	sv.RpcCall(common.RpcDPosSetConsPerf, int32(op), outArgs)

	er, ok := outArgs["err"]
	if !ok {
		return nil, errors.New("api not support")
	}
	if er != nil {
		err := outArgs["err"].(error)
		return nil, err
	}
	delete(outArgs, "err")

	return outArgs, nil
}

func (l *DebugApi) GetConsPerf() (map[string]interface{}, error) {
	inArgs := make(map[string]interface{})
	outArgs := make(map[string]interface{})

	cc := qctx.NewChainContext(l.cfgFile)
	sv, err := cc.Service(qctx.ConsensusService)
	if err != nil {
		return nil, err
	}
	sv.RpcCall(common.RpcDPosGetConsPerf, inArgs, outArgs)

	er, ok := outArgs["err"]
	if !ok {
		return nil, errors.New("api not support")
	}
	if er != nil {
		err := outArgs["err"].(error)
		return nil, err
	}
	delete(outArgs, "err")

	return outArgs, nil
}

//func (l *DebugApi) Rollback(hash types.Hash) error {
//	lv := process.NewLedgerVerifier(l.ledger)
//	return lv.Rollback(hash)
//}
