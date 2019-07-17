package api

import (
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
	r := make(map[string]types.Hash)
	child, err := d.ledger.GetChild(hash)
	if err == nil {
		r["child"] = child
	}
	link, _ := d.ledger.GetLinkBlock(hash)
	if !link.IsZero() {
		r["receiver"] = link
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
