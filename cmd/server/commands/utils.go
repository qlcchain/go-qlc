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
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/log"
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
		go func() {
			ledgerService, err := chainContext.Service(context.LedgerService)
			if err != nil {
				fmt.Println(err)
				return
			}

			l := ledgerService.(*chain.LedgerService).Ledger
			accounts := chainContext.Accounts()
			for _, account := range accounts {
				err := l.SearchPending(account.Address(), func(key *types.PendingKey, value *types.PendingInfo) error {
					fmt.Printf("%s receive %s[%s] from %s (%s)\n", key.Address, value.Type.String(), value.Source.String(), value.Amount.String(), key.Hash.String())
					if send, err := l.GetStateBlock(key.Hash); err != nil {
						fmt.Println(err)
					} else {
						err = receive(send, account)
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
		}()
	}

	return nil
}

func registerServices() error {
	accounts := chainContext.Accounts()
	cfg, err := chainContext.Config()
	if err != nil {
		return err
	}

	logService := log.NewLogService(cfgPathP)
	_ = logService.Init()
	ledgerService := chain.NewLedgerService(cfgPathP)
	_ = chainContext.Register(context.LedgerService, ledgerService)

	if !chainContext.HasService(context.WalletService) {
		walletService := chain.NewWalletService(cfgPathP)
		_ = chainContext.Register(context.WalletService, walletService)
	}

	if len(cfg.P2P.BootNodes) > 0 {
		netService, err := chain.NewP2PService(cfgPathP)
		if err != nil {
			fmt.Println(err)
			return err
		}
		_ = chainContext.Register(context.P2PService, netService)
	}
	consensusService := chain.NewConsensusService(cfgPathP)
	_ = chainContext.Register(context.ConsensusService, consensusService)
	if rpcService, err := chain.NewRPCService(cfgPathP); err != nil {
		return err
	} else {
		_ = chainContext.Register(context.RPCService, rpcService)
	}
	if sqliteService, err := chain.NewSqliteService(cfgPathP); err != nil {
		return err
	} else {
		_ = chainContext.Register(context.IndexService, sqliteService)
	}

	if cfg.PoV.PovEnabled {
		povService := chain.NewPoVService(cfgPathP)
		_ = chainContext.Register(context.PovService, povService)
		if len(cfg.PoV.Coinbase) > 0 {
			minerService := chain.NewMinerService(cfgPathP, povService.GetPoVEngine())
			_ = chainContext.Register(context.MinerService, minerService)
		}
	}

	if len(accounts) > 0 && cfg.AutoGenerateReceive {
		go func() {
			for {
				select {
				case <-exitChan:
					fmt.Println("exit generateReceiveBlock Loop...")
					return
				case blk := <-blocksChan:
					accounts := chainContext.Accounts()
					for _, account := range accounts {
						addr := account.Address()
						if blk.Type == types.Send || blk.Type == types.ContractSend {
							address := types.Address(blk.Link)
							if addr.String() == address.String() {
								var balance types.Balance
								if blk.Token == common.ChainToken() {
									balance, _ = common.RawToBalance(blk.Balance, "QLC")
									fmt.Printf("receive block from [%s] to[%s] balance[%s]\n", blk.Address.String(), address.String(), balance)
								} else {
									fmt.Printf("receive block from [%s] to[%s] balance[%s]", blk.Address.String(), address.String(), blk.Balance.String())
								}
								err := receive(blk, account)
								if err != nil {
									fmt.Printf("err[%s] when generate receive block.\n", err)
								}
								break
							}
						}
					}
				}
			}
		}()

		_ = chainContext.EventBus().Subscribe(common.EventConfirmedBlock, func(blk *types.StateBlock) {
			if blk != nil {
				blocksChan <- blk
			}
		})
	}

	return nil
}

func receive(sendBlock *types.StateBlock, account *types.Account) (err error) {
	rpcService, err := chainContext.Service(context.RPCService)
	if err != nil {
		return
	}

	ledgerService, err := chainContext.Service(context.LedgerService)
	if err != nil {
		return
	}

	l := ledgerService.(*chain.LedgerService).Ledger

	if rpcService.Status() != int32(common.Started) || ledgerService.Status() != int32(common.Started) {
		fmt.Println("rpc or ledger service not started")
		return
	}

	client, err := rpcService.(*chain.RPCService).RPC().Attach()
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
	fmt.Println(util.ToIndentString(&receiveBlock))

	var h types.Hash
	err = client.Call(&h, "ledger_process", &receiveBlock)
	if err != nil {
		fmt.Println(util.ToString(&receiveBlock))
		fmt.Println("process block error: ", err)
		return
	}

	return nil
}
