package commands

import (
	"fmt"
	"reflect"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
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
		_ = ctx.NetService.MessageEvent().GetEvent("consensus").Subscribe(p2p.EventConfirmedBlock, func(v interface{}) {

			if b, ok := v.(*types.StateBlock); ok {
				if b.Address.ToHash() != b.Link {
					s := ctx.Wallet.Wallet.NewSession(account)
					if isValid, err := s.VerifyPassword(password); isValid && err == nil {
						if a, err := s.GetRawKey(types.Address(b.Link)); err == nil {
							addr := a.Address()
							if addr.ToHash() == b.Link {
								balance, _ := mock.RawToBalance(b.Balance, "QLC")
								fmt.Printf("receive block from [%s] to[%s] amount[%d]\n", b.Address.String(), addr.String(), balance)
								err = receive(b, s, addr)
								if err != nil {
									fmt.Printf("err[%s] when generate receive block.\n", err)
								}
							}
						}
					}
				}
			}
		})
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

func receive(sendBlock types.Block, session *wallet.Session, address types.Address) error {
	l := ctx.Ledger.Ledger
	n := ctx.NetService

	receiveBlock, err := session.GenerateReceiveBlock(sendBlock)
	if err != nil {
		return err
	}
	fmt.Println(jsoniter.MarshalToString(&receiveBlock))
	if r, err := l.Process(receiveBlock); err != nil || r == ledger.Other {
		fmt.Println(jsoniter.MarshalToString(&receiveBlock))
		fmt.Println("process block error", err)
		return err
	} else {
		fmt.Println("receive block, ", receiveBlock.GetHash())

		meta, err := l.GetAccountMeta(address)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println(jsoniter.MarshalToString(&meta))
		pushBlock := protos.PublishBlock{
			Blk: receiveBlock,
		}
		bytes, err := protos.PublishBlockToProto(&pushBlock)
		if err != nil {
			fmt.Println(err)
			return err
		} else {
			n.Broadcast(p2p.PublishReq, bytes)
		}
	}
	return nil
}

func receiveblock(sendBlock types.Block, address types.Address, prk ed25519.PrivateKey) error {
	l := ctx.Ledger.Ledger
	n := ctx.NetService

	receiveBlock, err := l.GenerateReceiveBlock(sendBlock, prk)
	if err != nil {
		return err
	}
	fmt.Println(jsoniter.MarshalToString(&receiveBlock))
	if r, err := l.Process(receiveBlock); err != nil || r == ledger.Other {
		fmt.Println(jsoniter.MarshalToString(&receiveBlock))
		fmt.Println("process block error", err)
		return err
	} else {
		fmt.Println("receive block, ", receiveBlock.GetHash())

		meta, err := l.GetAccountMeta(address)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println(jsoniter.MarshalToString(&meta))
		pushBlock := protos.PublishBlock{
			Blk: receiveBlock,
		}
		bytes, err := protos.PublishBlockToProto(&pushBlock)
		if err != nil {
			fmt.Println(err)
			return err
		} else {
			n.Broadcast(p2p.PublishReq, bytes)
		}
	}
	return nil
}
