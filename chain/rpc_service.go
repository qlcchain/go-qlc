/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc"
)

type RPCService struct {
	common.ServiceLifecycle
	rpc    *rpc.RPC
	logger *zap.SugaredLogger
}

func NewRPCService(cfgFile string) (*RPCService, error) {
	rpc, err := rpc.NewRPC(cfgFile)
	if err != nil {
		return nil, err
	}
	return &RPCService{rpc: rpc, logger: log.NewLogger("rpc_service")}, nil
}

func (rs *RPCService) Init() error {
	if !rs.PreInit() {
		return errors.New("pre init fail")
	}
	defer rs.PostInit()

	return nil
}

func (rs *RPCService) Start() error {
	if !rs.PreStart() {
		return errors.New("pre start fail")
	}
	err := rs.rpc.StartRPC()
	if err != nil {
		rs.logger.Error(err)
		return err
	}
	rs.PostStart()
	return nil
}

func (rs *RPCService) Stop() error {
	if !rs.PreStop() {
		return errors.New("pre stop fail")
	}
	defer rs.PostStop()

	rs.rpc.StopRPC()
	rs.logger.Info("rpc stopped")
	return nil
}

func (rs *RPCService) Status() int32 {
	return rs.State()
}

func (rs *RPCService) RPC() *rpc.RPC {
	return rs.rpc
}
