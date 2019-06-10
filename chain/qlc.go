/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/config"
)

type QlcContext struct {
	Config      *config.Config
	Wallet      *services.WalletService
	Ledger      *services.LedgerService
	NetService  *services.P2PService
	DPosService *services.DPosService
	RPC         *services.RPCService
	PoVService  *services.PoVService
	Miner       *services.MinerService
}

func New(cfg *config.Config) (*QlcContext, error) {
	return &QlcContext{Config: cfg}, nil
}
