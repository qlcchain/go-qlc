/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/log"
)

type Rollback struct {
	common.ServiceLifecycle
	Ledger *ledger.Ledger
	closed chan bool
	logger *zap.SugaredLogger
}

func NewRollbackService(cfgFile string) *Rollback {
	return &Rollback{
		Ledger: ledger.NewLedger(cfgFile),
		closed: make(chan bool, 1),
		logger: log.NewLogger("ledger_service"),
	}
}

func (rs *Rollback) Init() error {
	if !rs.PreInit() {
		return errors.New("pre init fail")
	}
	defer rs.PostInit()

	return nil
}

func (rs *Rollback) Start() error {
	if !rs.PreStart() {
		return errors.New("pre start fail")
	}
	defer rs.PostStart()
	rs.rollBack()
	return nil
}

func (rs *Rollback) Stop() error {
	if !rs.PreStop() {
		return errors.New("pre stop fail")
	}
	defer rs.PostStop()
	rs.closed <- true
	return nil
}

func (rs *Rollback) Status() int32 {
	return rs.State()
}

func (rs *Rollback) rollBack() {
	go func() {
		ledger := rs.Ledger
		verify := process.NewLedgerVerifier(ledger)
		for {
			select {
			case blk := <-ledger.RollbackChan:
				if err := verify.RollbackBlock(blk); err != nil {
					fmt.Println(err)
				}
			case <-rs.closed:
				return
			}
		}
	}()
}
