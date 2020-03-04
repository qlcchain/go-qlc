/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"

	"github.com/qlcchain/go-qlc/log"
	rpc "github.com/qlcchain/jsonrpc2"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/chain/context"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

type LogService struct {
	common.ServiceLifecycle
	cfg *config.Config
}

func NewLogService(cfgFile string) *LogService {
	cc := context.NewChainContext(cfgFile)
	cfg, _ := cc.Config()
	return &LogService{cfg: cfg}
}

func (ls *LogService) Init() error {
	if !ls.PreInit() {
		return errors.New("LogService pre init fail")
	}
	defer ls.PostInit()

	// enable rpc debug log
	l := zap.ErrorLevel
	if err := l.Set(ls.cfg.LogLevel); err == nil {
		if l.Enabled(zap.DebugLevel) {
			rpc.IsDebug = true
		}
	}

	return log.Setup(ls.cfg)
}

func (ls *LogService) Start() error {
	if !ls.PreStart() {
		return errors.New("LogService pre start fail")
	}
	defer ls.PostStart()

	return nil
}

func (ls *LogService) Stop() error {
	if !ls.PreStop() {
		return errors.New("LogService pre stop fail")
	}
	defer ls.PostStop()

	return log.Teardown()
}

func (ls *LogService) Status() int32 {
	return ls.State()
}
