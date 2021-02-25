package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	rpc "github.com/qlcchain/jsonrpc2"
	"go.uber.org/zap"

	qctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/chain/version"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/storage"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus/dpos"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/mock"
)

type DebugApi struct {
	ledger ledger.Store
	logger *zap.SugaredLogger
	eb     event.EventBus
	feb    *event.FeedEventBus
	cc     *qctx.ChainContext

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
		cc:      cc,
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
	unCount, err := l.ledger.CountBlocksCache()
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

func (l *DebugApi) Action(at storage.ActionType, t int) (string, error) {
	r, err := l.ledger.Action(at, t)
	if err != nil {
		return "", err
	}
	if r == nil {
		return "", nil
	} else {
		if s, ok := r.(string); ok {
			return s, nil
		} else {
			return "", errors.New("error action")
		}
	}
}

func (l *DebugApi) BlockLink(hash types.Hash) (map[string]types.Hash, error) {
	r := make(map[string]types.Hash)
	child, err := l.ledger.GetBlockChild(hash)
	if err == nil {
		r["child"] = child
	}
	link, _ := l.ledger.GetBlockLink(hash)
	if !link.IsZero() {
		r["receiver"] = link
	}
	return r, nil
}

func (l *DebugApi) BlockLinks(hash types.Hash) (map[string][]types.Hash, error) {
	r := make(map[string][]types.Hash)
	r["child"] = make([]types.Hash, 0)
	r["receiver"] = make([]types.Hash, 0)
	err := l.ledger.GetStateBlocksConfirmed(func(block *types.StateBlock) error {
		if block.Previous == hash {
			r["child"] = append(r["child"], block.GetHash())
		}
		if block.Link == hash {
			r["receiver"] = append(r["receiver"], block.GetHash())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (l *DebugApi) BlocksCountByType(typ string) (map[string]int64, error) {
	r := make(map[string]int64)
	if err := l.ledger.GetStateBlocksConfirmed(func(block *types.StateBlock) error {
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

//func (l *DebugApi) SyncCacheBlocks() ([]types.Hash, error) {
//	blocks := make([]types.Hash, 0)
//	err := l.ledger.GetSyncCacheBlocks(func(block *types.StateBlock) error {
//		blocks = append(blocks, block.GetHash())
//		return nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	return blocks, nil
//}

//func (l *DebugApi) SyncCacheBlocksCount() (int64, error) {
//	var n int64
//	err := l.ledger.GetSyncCacheBlocks(func(block *types.StateBlock) error {
//		n++
//		return nil
//	})
//	if err != nil {
//		return 0, err
//	}
//	return n, nil
//}

func (l *DebugApi) Representative(address types.Address) (*APIRepresentative, error) {
	balance := types.ZeroBalance
	vote := types.ZeroBalance
	network := types.ZeroBalance
	total := types.ZeroBalance
	err := l.ledger.GetAccountMetas(func(am *types.AccountMeta) error {
		t := am.Token(config.ChainToken())
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

type APIPendingInfo struct {
	*types.PendingKey
	*types.PendingInfo
	TokenName string `json:"tokenName"`
	Timestamp int64  `json:"timestamp"`
	Used      bool   `json:"used"`
}

func (l *DebugApi) AccountPending(address types.Address, hash types.Hash) (*APIPendingInfo, error) {
	ap := new(APIPendingInfo)
	key := &types.PendingKey{
		Address: address,
		Hash:    hash,
	}
	info, err := l.ledger.GetPending(key)
	if err != nil {
		return nil, err
	}

	token, err := l.ledger.GetTokenById(info.Type)
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
	err := l.ledger.GetPendings(func(pendingKey *types.PendingKey, pendingInfo *types.PendingInfo) error {
		token, err := l.ledger.GetTokenById(pendingInfo.Type)
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

	sv, err := l.getConsensusService()
	if err != nil {
		return nil, err
	}
	sv.RpcCall(common.RpcDPoSOnlineInfo, nil, repOnline)

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
	for _, addr := range contractaddress.ChainContractAddressList {
		var n int64 = 0
		iterator := l.ledger.NewVMIterator(&addr)
		if err := iterator.Next(nil, func(key []byte, value []byte) error {
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

	sv, err := l.getConsensusService()
	if err != nil {
		return nil, err
	}
	sv.RpcCall(common.RpcDPoSConsInfo, inArgs, outArgs)

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

	sv, err := l.getConsensusService()
	if err != nil {
		return nil, err
	}
	sv.RpcCall(common.RpcDPoSSetConsPerf, dpos.PerfType(op), outArgs)

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

	sv, err := l.getConsensusService()
	if err != nil {
		return nil, err
	}
	sv.RpcCall(common.RpcDPoSGetConsPerf, inArgs, outArgs)

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

type CacheStat struct {
	Index  int    `json:"index"`
	Key    int    `json:"key"`
	Block  int    `json:"block"`
	Delete int    `json:"delete"`
	Start  string `json:"start"`
	Span   string `json:"span"`
}

func (l *DebugApi) GetCache() error {
	l.ledger.Cache().GetCache()
	return nil
}

func (l *DebugApi) GetCacheStat() []*CacheStat {
	cs := l.ledger.GetCacheStat()
	cas := make([]*CacheStat, 0)
	for i := len(cs) - 1; i >= 0; i-- {
		c := cs[i]
		ca := new(CacheStat)
		ca.Index = c.Index
		ca.Key = c.Key
		ca.Block = c.Block
		ca.Start = time.Unix(c.Start/1000000000, 0).Format("2006-01-02 15:04:05")
		if c.End == 0 {
			ca.Span = "ms"
		} else {
			ca.Span = strconv.FormatInt((c.End-c.Start)/1000000, 10) + "ms"
		}
		cas = append(cas, ca)
	}

	return cas
}

func (l *DebugApi) GetCacheStatus() map[string]string {
	return l.ledger.GetCacheStatue()
}

func (l *DebugApi) GetUCache() error {
	l.ledger.UCache().GetCache()
	return nil
}

func (l *DebugApi) GetUCacheStat() []*CacheStat {
	cs := l.ledger.GetUCacheStat()
	cas := make([]*CacheStat, 0)
	for i := len(cs) - 1; i >= 0; i-- {
		c := cs[i]
		ca := new(CacheStat)
		ca.Index = c.Index
		ca.Key = c.Key
		ca.Block = c.Block
		ca.Delete = c.Delete
		ca.Start = time.Unix(c.Start/1000000000, 0).Format("2006-01-02 15:04:05")
		if c.End == 0 {
			ca.Span = "ms"
		} else {
			ca.Span = strconv.FormatInt((c.End-c.Start)/1000000, 10) + "ms"
		}
		cas = append(cas, ca)
	}

	return cas
}

func (l *DebugApi) GetUCacheStatus() map[string]string {
	return l.ledger.GetUCacheStatue()
}

//func (l *DebugApi) Rollback(hash types.Hash) error {
//	lv := process.NewLedgerVerifier(l.ledger)
//	return lv.Rollback(hash)
//}

type UncheckInfo struct {
	Hash      types.Hash `json:"hash"`
	GapType   string     `json:"gapType"`
	GapHash   types.Hash `json:"gapHash"`
	GapHeight uint64     `json:"gapHeight"`
}

func (l *DebugApi) UncheckAnalysis() ([]*UncheckInfo, error) {
	uis := make([]*UncheckInfo, 0)

	err := l.ledger.GetUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
		switch unCheckType {
		case types.UncheckedKindPrevious:
		case types.UncheckedKindLink:
			if has, _ := l.ledger.HasStateBlockConfirmed(link); has {
				ui := &UncheckInfo{
					Hash:    block.GetHash(),
					GapType: "gap link",
					GapHash: link,
				}
				uis = append(uis, ui)
			}
		case types.UncheckedKindPublish:
			if has, _ := l.ledger.HasStateBlockConfirmed(link); has {
				ui := &UncheckInfo{
					Hash:    block.GetHash(),
					GapType: "gap publish",
					GapHash: link,
				}
				uis = append(uis, ui)
			}
		case types.UncheckedKindTokenInfo:
			ui := &UncheckInfo{
				Hash:    block.GetHash(),
				GapType: "gap token",
			}
			uis = append(uis, ui)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = l.ledger.WalkGapPovBlocks(func(block *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		ui := &UncheckInfo{
			Hash:      block.GetHash(),
			GapType:   "gap pov",
			GapHeight: height,
		}
		uis = append(uis, ui)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return uis, nil
}

func (l *DebugApi) uncheckBlock(hash types.Hash) *UncheckInfo {
	ui := new(UncheckInfo)
	find := false
	var uncheckTypeMap = map[types.UncheckedKind]string{
		types.UncheckedKindPrevious:  "gap previous",
		types.UncheckedKindLink:      "gap link",
		types.UncheckedKindPublish:   "gap publish",
		types.UncheckedKindTokenInfo: "gap token",
	}

	err := l.ledger.GetUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
		if block.GetHash() == hash {
			ui = &UncheckInfo{
				Hash:    block.GetHash(),
				GapType: uncheckTypeMap[unCheckType],
				GapHash: link,
			}
			find = true
		}
		return nil
	})
	if err != nil {
		return nil
	}

	err = l.ledger.WalkGapPovBlocks(func(block *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		if block.GetHash() == hash {
			ui = &UncheckInfo{
				Hash:      block.GetHash(),
				GapType:   "gap pov",
				GapHeight: height,
			}
			find = true
		}
		return nil
	})
	if err != nil {
		return nil
	}

	if find {
		return ui
	} else {
		return nil
	}
}

func (l *DebugApi) UncheckBlock(hash types.Hash) ([]*UncheckInfo, error) {
	uis := make([]*UncheckInfo, 0)
	ui := new(UncheckInfo)

	for {
		ui = l.uncheckBlock(hash)
		if ui != nil {
			fmt.Printf("%-64s\t%-16s\t\t%-64s\t%d\n", ui.Hash, ui.GapType, ui.GapHash, ui.GapHeight)
			if ui.GapType == "gap previous" || ui.GapType == "gap link" {
				hash = ui.GapHash
				if has, _ := l.ledger.HasStateBlockConfirmed(ui.GapHash); has {
					break
				}
			} else {
				break
			}
		} else {
			fmt.Printf("%s there is no uncheck info\n", hash)
			return nil, errors.New("there is no uncheck info")
		}
	}

	uis = append(uis, ui)
	return uis, nil
}

func (l *DebugApi) UncheckBlocks() ([]*APIUncheckBlock, error) {
	unchecks := make([]*APIUncheckBlock, 0)
	if err := l.ledger.GetUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
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
	}); err != nil {
		return nil, err
	}

	if err := l.ledger.WalkGapPovBlocks(func(blk *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		uncheck := new(APIUncheckBlock)
		uncheck.Block = blk
		uncheck.Hash = blk.GetHash()
		uncheck.UnCheckType = "GapPovHeight"
		uncheck.SyncType = sync
		uncheck.Height = height
		unchecks = append(unchecks, uncheck)
		return nil
	}); err != nil {
		return nil, err
	}

	if err := l.ledger.WalkGapDoDSettleStateBlock(func(blk *types.StateBlock, sync types.SynchronizedKind) error {
		uncheck := new(APIUncheckBlock)
		uncheck.Block = blk
		uncheck.Hash = blk.GetHash()
		uncheck.UnCheckType = "GapDoDSettleState"
		uncheck.SyncType = sync
		unchecks = append(unchecks, uncheck)
		return nil
	}); err != nil {
		return nil, err
	}
	return unchecks, nil
}

func (l *DebugApi) UncheckBlocksCount() (map[string]int, error) {
	unchecks := make(map[string]int)

	err := l.ledger.GetUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
		hash := block.GetHash()
		if hash.IsZero() {
			return fmt.Errorf("invalid block : %s", block)
		}
		switch unCheckType {
		case types.UncheckedKindPrevious:
			unchecks["GapPrevious"] = unchecks["GapPrevious"] + 1
		case types.UncheckedKindLink:
			unchecks["GapLink"] = unchecks["GapLink"] + 1
		case types.UncheckedKindTokenInfo:
			unchecks["GapTokenInfo"] = unchecks["GapTokenInfo"] + 1
		case types.UncheckedKindPublish:
			unchecks["GapPublish"] = unchecks["GapPublish"] + 1
		}
		unchecks["Total"] = unchecks["Total"] + 1
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = l.ledger.WalkGapPovBlocks(func(blk *types.StateBlock, height uint64, sync types.SynchronizedKind) error {
		unchecks["GapPovHeight"] = unchecks["GapPovHeight"] + 1
		unchecks["Total"] = unchecks["Total"] + 1
		return nil
	})

	if err := l.ledger.WalkGapDoDSettleStateBlock(func(blk *types.StateBlock, sync types.SynchronizedKind) error {
		unchecks["GapDoDSettleState"] = unchecks["GapDoDSettleState"] + 1
		unchecks["Total"] = unchecks["Total"] + 1
		return nil
	}); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return unchecks, nil
}

func (l *DebugApi) UncheckBlocksCountStore() (map[string]uint64, error) {
	count1, err := l.ledger.CountUncheckedBlocks()
	if err != nil {
		return nil, err
	}
	count2, err := l.ledger.CountUncheckedBlocksStore()
	if err != nil {
		return nil, err
	}
	unchecks := make(map[string]uint64)
	unchecks["cache"] = count1
	unchecks["store"] = count2
	return unchecks, nil
}

func (l *DebugApi) FeedConsensus() error {
	sv, err := l.getConsensusService()
	if err != nil {
		return err
	}
	sv.RpcCall(common.RpcDPoSFeed, nil, nil)
	return nil
}

func (l *DebugApi) DebugConsensus() error {
	sv, err := l.getConsensusService()
	if err != nil {
		return err
	}
	sv.RpcCall(common.RpcDPoSDebug, nil, nil)
	return nil
}

func (l *DebugApi) getConsensusService() (common.InterceptCall, error) {
	cc := qctx.NewChainContext(l.cfgFile)
	sv, err := cc.Service(qctx.ConsensusService)
	if err != nil {
		return nil, err
	}
	return sv.(common.InterceptCall), nil
}

func (l *DebugApi) GetPrivacyInfo() (map[string]interface{}, error) {
	inArgs := make(map[string]interface{})
	outArgs := make(map[string]interface{})

	l.feb.RpcSyncCall(&topic.EventRPCSyncCallMsg{Name: "Debug.PrivacyInfo", In: inArgs, Out: outArgs})

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

func (l *DebugApi) BadgerTableSize(keyPrefixs []storage.KeyPrefix) (int64, error) {
	var size int64
	for _, p := range keyPrefixs {
		prefix, _ := storage.GetKeyOfParts(p)
		if err := l.ledger.DBStore().Iterator(prefix, nil, func(key []byte, val []byte) error {
			size = int64(len(key)) + int64(len(val)) + size
			return nil
		}); err != nil {
			return 0, err
		}
	}
	return size, nil
}

type NodeStatus struct {
	BuildTime    string        `json:"buildTime"`
	GitRev       string        `json:"gitRev"`
	Version      string        `json:"version"`
	Mode         string        `json:"mode"`
	LedgerStatus *LedgerStatus `json:"ledgerStatus"`
	NetStatus    *NetStatus    `json:"netStatus"`
	POVStatus    *POVStatus    `json:"povStatus"`
}

type LedgerStatus struct {
	BlockCount          uint64 `json:"blockCount"`
	UnCheckedBlockCount uint64 `json:"unCheckedBlockCount"`
}

type NetStatus struct {
	PeerId                    string            `json:"peerId"`
	ConnectedPeers            []*types.PeerInfo `json:"connectedPeers"`
	OnlineRepresentatives     []types.Address   `json:"onlineRepresentatives"`
	OnlineRepresentativeInfos *OnlineRepTotal   `json:"onlineRepresentativeInfos"`
}

type POVStatus struct {
	PovEnabled     bool                 `json:"povEnabled"`
	SyncState      int                  `json:"syncState"`
	SyncStateStr   string               `json:"syncStateStr"`
	PovLedgerStats *PovLedgerStats      `json:"povLedgerStats"`
	PovMiningInfo  *PovApiGetMiningInfo `json:"povMiningInfo"`
	LatestHeader   *PovApiHeader        `json:"latestHeader"`
}

func (l *DebugApi) NodeStatus() (*NodeStatus, error) {
	nodeStatus := new(NodeStatus)
	nodeStatus.BuildTime = version.BuildTime
	nodeStatus.Mode = version.Mode
	nodeStatus.GitRev = version.GitRev
	nodeStatus.Version = version.Version
	sbCount, err := l.ledger.CountStateBlocks()
	if err != nil {
		return nil, err
	}
	unCount, err := l.ledger.CountUncheckedBlocksStore()
	if err != nil {
		return nil, err
	}
	nodeStatus.LedgerStatus = &LedgerStatus{
		BlockCount:          sbCount,
		UnCheckedBlockCount: unCount,
	}
	nodeStatus.NetStatus = new(NetStatus)
	cfg, _ := l.cc.Config()
	nodeStatus.NetStatus.PeerId = cfg.P2P.ID.PeerID
	p := l.cc.GetConnectPeersInfo()
	nodeStatus.NetStatus.ConnectedPeers = p
	nodeStatus.NetStatus.OnlineRepresentativeInfos = onlineRepsInfo(l.ledger)
	as, _ := l.ledger.GetOnlineRepresentations()
	nodeStatus.NetStatus.OnlineRepresentatives = as

	nodeStatus.POVStatus = new(POVStatus)
	povLedger, err := getLedgerStats(l.ledger)
	if err != nil {
		return nil, err
	}
	nodeStatus.POVStatus.PovLedgerStats = povLedger
	if cfg.PoV.PovEnabled {
		if MiningInfo, err := getMiningInfo(l.ledger, l.feb); err != nil {
			return nil, err
		} else {
			nodeStatus.POVStatus.PovMiningInfo = MiningInfo
		}
	}
	nodeStatus.POVStatus.PovEnabled = cfg.PoV.PovEnabled
	ss := l.cc.PoVState()
	nodeStatus.POVStatus.SyncState = int(ss)
	nodeStatus.POVStatus.SyncStateStr = ss.String()
	latestHeader, _ := getLatestHeader(l.ledger)
	nodeStatus.POVStatus.LatestHeader = latestHeader
	return nodeStatus, nil
}
