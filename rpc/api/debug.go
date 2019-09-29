package api

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus/dpos"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type DebugApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
	eb     event.EventBus
}

func NewDebugApi(l *ledger.Ledger, eb event.EventBus) *DebugApi {
	return &DebugApi{ledger: l, logger: log.NewLogger("api_debug"), eb: eb}
}

type APIUncheckBlock struct {
	Block       *types.StateBlock      `json:"block"`
	Link        types.Hash             `json:"link"`
	UnCheckType string                 `json:"uncheckType"`
	SyncType    types.SynchronizedKind `json:"syncType"`
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

func (l *DebugApi) UncheckBlocks() ([]*APIUncheckBlock, error) {
	unchecks := make([]*APIUncheckBlock, 0)
	err := l.ledger.WalkUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
		uncheck := new(APIUncheckBlock)
		uncheck.Block = block
		uncheck.Link = link

		switch unCheckType {
		case types.UncheckedKindPrevious:
			uncheck.UnCheckType = "GapPrevious"
		case types.UncheckedKindLink:
			uncheck.UnCheckType = "GapLink"
		case types.UncheckedKindTokenInfo:
			uncheck.UnCheckType = "GapTokenInfo"
		}

		uncheck.SyncType = sync
		unchecks = append(unchecks, uncheck)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return unchecks, nil
}

func (l *DebugApi) Dump() (string, error) {
	return l.ledger.Dump()
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

func (l *DebugApi) Representations(address *types.Address) (map[types.Address]map[string]*types.Benefit, error) {
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

func (l *DebugApi) AccountPending(address types.Address) (*APIPendingInfo, error) {
	vmContext := vmstore.NewVMContext(l.ledger)
	ap := new(APIPendingInfo)
	err := l.ledger.SearchAllKindPending(address, func(key *types.PendingKey, info *types.PendingInfo, kind types.PendingKind) error {
		token, err := abi.GetTokenById(vmContext, info.Type)
		if err != nil {
			return err
		}
		tokenName := token.TokenName
		blk, err := l.ledger.GetStateBlockConfirmed(key.Hash)
		if err != nil {
			return err
		}
		var used bool
		if kind == types.PendingUsed {
			used = true
		} else {
			used = false
		}
		ap = &APIPendingInfo{
			PendingKey:  key,
			PendingInfo: info,
			TokenName:   tokenName,
			Timestamp:   blk.Timestamp,
			Used:        used,
		}
		return nil
	})
	if err != nil {
		return nil, err
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

func (l *DebugApi) SyncBlocks() ([]types.Hash, error) {
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

func (l *DebugApi) GetOnlineInfo() (map[uint64]*dpos.RepOnlinePeriod, error) {
	repOnline := make(map[uint64]*dpos.RepOnlinePeriod, 0)
	l.eb.Publish(common.EventRpcSyncCall, "DPoS.Online", "info", repOnline)
	return repOnline, nil
}
