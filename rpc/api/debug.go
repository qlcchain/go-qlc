package api

import (
	"fmt"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type DebugApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

func NewDebugApi(l *ledger.Ledger) *DebugApi {
	return &DebugApi{ledger: l, logger: log.NewLogger("api_debug")}
}

type APIUncheckBlock struct {
	Block       *types.StateBlock      `json:"block"`
	Link        types.Hash             `json:"link"`
	UnCheckType string                 `json:"uncheckType"`
	SyncType    types.SynchronizedKind `json:"syncType"`
}

func (d *DebugApi) UncheckBlocks() ([]*APIUncheckBlock, error) {
	unchecks := make([]*APIUncheckBlock, 0)
	err := d.ledger.WalkUncheckedBlocks(func(block *types.StateBlock, link types.Hash, unCheckType types.UncheckedKind, sync types.SynchronizedKind) error {
		uncheck := new(APIUncheckBlock)
		uncheck.Block = block
		uncheck.Link = link
		if unCheckType == 0 {
			uncheck.UnCheckType = "GapPrevious"
		}
		if unCheckType == 1 {
			uncheck.UnCheckType = "GapLink"
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

func (d *DebugApi) BlockLink(hash types.Hash) (map[string]types.Hash, error) {
	blk, err := d.ledger.GetStateBlock(hash)
	if err != nil {
		return nil, fmt.Errorf("get block err: %s", hash.String())
	}
	r := make(map[string]types.Hash)
	children, _ := d.ledger.GetChildren(hash)
	for key, val := range children {
		if val == 0 {
			cblk, err := d.ledger.GetStateBlock(key)
			if err != nil {
				return nil, fmt.Errorf("get child1 err: %s", key.String())
			}
			if cblk.GetAddress() != blk.GetAddress() {
				return nil, fmt.Errorf("address not equal, %s, %s", blk.GetHash(), cblk.GetHash())
			}
			r["child1"] = key
		}
		if val == 1 {
			cblk, err := d.ledger.GetStateBlock(key)
			if err != nil {
				return nil, fmt.Errorf("get child2 err: %s", key.String())
			}
			if cblk.GetAddress() == blk.GetAddress() {
				return nil, fmt.Errorf("address equal, %s, %s", blk.GetHash(), cblk.GetHash())
			}
			r["child2"] = key
		}
	}
	link, _ := d.ledger.GetLinkBlock(hash)
	if !link.IsZero() {
		r["receiver"] = link
	}
	return r, nil
}

func (d *DebugApi) GetAccountMetasCache(address *types.Address) (map[types.Address]map[string]*types.AccountMeta, error) {
	r := make(map[types.Address]map[string]*types.AccountMeta)
	if address == nil {
		err := d.ledger.GetAccountMetasCache(types.ZeroAddress, func(address types.Address, am *types.AccountMeta, amCache *types.AccountMeta) error {
			amInfo := make(map[string]*types.AccountMeta)
			amInfo["db"] = am
			amInfo["cache"] = amCache
			r[address] = amInfo
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		err := d.ledger.GetAccountMetasCache(*address, func(address types.Address, am *types.AccountMeta, amCache *types.AccountMeta) error {
			amInfo := make(map[string]*types.AccountMeta)
			amInfo["db"] = am
			amInfo["cache"] = amCache
			r[address] = amInfo
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (d *DebugApi) GetRepresentationsCache(address *types.Address) (map[types.Address]map[string]*types.Benefit, error) {
	r := make(map[types.Address]map[string]*types.Benefit)
	if address == nil {
		err := d.ledger.GetRepresentationsCache(types.ZeroAddress, func(address types.Address, be *types.Benefit, beCache *types.Benefit) error {
			beInfo := make(map[string]*types.Benefit)
			beInfo["db"] = be
			beInfo["cache"] = beCache
			r[address] = beInfo
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		err := d.ledger.GetRepresentationsCache(*address, func(address types.Address, be *types.Benefit, beCache *types.Benefit) error {
			beInfo := make(map[string]*types.Benefit)
			beInfo["db"] = be
			beInfo["cache"] = beCache
			r[address] = beInfo
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}
