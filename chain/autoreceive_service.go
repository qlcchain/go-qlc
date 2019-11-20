/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/qlcchain/go-qlc/common/event"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

type AutoReceiveService struct {
	common.ServiceLifecycle
	subscriber *event.ActorSubscriber
	cfgFile    string
	blockCache chan *types.StateBlock
	quit       chan interface{}
	state      uint32
	logger     *zap.SugaredLogger
}

func NewAutoReceiveService(cfgFile string) *AutoReceiveService {
	return &AutoReceiveService{cfgFile: cfgFile, blockCache: make(chan *types.StateBlock, 100),
		quit: make(chan interface{}), state: 0, logger: log.NewLogger("auto_receive_service")}
}

func (as *AutoReceiveService) Init() error {
	if !as.PreInit() {
		return errors.New("pre init fail")
	}
	defer as.PostInit()
	cc := context.NewChainContext(as.cfgFile)
	bus := cc.EventBus()
	as.subscriber = event.NewActorSubscriber(event.Spawn(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *types.StateBlock:
			if msg != nil {
				as.blockCache <- msg
			}
		}
	}), bus)

	// TODO: Subscribe by address
	return as.subscriber.Subscribe(topic.EventConfirmedBlock)
}

func (as *AutoReceiveService) Start() error {
	if !as.PreStart() {
		return errors.New("pre start fail")
	}
	defer as.PostStart()

	cc := context.NewChainContext(as.cfgFile)

	go func() {
		for {
			ledgerService, _ := cc.Service(context.LedgerService)
			//ledger service started
			if ledgerService != nil && ledgerService.Status() == int32(common.Started) {
				break
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
		ledgerService, _ := cc.Service(context.LedgerService)

		l := ledgerService.(*LedgerService).Ledger
		accounts := cc.Accounts()
		for _, account := range accounts {
			a := account
			err := l.SearchPending(a.Address(), func(key *types.PendingKey, value *types.PendingInfo) error {
				as.logger.Debugf("%s receive %s[%s] from %s (%s)\n", key.Address, value.Type.String(), value.Source.String(), value.Amount.String(), key.Hash.String())
				if send, err := l.GetStateBlock(key.Hash); err != nil {
					as.logger.Error(err)
				} else {
					err = ReceiveBlock(send, a, cc)
					if err != nil {
						as.logger.Debugf("err[%s] when generate receive block.\n", err)
					}
				}
				return nil
			})

			if err != nil {
				as.logger.Error(err)
			}
		}
		atomic.StoreUint32(&as.state, 1)
	}()

	// auto receive
	go func() {
		for {
			select {
			case <-as.quit:
				atomic.StoreUint32(&as.state, 0)
				if err := as.subscriber.Unsubscribe(topic.EventConfirmedBlock); err != nil {
					as.logger.Error(err)
				}
				return
			case blk := <-as.blockCache:
				// waiting ledger service start and process pending
				if atomic.LoadUint32(&as.state) != 0 {
					accounts := cc.Accounts()
					for _, account := range accounts {
						addr := account.Address()
						rxAddr := types.Address(blk.Link)
						if (blk.Type == types.Send || blk.Type == types.ContractSend) && addr == rxAddr {
							var balance types.Balance
							if blk.Token == common.ChainToken() {
								balance, _ = common.RawToBalance(blk.Balance, "QLC")
								as.logger.Debugf("receive block from [%s] to [%s] balance [%s]", blk.Address.String(), rxAddr.String(), balance)
							} else {
								as.logger.Debugf("receive block from [%s] to [%s] balance [%s]", blk.Address.String(), rxAddr.String(), blk.Balance.String())
							}
							err := ReceiveBlock(blk, account, cc)
							if err != nil {
								as.logger.Errorf("err[%s] when generate receive block.", err)
							}
							break
						}
					}
				}
			}
		}
	}()

	return nil
}

func (as *AutoReceiveService) Stop() error {
	if !as.PreStop() {
		return errors.New("pre stop fail")
	}
	defer as.PostStop()

	as.quit <- struct{}{}

	return as.subscriber.Unsubscribe(topic.EventConfirmedBlock)
}

func (as *AutoReceiveService) Status() int32 {
	return as.State()
}
