/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package services

import (
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc"
	"go.uber.org/zap"
)

type RPCService struct {
	rpc    *rpc.RPC
	logger *zap.SugaredLogger
}

func NewRPCService(cfg *config.Config, dpos *DPosService) *RPCService {
	return &RPCService{rpc: rpc.NewRPC(cfg, dpos.dpos), logger: log.NewLogger("rpc_service")}
}

func (rs *RPCService) Init() error {
	rs.logger.Debug("rpc service init")
	return nil
}

func (rs *RPCService) Start() error {
	err := rs.rpc.StartRPC()
	if err != nil {
		rs.logger.Error(err)
		return err
	}
	return nil
}

func (rs *RPCService) Stop() error {
	rs.logger.Info("rpc stopping...")
	rs.rpc.StopRPC()
	rs.logger.Info("rpc stopped")
	return nil
}

func (rs *RPCService) Status() int32 {
	panic("implement me")
}

func (rs *RPCService) RPC() *rpc.RPC {
	return rs.rpc
}
