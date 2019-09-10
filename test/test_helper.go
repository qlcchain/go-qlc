/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package test

import (
	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/rpc"
)

func NewChain(dir string, fn func(config *config.Config)) (*context.ChainContext, error) {
	cm := config.NewCfgManager(dir)
	ctx := context.NewChainContext(cm.ConfigFile)
	cfg, err := ctx.Config()
	if err != nil {
		return nil, err
	}
	fn(cfg)

	return ctx, nil
}

func Ledger(ctx *context.ChainContext) (*ledger.Ledger, error) {
	ls, err := ctx.Service(context.LedgerService)
	if err != nil {
		return nil, err
	}
	return ls.(*chain.LedgerService).Ledger, nil
}

func RPC(ctx *context.ChainContext) (*rpc.RPC, error) {
	rs, err := ctx.Service(context.RPCService)
	if err != nil {
		return nil, err
	}
	return rs.(*chain.RPCService).RPC(), nil
}
