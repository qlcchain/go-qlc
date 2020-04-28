/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"github.com/qlcchain/go-qlc/chain/context"
)

//RegisterServices register services to chain context
func RegisterServices(cc *context.ChainContext) error {
	cfgFile := cc.ConfigFile()
	cfg, err := cc.Config()
	if err != nil {
		return err
	}

	logService := NewLogService(cfgFile)
	_ = cc.Register(context.LogService, logService)
	_ = logService.Init()
	ledgerService := NewLedgerService(cfgFile)
	_ = cc.Register(context.LedgerService, ledgerService)

	if !cc.HasService(context.WalletService) {
		walletService := NewWalletService(cfgFile)
		_ = cc.Register(context.WalletService, walletService)
	}

	if cfg.P2P.IsBootNode {
		httpService := NewHttpService(cfgFile)
		_ = cc.Register(context.BootNodeHttpService, httpService)
	}

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

	if cfg.PoV.PovEnabled {
		povService := NewPoVService(cfgFile)
		_ = cc.Register(context.PovService, povService)
		minerService := NewMinerService(cfgFile, povService.GetPoVEngine())
		_ = cc.Register(context.MinerService, minerService)
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

	resendBlockService := NewResendBlockService(cfgFile)
	_ = cc.Register(context.ResendBlockService, resendBlockService)

	if cfg.Privacy.Enable {
		privacyService := NewPrivacyService(cfgFile)
		_ = cc.Register(context.PrivacyService, privacyService)
	}

	if cfg.WhiteList.Enable {
		permService := NewPermissionService(cfgFile)
		_ = cc.Register(context.PermissionService, permService)
	}

	return nil
}
