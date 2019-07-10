/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	rpc "github.com/qlcchain/jsonrpc2"
)

func runNode() error {
	err := registerServices()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = chainContext.Init()
	if err != nil {
		return err
	}

	err = chainContext.Start()
	if err != nil {
		return err
	}
	accounts := chainContext.Accounts()
	cfg, _ := chainContext.Config()

	//search pending and generate receive block
	if len(accounts) > 0 && cfg.AutoGenerateReceive {
		rpcService, err := chainContext.Service(context.RPCService)
		if err != nil {
			fmt.Println(err)
			return err
		}

		ledgerService, err := chainContext.Service(context.LedgerService)
		if err != nil {
			fmt.Println(err)
			return err
		}

		l := ledgerService.(*chain.LedgerService).Ledger

		if rpcService.Status() != int32(common.Started) || ledgerService.Status() != int32(common.Started) {
			return fmt.Errorf("rpc or ledger service not started")
		}

		go func(l *ledger.Ledger, accounts []*types.Account) {
			client, err := rpcService.(*chain.RPCService).RPC().Attach()
			if err != nil {
				return
			}
			defer func() {
				if client != nil {
					client.Close()
				}
			}()

			for _, account := range accounts {
				err := l.SearchPending(account.Address(), func(key *types.PendingKey, value *types.PendingInfo) error {
					fmt.Printf("%s receive %s[%s] from %s (%s)\n", key.Address, value.Type.String(), value.Source.String(), value.Amount.String(), key.Hash.String())
					if send, err := l.GetStateBlock(key.Hash); err != nil {
						fmt.Println(err)
					} else {
						err = receive(send, account, l, client)
						if err != nil {
							fmt.Printf("err[%s] when generate receive block.\n", err)
						}
					}
					return nil
				})

				if err != nil {
					fmt.Println(err)
				}
			}
		}(l, accounts)
	}

	return nil
}

func registerServices() error {
	accounts := chainContext.Accounts()
	cfg, err := chainContext.Config()
	if err != nil {
		return err
	}

	logService := log.NewLogService(cfg)
	_ = logService.Init()
	ledgerService := chain.NewLedgerService(cfg)
	_ = chainContext.Register(context.LedgerService, ledgerService)

	if !chainContext.HasService(context.WalletService) {
		walletService := chain.NewWalletService(cfg)
		_ = chainContext.Register(context.WalletService, walletService)
	}

	if len(cfg.P2P.BootNodes) > 0 {
		netService, err := chain.NewP2PService(cfg)
		if err != nil {
			fmt.Println(err)
			return err
		}
		_ = chainContext.Register(context.P2PService, netService)
	}
	consensusService := chain.NewConsensusService(cfg, accounts)
	_ = chainContext.Register(context.ConsensusService, consensusService)
	if rpcService, err := chain.NewRPCService(cfg); err != nil {
		return err
	} else {
		_ = chainContext.Register(context.RPCService, rpcService)
	}
	if sqliteService, err := chain.NewSqliteService(cfg); err != nil {
		return err
	} else {
		_ = chainContext.Register(context.IndexService, sqliteService)
	}

	if cfg.PoV.PovEnabled {
		povService := chain.NewPoVService(cfg, accounts)
		_ = chainContext.Register(context.PovService, povService)
		if len(cfg.PoV.Coinbase) > 0 {
			minerService := chain.NewMinerService(cfg, povService.GetPoVEngine())
			_ = chainContext.Register(context.MinerService, minerService)
		}
	}

	if len(accounts) > 0 && cfg.AutoGenerateReceive {
		go generateReceiveBlockLoop(accounts)
		eb := event.GetEventBus(cfg.LedgerDir())
		_ = eb.Subscribe(string(common.EventConfirmedBlock), func(blk *types.StateBlock) {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()

			rpcService, err := chainContext.Service(context.RPCService)
			if err != nil {
				fmt.Println(err)
				return
			}

			ledgerService, err := chainContext.Service(context.LedgerService)
			if err != nil {
				fmt.Println(err)
				return
			}

			l := ledgerService.(*chain.LedgerService).Ledger

			if rpcService.Status() != int32(common.Started) || ledgerService.Status() != int32(common.Started) {
				fmt.Println("rpc or ledger service not started")
				return
			}

			go func(accounts []*types.Account) {
				client, err := rpcService.(*chain.RPCService).RPC().Attach()
				if err != nil {
					return
				}
				defer func() {
					if client != nil {
						client.Close()
					}
				}()
				for _, value := range accounts {
					addr := value.Address()
					if blk.Type == types.Send {
						address := types.Address(blk.Link)
						if addr.String() == address.String() {
							var balance types.Balance
							if blk.Token == common.ChainToken() {
								balance, _ = common.RawToBalance(blk.Balance, "QLC")
								fmt.Printf("receive block from [%s] to[%s] balance[%s]\n", blk.Address.String(), address.String(), balance)
							} else {
								fmt.Printf("receive block from [%s] to[%s] balance[%s]", blk.Address.String(), address.String(), blk.Balance.String())
							}
							err := receive(blk, value, l, client)
							if err != nil {
								fmt.Printf("err[%s] when generate receive block.\n", err)
							}
							break
						}
					}
				}
			}(accounts)
		})
	}

	return nil
}

func receive(sendBlock *types.StateBlock, account *types.Account, l *ledger.Ledger, client *rpc.Client) error {
	receiveBlock, err := l.GenerateReceiveBlock(sendBlock, account.PrivateKey())
	if err != nil {
		return err
	}
	fmt.Println(util.ToIndentString(&receiveBlock))

	var h types.Hash
	err = client.Call(&h, "ledger_process", &receiveBlock)
	if err != nil {
		fmt.Println(util.ToString(&receiveBlock))
		fmt.Println("process block error: ", err)
		return err
	}

	return nil
}
