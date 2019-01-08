package commands

import (
	"fmt"
	"reflect"

	"github.com/qlcchain/go-qlc/config"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"github.com/qlcchain/go-qlc/p2p/protos"
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
			logger.Error(err)
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
	ctx.Ledger = ledger.NewLedgerService(cfg)
	ctx.Wallet = wallet.NewWalletService(cfg)
	ctx.NetService, err = p2p.NewQlcService(cfg)
	logService := log.NewLogService(cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}
	ctx.DPosService, err = consensus.NewDposService(cfg, ctx.NetService, account, password)
	if err != nil {
		return err
	}

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
								logger.Debugf("receive block from [%s] to[%s] amount[%d]", b.Address.String(), addr.String(), balance)
								err = receive(b, s, addr)
								if err != nil {
									logger.Debugf("err[%s] when generate receive block.", err)
								}
							}
						}
					}
				}
			}
		})
	}

	services = []common.Service{logService, ctx.NetService, ctx.DPosService, ctx.Ledger, ctx.Wallet}

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
		logger.Debugf("%s start successful.", reflect.TypeOf(service))
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
	logger.Debug(jsoniter.MarshalToString(&receiveBlock))
	if r, err := l.Process(receiveBlock); err != nil || r == ledger.Other {
		logger.Debug(jsoniter.MarshalToString(&receiveBlock))
		logger.Error("process block error", err)
		return err
	} else {
		logger.Info("receive block, ", receiveBlock.GetHash())

		meta, err := l.GetAccountMeta(address)
		if err != nil {
			logger.Error(err)
			return err
		}
		logger.Debug(jsoniter.MarshalToString(&meta))
		pushBlock := protos.PublishBlock{
			Blk: receiveBlock,
		}
		bytes, err := protos.PublishBlockToProto(&pushBlock)
		if err != nil {
			logger.Error(err)
			return err
		} else {
			n.Broadcast(p2p.PublishReq, bytes)
		}
	}
	return nil
}
