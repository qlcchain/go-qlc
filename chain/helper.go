/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"fmt"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

//RegisterServices register services to chain context
func RegisterServices(cc *context.ChainContext) error {
	cfgFile := cc.ConfigFile()
	cfg, err := cc.Config()
	if err != nil {
		return err
	}

	logService := log.NewLogService(cfgFile)
	_ = logService.Init()
	ledgerService := NewLedgerService(cfgFile)
	_ = cc.Register(context.LedgerService, ledgerService)

	if !cc.HasService(context.WalletService) {
		walletService := NewWalletService(cfgFile)
		_ = cc.Register(context.WalletService, walletService)
	}

	rollbackService := NewRollbackService(cfgFile)
	_ = cc.Register(context.RollbackService, rollbackService)

	if len(cfg.P2P.BootNodes) > 0 {
		netService, err := NewP2PService(cfgFile)
		if err != nil {
			return err
		}
		_ = cc.Register(context.P2PService, netService)
	}
	consensusService := NewConsensusService(cfgFile)
	_ = cc.Register(context.ConsensusService, consensusService)
	if rpcService, err := NewRPCService(cfgFile); err != nil {
		return err
	} else {
		_ = cc.Register(context.RPCService, rpcService)
	}
	if sqliteService, err := NewSqliteService(cfgFile); err != nil {
		return err
	} else {
		_ = cc.Register(context.IndexService, sqliteService)
	}

	if cfg.PoV.PovEnabled {
		povService := NewPoVService(cfgFile)
		_ = cc.Register(context.PovService, povService)
		if len(cfg.PoV.Coinbase) > 0 {
			minerService := NewMinerService(cfgFile, povService.GetPoVEngine())
			_ = cc.Register(context.MinerService, minerService)
		}
	}

	accounts := cc.Accounts()
	if len(accounts) > 0 && cfg.AutoGenerateReceive {
		autoReceiveService := NewAutoReceiveService(cfgFile)
		_ = cc.Register(context.AutoReceiveService, autoReceiveService)
	}

	if cfg.Metrics.Enable {
		metricsService := NewMetricsService(cfgFile)
		_ = cc.Register(context.MetricsService, metricsService)
	}

	chainManageService := NewChainManageService(cfgFile)
	_ = cc.Register(context.ChainManageService, chainManageService)

	return nil
}

// ReceiveBlock generate receive block
func ReceiveBlock(sendBlock *types.StateBlock, account *types.Account, cc *context.ChainContext) (err error) {
	rpcService, err := cc.Service(context.RPCService)
	if err != nil {
		return
	}

	ledgerService, err := cc.Service(context.LedgerService)
	if err != nil {
		return
	}

	l := ledgerService.(*LedgerService).Ledger

	if rpcService.Status() != int32(common.Started) || ledgerService.Status() != int32(common.Started) {
		return fmt.Errorf("rpc or ledger service not started")
	}

	client, err := rpcService.(*RPCService).RPC().Attach()
	if err != nil {
		return
	}
	defer func() {
		if client != nil {
			client.Close()
		}
	}()
	var receiveBlock *types.StateBlock
	if sendBlock.Type == types.Send {
		receiveBlock, err = l.GenerateReceiveBlock(sendBlock, account.PrivateKey())
		if err != nil {
			return
		}
	} else if sendBlock.Type == types.ContractSend && sendBlock.Link == types.Hash(types.RewardsAddress) {
		sendHash := sendBlock.GetHash()
		err = client.Call(&receiveBlock, "rewards_getReceiveRewardBlock", &sendHash)
		if err != nil {
			return
		}
	}
	if receiveBlock != nil {
		var h types.Hash
		err = client.Call(&h, "ledger_process", &receiveBlock)
		if err != nil {
			return
		}
	}
	return nil
}
