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

	"github.com/qlcchain/go-qlc/chain"
	ss "github.com/qlcchain/go-qlc/chain/services"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	cmn "github.com/tendermint/tmlibs/common"
)

func runNode(seed types.Seed, cfg *config.Config) error {
	err := initNode(seed, cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	services, err := startNode()
	if err != nil {
		fmt.Println(err)
	}

	//l := ctx.Ledger.Ledger
	//genesis := common.QLCGenesisBlock
	////var key []byte
	////key = append(key, types.MintageAddress[:]...)
	////key = append(key, genesis.Token[:]...)
	//if err := l.SetStorage(types.MintageAddress[:], genesis.Token[:], genesis.Data); err != nil {
	//	fmt.Println(err)
	//}
	//verifier := process.NewLedgerVerifier(l)
	//if b, err := l.HasStateBlock(common.GenesisMintageHash); !b && err == nil {
	//	if err := l.AddStateBlock(&common.GenesisMintageBlock); err != nil {
	//		fmt.Println(err)
	//	}
	//}
	//
	//if b, err := l.HasStateBlock(common.QLCGenesisBlockHash); !b && err == nil {
	//	if err := verifier.BlockProcess(&common.QLCGenesisBlock); err != nil {
	//		fmt.Println(err)
	//	}
	//}

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

func initNode(seed types.Seed, cfg *config.Config) error {
	var err error
	var accounts []*types.Account
	var addr types.Address
	ctx, err = chain.New(cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !seed.IsZero() {
		for i := 0; i < 100; i++ {
			acc, _ := seed.Account(uint32(i))
			accounts = append(accounts, acc)
		}
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
	if err != nil {
		return err
	}
	ctx.RPC = ss.NewRPCService(cfg, ctx.DPosService)

	if !seed.IsZero() {

		for _, value := range accounts {
			addr = value.Address()
			_ = ctx.NetService.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmedBlock, func(v interface{}) {

				if b, ok := v.(*types.StateBlock); ok {
					if b.Type == types.Send {
						address := types.Address(b.Link)

						if addr.String() == address.String() {
							balance, _ := common.RawToBalance(b.Balance, "QLC")
							fmt.Printf("receive block from [%s] to[%s] amount[%d]\n", b.Address.String(), address.String(), balance)
							err = receive(b, value)
							if err != nil {
								fmt.Printf("err[%s] when generate receive block.\n", err)
							}
						}
					}
				}
			})
		}
	}

	services = []common.Service{ctx.Ledger, ctx.NetService, ctx.Wallet, ctx.DPosService, ctx.RPC}

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

func receive(sendBlock *types.StateBlock, account *types.Account) error {
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
