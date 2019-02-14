/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/test/mock"
	"github.com/qlcchain/go-qlc/wallet"
	cmn "github.com/tendermint/tmlibs/common"
)

func runNode(account types.Address, password string, cfg *config.Config) error {
	err := initNode(account, password, cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	services, err := startNode()
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

func initNode(account types.Address, password string, cfg *config.Config) error {
	var err error
	ctx, err = chain.New(cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	logService := log.NewLogService(cfg)
	_ = logService.Init()
	ctx.Ledger = ledger.NewLedgerService(cfg)
	ctx.Wallet = wallet.NewWalletService(cfg)
	ctx.NetService, err = p2p.NewQlcService(cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	ctx.DPosService, err = consensus.NewDposService(cfg, ctx.NetService, account, password)
	if err != nil {
		return err
	}
	ctx.RPC = rpc.NewRPCService(cfg, ctx.DPosService)

	if !account.IsZero() {
		s := ctx.Wallet.Wallet.NewSession(account)

		var cache sync.Map

		if isValid, err := s.VerifyPassword(password); isValid && err == nil {
			_ = ctx.NetService.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmedBlock, func(v interface{}) {

				if b, ok := v.(*types.StateBlock); ok {
					address := types.Address(b.Link)
					// genesis block
					if b.Address.ToHash() == b.Link {
						return
					}

					if _, ok := cache.Load(address); !ok {
						if account, err := s.GetRawKey(address); err == nil {
							cache.Store(address, account)
						}
					}

					if account, ok := cache.Load(address); ok {
						balance, _ := mock.RawToBalance(b.Balance, "QLC")
						fmt.Printf("receive block from [%s] to[%s] amount[%d]\n", b.Address.String(), address.String(), balance)
						err = receive(b, account.(*types.Account))
						if err != nil {
							fmt.Printf("err[%s] when generate receive block.\n", err)
						}

					}
				}
			})
		} else {
			fmt.Println("invalid password")
		}
	}

	services = []common.Service{ctx.NetService, ctx.DPosService, ctx.Ledger, ctx.Wallet, ctx.RPC}

	return nil
}

func startNode() ([]common.Service, error) {
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

	return services, nil
}

func receive(sendBlock types.Block, account *types.Account) error {
	l := ctx.Ledger.Ledger

	receiveBlock, err := l.GenerateReceiveBlock(sendBlock, account.PrivateKey())
	if err != nil {
		return err
	}
	fmt.Println(util.ToString(&receiveBlock))

	client, err := ctx.RPC.RPC().Attach()
	defer client.Close()

	var h types.Hash
	err = client.Call(&h, "ledger_process", &receiveBlock)
	if err != nil {
		fmt.Println(util.ToString(&receiveBlock))
		fmt.Println("process block error", err)
	}

	return nil
}
