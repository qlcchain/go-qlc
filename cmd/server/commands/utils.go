/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/qlcchain/go-qlc/chain"
	ss "github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	cmn "github.com/tendermint/tmlibs/common"
)

func runNode(accounts []*types.Account, cfg *config.Config) error {
	err := initNode(accounts, cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	services, err := startNode(accounts)
	if err != nil {
		fmt.Println(err)
	}

	cmn.TrapSignal(func() {
		stopNode(services)
	})
	return nil
}

func stopNode(services []common.Service) {
	for _, service := range services {
		err := service.Stop()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func initNode(accounts []*types.Account, cfg *config.Config) error {
	var err error
	ctx, err = chain.New(cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	logService := log.NewLogService(cfg)
	_ = logService.Init()
	ctx.Ledger = ss.NewLedgerService(cfg)
	ctx.Wallet = ss.NewWalletService(cfg)
	ctx.NetService, err = p2p.NewQlcService(cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//ctx.DPosService = ss.NewDPosService(cfg, ctx.NetService, account, password)
	ctx.DPosService = ss.NewDPosService(cfg, ctx.NetService, accounts)
	ctx.RPC = ss.NewRPCService(cfg, ctx.DPosService)

	if len(accounts) > 0 && cfg.AutoGenerateReceive {
		_ = ctx.NetService.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmedBlock, func(v interface{}) {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()

			go func(accounts []*types.Account) {
				for _, value := range accounts {
					addr := value.Address()
					if b, ok := v.(*types.StateBlock); ok {
						if b.Type == types.Send {
							address := types.Address(b.Link)
							if addr.String() == address.String() {
								var balance types.Balance
								if b.Token == common.ChainToken() {
									balance, _ = common.RawToBalance(b.Balance, "QLC")
									fmt.Printf("receive block from [%s] to[%s] balance[%s]\n", b.Address.String(), address.String(), balance)
								} else {
									fmt.Printf("receive block from [%s] to[%s] balance[%s]", b.Address.String(), address.String(), b.Balance.String())
								}
								err = receive(b, value)
								if err != nil {
									fmt.Printf("err[%s] when generate receive block.\n", err)
								}
								break
							}
						}
					}
				}
			}(accounts)
		})
	}

	services = []common.Service{ctx.Ledger, ctx.NetService, ctx.Wallet, ctx.DPosService, ctx.RPC}

	return nil
}

func startNode(accounts []*types.Account) ([]common.Service, error) {
	for _, service := range services {
		err := service.Init()
		if err != nil {
			return nil, err
		}
		err = service.Start()
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s start successful.\n", reflect.TypeOf(service))
	}

	//search pending and generate receive block
	go func(l *ledger.Ledger, accounts []*types.Account) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()

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

	}(ctx.Ledger.Ledger, accounts)

	return services, nil
}

func receive(sendBlock *types.StateBlock, account *types.Account) error {
	if ctx.RPC.State() != int32(common.Started) || ctx.Ledger.State() != int32(common.Started) {
		return errors.New("rpc or ledger service not started")
	}
	l := ctx.Ledger.Ledger

	receiveBlock, err := l.GenerateReceiveBlock(sendBlock, account.PrivateKey())
	if err != nil {
		return err
	}
	fmt.Println(util.ToIndentString(&receiveBlock))

	client, err := ctx.RPC.RPC().Attach()
	if err != nil {
		fmt.Println("create rpc client error:", err)
		return err
	}
	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	var h types.Hash
	err = client.Call(&h, "ledger_process", &receiveBlock)
	if err != nil {
		fmt.Println(util.ToString(&receiveBlock))
		fmt.Println("process block error: ", err)
		return err
	}

	return nil
}
