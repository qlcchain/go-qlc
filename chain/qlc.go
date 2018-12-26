/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/wallet"
)

type QlcContext struct {
	Config      *config.Config
	Wallet      *wallet.WalletService
	Ledger      *ledger.LedgerService
	NetService  *p2p.QlcService
	DPosService *consensus.DposService
}

func New(cfg *config.Config) (*QlcContext, error) {
	return &QlcContext{Config: cfg}, nil
}
