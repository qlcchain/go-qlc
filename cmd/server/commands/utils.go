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

	ss "github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	cmn "github.com/tendermint/tmlibs/common"
)

func runNode(accounts []*types.Account, cfg *config.Config) error {
	err := initNode(accounts, cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	services, err := startNode(accounts, cfg)
	if err != nil {
		fmt.Println(err)
	}
	if err := initDb(); err != nil {
		fmt.Println(err)
		return err
	}
	cmn.TrapSignal(func() {
		stopNode(services)
	})
	return nil
}

func stopNode(services []common.Service) {
	for _, service := range services {
		fmt.Printf("%s stopping ...\n", reflect.TypeOf(service))
		err := service.Stop()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func initNode(accounts []*types.Account, cfg *config.Config) error {
	logService := log.NewLogService(cfg)
	_ = logService.Init()
	ledgerService = ss.NewLedgerService(cfg)
	walletService = ss.NewWalletService(cfg)
	netService, err := ss.NewP2PService(cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//ctx.DPosService = ss.NewDPosService(cfg, ctx.NetService, account, password)
	dPosService = ss.NewDPosService(cfg, accounts)
	if rPCService, err = ss.NewRPCService(cfg); err != nil {
		return err
	}
	if sqliteService, err = ss.NewSqliteService(cfg); err != nil {
		return err
	}

	povService = ss.NewPoVService(cfg, accounts)
	minerService = ss.NewMinerService(cfg, povService.GetPoVEngine())

	if len(accounts) > 0 && cfg.AutoGenerateReceive {
		eb := event.GetEventBus(cfg.LedgerDir())
		_ = eb.Subscribe(string(common.EventConfirmedBlock), func(blk *types.StateBlock) {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()

			go func(accounts []*types.Account) {
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
							err = receive(blk, value)
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

	services = []common.Service{netService, walletService, dPosService, rPCService, povService, minerService, sqliteService, ledgerService}

	return nil
}

func startNode(accounts []*types.Account, cfg *config.Config) ([]common.Service, error) {
	for _, service := range services {
		err := service.Init()
		if err != nil {
			return nil, err
		}
		err = service.Start()
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s start successfully.\n", reflect.TypeOf(service))
	}
	fmt.Println("qlc node start successfully")
	//search pending and generate receive block
	if len(accounts) > 0 && cfg.AutoGenerateReceive {
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

		}(ledgerService.Ledger, accounts)
	}

	return services, nil
}

func receive(sendBlock *types.StateBlock, account *types.Account) error {
	if rPCService.State() != int32(common.Started) || ledgerService.State() != int32(common.Started) {
		return errors.New("rpc or ledger service not started")
	}
	l := ledgerService.Ledger

	receiveBlock, err := l.GenerateReceiveBlock(sendBlock, account.PrivateKey())
	if err != nil {
		return err
	}
	fmt.Println(util.ToIndentString(&receiveBlock))

	client, err := rPCService.RPC().Attach()
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

func initDb() error {
	relation := sqliteService.Relation
	c, err := relation.BlocksCount()
	if err != nil {
		return err
	}
	if c == 0 {
		ledgerService.Ledger.GetStateBlocks(func(block *types.StateBlock) error {
			if err := relation.AddBlock(block); err != nil {
				return err
			}
			return nil
		})
	}
	return nil
}
